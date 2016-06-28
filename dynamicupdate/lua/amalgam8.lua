local json     = require("cjson.safe")
local math     = require("math")
local balancer = require("ngx.balancer")
local lock     = require("resty.lock")
local string   = require("resty.string")

local _M = { _VERSION = '0.5.0' }
local mt = { __index = _M }

local function decodeJson(input)
  local upstreams = {}
  local services = {}

  for name, upstream in pairs(input.upstreams) do
    upstreams[name] = {
      servers = {},
    }

    for _, server in ipairs(upstream.servers) do
       table.insert(upstreams[name].servers, {
		       host   = server.host,
		       port   = server.port,
		})
    end
  end

  for name, version_selector in pairs(input.services) do
     services[name] = {
	default = version_selector.default,
	selectors = version_selector.selectors
     }
  end

  return upstreams, services
end

local function unlocked(l, cause)
  local _, err = l:unlock()
  if err then
    ngx.log(ngx.ERR, "error unlocking lock:" .. err)
  end

  return cause
end

function _M.new(_, upstreams_dict, services_dict, locks_dict)
  local upstreams = ngx.shared[upstreams_dict]
  if not upstreams then
    return nil, "upstreams dictionary " .. upstreams_dict .. " not found"
  end

  local services = ngx.shared[services_dict]
  if not upstreams then
    return nil, "services dictionary " .. services_dict .. " not found"
  end

  local locks = ngx.shared[locks_dict]
  if not locks then
    return nil, "locks dictionary " .. locks_dict .. " not found"
  end

  local self = {
    upstreams_dict   = upstreams_dict,
    services_dict   = services_dict,
    locks_dict  = locks_dict,
  }

  return setmetatable(self, mt)
end

function _M.updateUpstream(self, name, upstream)
  if table.getn(upstream.servers) == 0 then
    return
  end

  local l = lock:new(self.locks_dict, {
    exptime  = 10,
    timeout  = 5,
    step     = 0.01,
    ratio    = 2,
    max_step = 0.1,
  })

  local _, err = l:lock(name)
  if err then
    return err
  end

  local current = {
    name         = name,
    upstream     = upstream,
  }

  for _, server in ipairs(upstream.servers) do
    if not server.host then
       return unlocked(l, "null host for server")
    end

    if not server.port then
       return unlocked(l, "null port for server")
    end
  end

  local current_serialized, err = json.encode(current)
  if err then
    return unlocked(l, err)
  end

  local _, err = ngx.shared[self.upstreams_dict]:set(name, current_serialized)
  if err then
    return unlocked(l, err)
  end

  -- ngx.log(ngx.NOTICE, "updated " .. name .. " with " .. table.getn(upstream.servers) .. " upstreams")

  return unlocked(l)
end

function _M.updateService(self, name, version_selector)
  local l = lock:new(self.locks_dict, {
    exptime  = 10,
    timeout  = 5,
    step     = 0.01,
    ratio    = 2,
    max_step = 0.1,
  })

  local _, err = l:lock(name)
  if err then
    return err
  end

  local current = {
    name         = name,
    version_selector   = version_selector,
  }

  if not version_selector.default then
     return unlocked(l, "null entry for default version")
  end

  -- convert the selectors field from string to lua table
  local selectors_fn, err = loadstring( "return " .. version_selector.selectors)
  if err then
     ngx.log(ngx.ERR, "failed to compile selector", err)
     return err
  end
  current.version_selector.selectors = selectors_fn()

  local current_serialized, err = json.encode(current)
  if err then
    return unlocked(l, err)
  end

  local _, err = ngx.shared[self.services_dict]:set(name, current_serialized)
  if err then
    return unlocked(l, err)
  end

  return unlocked(l)
end

function _M.updateProxy(self, input)
  local upstreams, services = decodeJson(input)

  for name, upstream in pairs(upstreams) do
    err = self:updateUpstream(name, upstream)
    if err then
      err = "error updating upstream " .. name .. ": " .. err
    end
  end

  if err then
    return err
  end

  for name, version_selector in pairs(services) do
    err = self:updateService(name, version_selector)
    if err then
      err = "error updating service " .. name .. ": " .. err
    end
  end

  if err then
    return err
  end
end

function _M.getUpstream(self, name)
  local serialized = ngx.shared[self.upstreams_dict]:get(name)
  if serialized then
    return json.decode(serialized)
  end
end

function _M.getService(self, name)
  local serialized = ngx.shared[self.services_dict]:get(name)
  if serialized then
    return json.decode(serialized)
  end
end

function _M.init(upstreams_dict, services_dict, locks_dict)
  local a8proxy, err = _M:new(upstreams_dict, services_dict, locks_dict)
  if err then
    ngx.log(ngx.ERR, "error creating a8proxy instance: " .. err)
    return
  end
end

function _M.handle(upstreams_dict, services_dict, locks_dict)
  local a8proxy, err = _M:new(upstreams_dict, services_dict, locks_dict)
  if err then
    ngx.log(ngx.ERR, "error creating a8proxy instance: " .. err)
    ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
  end

  if ngx.req.get_method() == "PUT" or ngx.req.get_method() == "POST" then
    ngx.req.read_body()

    local input, err = json.decode(ngx.req.get_body_data())
    if err then
      ngx.log(ngx.ERR, "error decoding input json: " .. err)
      ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
    end

    local err = a8proxy:updateProxy(input)
    if err then
      ngx.log(ngx.ERR, "error updating proxy: " .. err)
      ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
    end
  else
    return ngx.exit(ngx.HTTP_BAD_REQUEST)
  end
end

function _M.balance(upstreams_dict, services_dict, locks_dict)
  local name = ngx.var.a8proxy_upstream
  if not name then
    ngx.log(ngx.ERR, "$a8proxy_upstream is not specified")
    return
  end

  local a8proxy, err = _M:new(upstreams_dict, services_dict, locks_dict)
  if err then
     ngx.log(ngx.ERR, "error creating a8proxy instance: " .. err)
     return
  end

  local upstream, err = a8proxy:getUpstream(name)
  if err then
     ngx.log(ngx.ERR, "error getting upstream " .. name .. ": " .. err)
     return
  end

  if not upstream or table.getn(upstream.upstream.servers) == 0 then
     ngx.log(ngx.ERR, "upstream " .. name .. " is not known or has no known instances")
     return
  end

  local index    = math.random(1, table.getn(upstream.upstream.servers)) % table.getn(upstream.upstream.servers) + 1
  local upstream = upstream.upstream.servers[index]

  local _, err = balancer.set_current_peer(upstream.host, upstream.port)
  if err then
    ngx.log(ngx.ERR, "failed to set current peer for upstream " .. name .. ": " .. err)
    return
  end
end

function _M.versioninfo(upstreams_dict, services_dict, locks_dict)
  local name = ngx.var.service_name
  if not name then
    ngx.log(ngx.ERR, "$service_name is not specified")
    return
  end

  local a8proxy, err = _M:new(upstreams_dict, services_dict, locks_dict)
  if err then
     ngx.log(ngx.ERR, "error creating a8proxy instance: " .. err)
     return
  end

  local service, err = a8proxy:getService(name)
  if err then
     ngx.log(ngx.ERR, "error getting service " .. name .. ": " .. err)
     return
  end

  if not service or not service.version_selector then
     ngx.log(ngx.ERR, "service " .. name .. " or version_selector is not known")
     return
  end
  return service.version_selector.default, service.version_selector.selectors
end

return _M
