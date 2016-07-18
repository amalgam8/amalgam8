local json     = require("cjson.safe")
local math     = require("math")
local balancer = require("ngx.balancer")
local string   = require("resty.string")

local Amalgam8 = { _VERSION = '0.2.0',
	     upstreams_dict = {},
	     faults_dict = {},
	     services_dict = {}
}

local function isValidString(input)
   if input and type(input) == 'string' and input ~= '' then
      return true
   else
      return false
   end
end

local function isValidNumber(input)
   if input and type(input) == 'number' then
      return true
   else
      return false
   end
end

function Amalgam8:decodeJson(input)
   local upstreams = {}
   local services = {}
   local faults = {}

   if input.upstreams then
      for name, upstream in pairs(input.upstreams) do
         if upstream.servers then
            upstreams[name] = {
               servers = {},
            }

            for _, server in ipairs(upstream.servers) do
               if isValidString(server.host) and isValidNumber(server.port) and server.port > 0 then
                  table.insert(upstreams[name].servers, {
                                  host   = server.host,
                                  port   = server.port,
                  })
               else
                  return nil, nil, nil, "invalid data format for upstream server host or port"
               end
            end
         end
      end
   end
   
   if input.services then
      for name, metadata in pairs(input.services) do
         if isValidString(metadata.default) then
            if not isValidString(metadata.service_type) then
               metadata.service_type = 'http'
            end
            if metadata.service_type == 'http' or metadata.service_type == 'https' then
               services[name] = {
                  default = metadata.default,
                  service_type = metadata.service_type,
                  selectors = metadata.selectors
               }
            else
               return nil, nil, nil, "invalid service_type. Only http|https is allowed"
            end
         else
            return nil, nil, nil, "invalid/empty value for default service version"
         end
      end
   end

   if input.faults then
      for i, fault in ipairs(input.faults) do
         if isValidString(fault.source) and isValidString(fault.destination) and isValidString(fault.header) and isValidString(fault.pattern) then
            if fault.source == self.source_service then
               if not isValidNumber(fault.delay) then
                  fault.delay = 0.0
               end
               if not isValidNumber(fault.return_code) then
                  fault.return_code = nil
               end
               if not isValidNumber(fault.abort_probability) then
                  fault.abort_probability = 0.0
               end
               if not isValidNumber(fault.delay_probability) then
                  fault.delay_probability = 0.0
               end
               if (fault.abort_probability > 0.0 and fault.return_code ) or (fault.delay_probability > 0.0 and fault.delay > 0.0) then
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
               else
                  return nil, nil, nil, "Invalid fault entry. Atleast one of abort/delay fault details should be non zero"
               end
            else
               ngx.log(ngx.DEBUG, "Ignoring fault entry " .. fault.source .. " as it does not match source_service " .. self.source_service)
            end
         else
            return nil, nil, nil, "Invalid fault entry. Found empty source/destination/header/pattern" .. fault.source .. "," .. fault.destination .. ",".. fault.header .. "," .. fault.pattern
         end
      end
   else
      ngx.log(ngx.DEBUG, "NOTE: No fault entries specified")
   end
   return upstreams, services, faults, nil
end

local function getMyName()
   local service_name = os.getenv("A8_SERVICE")
--   local service_version = os.getenv("A8_SERVICE_VERSION")

   if not isValidString(service_name) then
      return nil, "A8_SERVICE environment variable is either empty or not set"
   end

   return service_name, nil
   -- if isValidString(service_version) then
   --    return service_name .. "_" .. service_version, nil
   -- else
   --    return service_name, nil
   -- end
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

   local source_service, err = getMyName()
   if err then
      return err
   end
 
   self.upstreams_dict = upstreams
   self.services_dict = services
   self.faults_dict = faults
   self.source_service = source_service
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

function Amalgam8:updateService(name, metadata)

   local current = {
      name         = name,
      metadata   = metadata,
   }

   if not isValidString(metadata.default) then
      return "null entry for default version"
   end

   if metadata.selectors then
      -- convert the selectors field from string to lua table
      local selectors_fn, err = loadstring( "return " .. metadata.selectors)
      if err then
         ngx.log(ngx.ERR, "failed to compile selector", err)
         return err
      end
      current.metadata.selectors = selectors_fn()
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
   local upstreams, services, faults, err = self:decodeJson(input)

   if err then
      return err
   end

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

   for name, metadata in pairs(services) do
      err = self:updateService(name, metadata)
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

function Amalgam8:getFault(index)
   local serialized = self.faults_dict:get(index)
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
         ngx.say("error decoding input json: " .. err)
         ngx.exit(ngx.HTTP_BAD_REQUEST)
      end

      -- TODO: FIXME. There is no concept of locking so far.
      self:resetState()
      err = self:updateProxy(input)
      if err then
         ngx.log(ngx.ERR, "error updating proxy: " .. err)
         ngx.say("error updating internal state: " .. err)
         ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
      end
   elseif ngx.req.get_method() == "GET" then
      local state = {
         services = {},
         upstreams = {},
         faults = {}
      }

      -- TODO: This fetches utmost 1024 keys only
      local services = self.services_dict:get_keys()
      local upstreams = self.upstreams_dict:get_keys()
      local faults = self.faults_dict:get_keys()

      for _,name in ipairs(services) do
         state.services[name] = self:getService(name)
      end

      for _,name in ipairs(upstreams) do
         state.upstreams[name] = self:getUpstream(name)
      end
      local len = self.faults_dict:get("len")
      for i=1,len do
         table.insert(state.faults, self:getFault(i))
      end

      local output, err = json.encode(state)
      if err then
         ngx.say("error encoding state as JSON:" .. err)
         ngx.log(ngx.ERR, "error encoding state as JSON:" .. err)
         nex.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
      end
      ngx.header["content-type"] = "application/json"
      ngx.say(output)
      ngx.exit(ngx.HTTP_OK)
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

function Amalgam8:get_service_metadata()
   local name = ngx.var.service_name
   if not name then
      ngx.log(ngx.ERR, "$service_name is not specified")
      return nil, nil, nil
   end
   
   local service, err = self:getService(name)
   if err then
      ngx.log(ngx.ERR, "error getting service " .. name .. ": " .. err)
      return nil, nil, nil
   end
   
   if not service or not service.metadata then
      ngx.log(ngx.ERR, "service " .. name .. " or metadata is not known")
      return nil, nil, nil
   end
   
   return service.metadata.service_type, service.metadata.default, service.metadata.selectors
end

function Amalgam8:get_source_service()
   return self.source_service
end

return Amalgam8
