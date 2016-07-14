local json     = require("cjson.safe")
local math     = require("math")
local balancer = require("ngx.balancer")
local string   = require("resty.string")

local Amalgam8 = { _VERSION = '0.2.0',
	     upstreams_dict = {},
	     faults_dict = {},
	     services_dict = {}
}

local function decodeJson(input)
   local upstreams = {}
   local services = {}
   local faults = {}

   if input.upstreams then
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
   end

   if input.services then
      for name, version_selector in pairs(input.services) do
         services[name] = {
            default = version_selector.default,
            selectors = version_selector.selectors
         }
      end
   end

   if input.faults then
      if not ngx.var.server_name then
         ngx.log(ngx.WARNING, "Server_name variable is empty! Cannot generate fault injection rules")
      else
         for i, fault in ipairs(input.faults) do
            if fault.source == ngx.var.source_service then
               table.insert(faults, {
                               source = fault.source,
                               destination = fault.destination,
                               header =  fault.header,
                               pattern = fault.pattern,
                               delay = fault.delay,
                               delay_probability = fault.delay_probability,
                               abort_probability = fault.abort_probability,
                               return_code = fault.return_code
               })
            end
         end
      end
   end

   return upstreams, services, faults
end

function Amalgam8:new(upstreams_dict, services_dict, faults_dict)
   o = {}
   setmetatable(o, self)
   self.__index = self

   local upstreams = ngx.shared[upstreams_dict]
   if not upstreams then
      return "upstreams dictionary " .. upstreams_dict .. " not found"
   end

   local services = ngx.shared[services_dict]
   if not services then
      return "services dictionary " .. services_dict .. " not found"
   end

   local faults = ngx.shared[faults_dict]
   if not faults then
      return "faults dictionary " .. faults_dict .. " not found"
   end

   self.upstreams_dict = upstreams
   self.services_dict = services
   self.faults_dict = faults
   return o
end

function Amalgam8:updateUpstream(name, upstream)
   if table.getn(upstream.servers) == 0 then
      return nil
   end

   local current = {
      name         = name,
      upstream     = upstream,
   }

   for _, server in ipairs(upstream.servers) do
      if not server.host then
         return "null host for server"
      end

      if not server.port then
         return "null port for server"
      end
   end

   local current_serialized, err = json.encode(current)
   if err then
      return err
   end

   local _, err = self.upstreams_dict:set(name, current_serialized)
   if err then
      return err
   end

   return nil
end

function Amalgam8:updateService(name, version_selector)

   local current = {
      name         = name,
      version_selector   = version_selector,
   }

   if not version_selector.default then
      return "null entry for default version"
   end

   if version_selector.selectors then
      -- convert the selectors field from string to lua table
      local selectors_fn, err = loadstring( "return " .. version_selector.selectors)
      if err then
         ngx.log(ngx.ERR, "failed to compile selector", err)
         return err
      end
      current.version_selector.selectors = selectors_fn()
   end

   local current_serialized, err = json.encode(current)
   if err then
      return err
   end

   local _, err = self.services_dict:set(name, current_serialized)
   if err then
      return err
   end

   return nil
end

function Amalgam8:updateFault(i, fault)

   local name = tostring(i)

   local current = {
      name  = name,
      fault = fault,
   }

   -- no error checks here, assuming that the controller has done them already.
   -- FIXME: still need to define rule prioritization in the presence of wildcards.

   local current_serialized, err = json.encode(current)
   if err or not current_serialized then
      ngx.log(ngx.ERR, "JSON encoding failed for fault fault between " .. fault.source .. " and " .. fault.destination .. " err:" .. err)
      return err
   end

   -- local _, err = table.insert(self.faults_dict, current_serialized)
   local _, err = self.faults_dict:set(i, current_serialized)
   if err then
      ngx.log(ngx.ERR, "Failed to update faults_dict for fault between " .. fault.source .. " and " .. fault.destination .. " err:" .. err)
      return err
   end

   return nil
end

function Amalgam8:updateProxy(input)
   local upstreams, services, faults = decodeJson(input)

   for name, upstream in pairs(upstreams) do
      err = self:updateUpstream(name, upstream)
      if err then
         err = "error updating upstream " .. name .. ": " .. err
	 break
      end
   end

   if err then
      return err
   end

   for name, version_selector in pairs(services) do
      err = self:updateService(name, version_selector)
      if err then
         err = "error updating service " .. name .. ": " .. err
	 break
      end
   end

   if err then
      return err
   end

   for i, fault in ipairs(faults) do
      err = self:updateFault(i, fault)
      if err then
         err = "error updating fault " .. fault.destination .. ": " .. err
	 break
      end
   end

   if err then
      return err
   end

   self.faults_dict:set("len", table.getn(faults))
end

function Amalgam8:getUpstream(name)
   local serialized = self.upstreams_dict:get(name)
   if serialized then
      return json.decode(serialized)
   end
end

function Amalgam8:getService(name)
   local serialized = self.services_dict:get(name)
   if serialized then
      return json.decode(serialized)
   end
end

--- FIXME
function Amalgam8:resetState()
   self.upstreams_dict:flush_all()
   self.upstreams_dict:flush_expired()
   self.services_dict:flush_all()
   self.services_dict:flush_expired()
   self.faults_dict:flush_all()
   self.faults_dict:flush_expired()
   self.faults_dict:set("len", 0) --marker
end

function Amalgam8:updateConfig()
   if ngx.req.get_method() == "PUT" or ngx.req.get_method() == "POST" then
      ngx.req.read_body()

      local input, err = json.decode(ngx.req.get_body_data())
      if err then
         ngx.log(ngx.ERR, "error decoding input json: " .. err)
         ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
      end

      -- TODO: FIXME. There is no concept of locking so far.
      self:resetState()
      err = self:updateProxy(input)
      if err then
         ngx.log(ngx.ERR, "error updating proxy: " .. err)
         ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
      end
   else
      return ngx.exit(ngx.HTTP_BAD_REQUEST)
   end
end

function Amalgam8:balance()
   local name = ngx.var.a8proxy_upstream
   if not name then
      ngx.log(ngx.ERR, "$a8proxy_upstream is not specified")
      return
   end

   local upstream, err = self:getUpstream(name)
   if err then
      ngx.log(ngx.ERR, "error getting upstream " .. name .. ": " .. err)
      return
   end

   if not upstream or table.getn(upstream.upstream.servers) == 0 then
      ngx.log(ngx.ERR, "upstream " .. name .. " is not known or has no known instances")
      return
   end

   --TODO: refactor. Need different LB functions
   local index    = math.random(1, table.getn(upstream.upstream.servers)) % table.getn(upstream.upstream.servers) + 1
   local upstream = upstream.upstream.servers[index]

   local _, err = balancer.set_current_peer(upstream.host, upstream.port)
   if err then
      ngx.log(ngx.ERR, "failed to set current peer for upstream " .. name .. ": " .. err)
      return
   end
end

function Amalgam8:get_version_rules()
   local name = ngx.var.service_name
   if not name then
      ngx.log(ngx.ERR, "$service_name is not specified")
      return nil, nil
   end

   local service, err = self:getService(name)
   if err then
      ngx.log(ngx.ERR, "error getting service " .. name .. ": " .. err)
      return nil, nil
   end

   if not service or not service.version_selector then
      ngx.log(ngx.ERR, "service " .. name .. " or version_selector is not known")
      return nil, nil
   end

   return service.version_selector.default, service.version_selector.selectors
end

return Amalgam8
