local json     = require("cjson.safe")
local math     = require("math")
local balancer = require("ngx.balancer")
local string   = require("resty.string")
local cmsgpack = require("cmsgpack.safe")

local ngx_log = ngx.log
local ngx_ERR = ngx.ERR
local ngx_DEBUG = ngx.DEBUG
local ngx_INFO = ngx.INFO
local ngx_WARN = ngx.WARN
local ngx_var = ngx.var
local ngx_shared = ngx.shared
local tostring = tostring

--TODO: This code has been tested with a single worker only
-----With multiple workers, we might need a read-write lock on the
-----shared state (a8_instances, a8_rules) so that workers do not see
-----intermediate state of system while an update is in progress. That could
-----potentially result in erroneous routing. Since the sidecar operates as a
-----"sidecar", along side the app, one worker should be sufficient in most
-----situations. Hence, no effort has been made to protect accesses to the
-----global shared state with locks. The performance of lua-resty-lock needs
-----to be evaluated at high load.
local Amalgam8 = { _VERSION = '0.3.0' }

local function is_valid_string(input)
   if input and type(input) == 'string' and input ~= '' then
      return true
   else
      return false
   end
end


local function is_valid_number(input)
   if input and type(input) == 'number' then
      return true
   else
      return false
   end
end


local function get_name_and_tags()
   local service_name = os.getenv("A8_SERVICE")

   if not is_valid_string(service_name) then
      return nil, nil, "A8_SERVICE environment variable is either empty or not set"
   end

   local tags = {}
   local name, version
   for x,y,z in string.gmatch(service_name, '([^:]+)(:?(.*))') do
      name = x
      version = y
      break
   end

   return name, version, nil
   -- if version then
   --    for t in string.gmatch(version, '([^:]+)') do
   --       table.insert(tags, t)
   --    end
   -- end
   
   -- return name, tags, nil
end


local function create_instance(i)
   local instance = {}
   instance.type = i.type
   instance.tags = table.concat(i.tags, ",")
   if not i.type then
      instance.type = 'http'
   end
   for x,y,z in string.gmatch(i.endpoint.value, '([^:]+)(:?(.*))') do
      instance.host = x
      instance.port = y
      break
   end
   if not instance.port then
      if instance.type == 'http' then
         instance.port = 80
      elseif instance.type == 'https' then
         instance.port = 443
      else
         ngx_log(ngx_WARN, "ignoring endpoint " .. instance.host .. ", missing port")
         return nil
      end
   end
   -- TODO: add support for hostnames in endpoints
   return instance
end


local function compare_rules(a, b)
   return a.priority < b.priority
end


local function compare_backends(a, b)
   return a.weight < b.weight
end

local function match_tags(src_tag_string, dst_tag_set) --assumes both are not nil
   -- local empty_src_tags = (string.len(src_tag_string) == 0)
   -- local empty_dst_tags = (not dst_tag_set)
   local match = true
   for _, t in ipairs(dst_tag_set) do
      if not string.find(src_tag_string, t, 1, true) then
         match = false
         break
      end
   end
   return match
end

local function check_and_preprocess_match(myname, mytags, match_type, match_sub_block)
   if not match_sub_block then
      if match_type ~= "none" then
         return true, nil
      else
         return false, nil -- return false for empty none block
      end
   end

   local match_cond = {}
   local is_my_tag_empty = (string.len(mytags) == 0)
   local match_found = false

   for _, m in ipairs(match_sub_block) do
      local check1 = false
      local check2 = false

      if m.source then 
         local empty_src_tags = (not m.source.tags)
         local empty_src_name = (not m.source.name)
         local tags = m.source.tags

         --src has a name
         -- name match, with empty rule src tags
         -- src tags and my tags matched
         check1 = (not empty_src_name and (m.source.name == myname) and (empty_src_tags or (not is_my_tag_empty and match_tags(mytags, tags))))

         -- src has no service name. But it should have tags
         check2 = (empty_src_name and (not is_my_tag_empty and match_tags(mytags, tags)))
      else
         -- empty source implies wildcard source. Rule match found for all/any. continue collecting other fields.
         if match_type ~= "none" then check1 = true end
      end
      
      if check1 or check2 then
         match_found = true
         if m.headers then
            for k,v in pairs(m.headers) do
               match_cond[k]=v
            end
         end
      end
   end

   return match_found, match_cond
end


local function is_rule_for_me(myname, mytags, rule)
   --by default we assume rule is for us if the
   -- A. (match.(all|any|none).source is nil) OR
   -- (B. (the source field in all|any sub match matches service name and tags) AND
   --  C. (the source field in none sub match DOES NOT match service name and tags))
   -- If the source field has no service name but tags are provided, then we compare our tags
   -- with the source tags.
   ---- If our service tags are empty, then try next source block
   -- if source field has service name but no service tags, then the rule applies to us if service name matches
   -- if source field has no service name but has service tags, then the rule applies to us if tags match
   ----if our service has no tags, then try next source block

   --right now, only ALL is supported

   if not rule.match then return true end

   local match = {}
   local all_headers, any_headers, none_headers
   local all_res, any_res, none_res
 
   if rule.match.source or rule.match.headers then
      match["source"] = rule.match.source
      match["headers"] = rule.match.headers
      if not rule.match.all then
         rule.match.all = {}
      end
      table.insert(rule.match.all, match)
   end

   all_res, all_headers = check_and_consolidate_match(myname, mytags, "all", rule.match.all)
   any_res, any_headers = check_and_consolidate_match(myname, mytags, "any", rule.match.any)
   none_res, none_headers = check_and_consolidate_match(myname, mytags, "none", rule.match.none)

   local res = (all_res or any_res) and not none_res
   if res then
      rule.match = {
         all = all_headers,
         any = any_headers,
         none = none_headers
      }
   end
   return res
end


local function create_rule(rule, myname, mytags)
   -- no match implies wildcard rule from * to *
   if not is_rule_for_me(myname, mytags, rule) then
      return nil
   end

   if not rule.priority then
      rule.priority = 0
   end

   if not rule.routes or not rule.actions then
      return nil
   end

   -- set default weights for backends where no weight is specified
   ---- Take the leftover weight and distribute it equally among
   ---- unweighted backends
   if rule.route.backends then
      local sum = 0
      local unweighted = 0
      for _, b in ipairs(rule.route.backends) do
         if b.weight then
            sum = sum + b.weight
         else
            unweighted = unweighted + 1
         end
      end
      if unweighted > 0 then
         local balance = 1 - sum
         for _, b in ipairs(rule.route.backends) do
            if not b.weight then
               b.weight = balance/unweighted
            end
         end
      end
      table.sort(rule.route.backends, compare_backends)
   end

   return rule
end


function Amalgam8:new()
   o = {}
   setmetatable(o, self)
   self.__index = self

   local myname, mytags, err = get_name_and_tags()
   if err then
      return err
   end

   self.myname = myname
   self.mytags = mytags
   return o
end


-- a8_instances stores registry info
-- a8_rules stores rules data, keyed by destination. Each entry is an array of
-- rules for that destination, ordered by priority.

-- Select rule from a8_rules[destination]
-- select backend and obtain the backend's tags
-- then do delay/abort if backend's tags match
-- then fetchInstances of the backend based on its tags and load balance.

-- Two functions are needed for rule matching:
--  get_instances(service_name, tags)
--  match_instance_tags(instance, tags)
function Amalgam8:update_state(input)
   local a8_instances = {}
   local a8_rules = {}
   local err

   if input.instances then
       for _, e in ipairs(input.instances) do
         if not a8_instances[e.service_name] then
            a8_instances[e.service_name] = {}
         end
         local instance = create_instance(e)
         if instance then
            table.insert(a8_instances[e.service_name], instance)
         end
       end

       for service, instances in pairs(a8_instances) do
         serialized = cmsgpack.pack(instances)
         _, err = ngx_shared.a8_instances:set(service, serialized)
         if err then
            err = "failed to update instances for service:"..service..":"..err
            return err
         end
      end

   end

   if input.rules then
      for _, r in ipairs(input.rules) do
         if not a8_rules[r.destination] then
            a8_rules[r.destination] = {}
         end
         local rule = create_rule(r, self.myname, self.mytags)
         if rule then
            table.insert(a8_rules[r.destination], rule)
         end
      end

      for service, rset in pairs(a8_rules) do
         table.sort(rset, compare_rules)
         serialized = cmsgpack.pack(rset)
         _, err = ngx_shared.a8_rules:set(service, serialized)
         if err then
            err = "failed to update rules for service:"..service..":"..err
            return err
         end
      end

   end

   return nil
end

function Amalgam8:get_instances(service)
   local serialized = ngx_shared.a8_instances:get(service)
   if serialized then
      return cmsgpack.unpack(serialized)
   end
end

function Amalgam8:get_rules(service)
   local serialized = ngx_shared.a8_rules:get(service)
   if serialized then
      return cmsgpack.unpack(serialized)
   end
end

function Amalgam8:reset_state()
   ngx_shared.a8_instances:flush_all()
   ngx_shared.a8_instances:flush_expired()
   ngx_shared.a8_rules:flush_all()
   ngx_shared.a8_rules:flush_expired()
end

function Amalgam8:proxy_admin()
   if ngx.req.get_method() == "PUT" or ngx.req.get_method() == "POST" then
      ngx.req.read_body()

      local input, err = cmsgpack.unpack(ngx.req.get_body_data())
      if err then
         ngx.status = ngx.HTTP_BAD_REQUEST
         ngx_log(ngx_ERR, "error decoding input json: " .. err)
         ngx.say("error decoding input json: " .. err)
         ngx.exit(ngx.status)
      end

      -- TODO: locking in multi-worker context.
      self:reset_state()
      err = self:update_state(input)
      if err then
         ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
         ngx_log(ngx_ERR, "error updating proxy: " .. err)
         ngx.say("error updating internal state: " .. err)
         ngx.exit(ngx.status)
      end
      ngx.exit(ngx.HTTP_OK)
   elseif ngx.req.get_method() == "GET" then
      local state = {
         instances = {},
         rules = {}
      }

      -- TODO: This fetches utmost 1024 keys only
      local instance_keys = ngx_shared.a8_instances:get_keys()
      local rule_keys = ngx_shared.a8_rules:get_keys()

      for _,service in ipairs(instance_keys) do
         state.instances[service] = self:get_instances(service)
      end
      for _,service in ipairs(rule_keys) do
         state.rules[service] = self:get_rules(service)
      end

      local output, err = json.encode(state)
      if err then
         ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
         ngx.say("error encoding state as JSON:" .. err)
         ngx_log(ngx_ERR, "error encoding state as JSON:" .. err)
         nex.exit(ngx.status)
      end
      ngx.header["content-type"] = "application/json"
      ngx.say(output)
      ngx.exit(ngx.HTTP_OK)
   else
      return ngx.exit(ngx.HTTP_BAD_REQUEST)
   end
end

-- function Amalgam8:balance()
--    local name = ngx_var.service_name
--    local tags = ngx.var.service_tags
--    if not name then
--       ngx.status = ngx.HTTP_BAD_GATEWAY
--       ngx_log(ngx_ERR, "$a8proxy_instance is not specified")
--       ngx.exit(ngx.status)
--       return
--    end

--    local instances = self:get_instances(name)
--    if not instances or table.getn(instances) == 0 then
--       ngx.status = ngx.HTTP_NOT_FOUND
--       ngx_log(ngx_ERR, "service " .. name .. " is not known or has no known instances")
--       ngx.exit(ngx.status)
--       return
--    end

--    --TODO: refactor. Need different LB functions
--    local index    = math.random(1, table.getn(instances)) % table.getn(instances) + 1
--    local instance = instances[index]

--    local _, err = balancer.set_current_peer(instance.host, instance.port)
--    if err then
--       ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
--       ngx_log(ngx_ERR, "failed to set current peer for instance " .. name .. ": " .. err)
--       ngx.exit(ngx.status)
--       return
--    end
-- end

-- function Amalgam8:get_service_metadata()
--    local name = ngx_var.service_name
--    if not name then
--       ngx_log(ngx_ERR, "$service_name is not specified")
--       return nil, nil, nil
--    end

--    local service, err = self:getService(name)
--    if err then
--       ngx_log(ngx_ERR, "error getting service " .. name .. ": " .. err)
--       return nil, nil, nil
--    end

--    if not service or not service.metadata then
--       ngx_log(ngx_ERR, "service " .. name .. " or metadata is not known")
--       return nil, nil, nil
--    end

--    return service.metadata.service_type, service.metadata.default, service.metadata.selectors
-- end

-- function Amalgam8:get_myname()
--    return self.myname
-- end

return Amalgam8
