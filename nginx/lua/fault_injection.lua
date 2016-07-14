--[[
   Amalgam8 fault injection library
]]

local json     = require("cjson.safe")
local math     = require("math")

function inject_faults(destination, faults_dict)
   --[[
      format of faults object: json array, where each element is an object of the form
      {
      "source":"",
      "destination":"",
      "header": "",
      "pattern":"",
      "delay": 0.0,
      "delay_probability":0.0,
      "abort_probability":0.0,
      "return_code":0,
      }
      source, destination cannot be wildcards
   ]]
   local len = faults_dict:get("len")
   if len == 0 then
      return
   end
   for i=1,len do
      local fault_str = faults_dict:get(i)
      local obj, err = json.decode(fault_str)
      if err then
         ngx.log(ngx.ERR, "error decoding fault spec for " .. destination .. " err:" .. err)
      else
         local fault = obj.fault
         if fault then
            if fault.destination == destination then
               -- ngx.log(ngx.NOTICE, "matched destination " .. destination)
               local header = ngx.req.get_headers()[fault.header]
               if header then
                  -- ngx.log(ngx.NOTICE, "matched header " .. fault.header .. "val:" .. header)
                  local m, err = ngx.re.match(header, fault.pattern, "o")
                  if m then
                     -- ngx.log(ngx.NOTICE, "pattern match successful " .. fault.pattern)
                     --TODO: Need to figure out whether header and value should be logged irrespective of destination, in terms of gremlin assertions.
                     ngx.var.gremlin_header_name = fault.header
                     ngx.var.gremlin_header_val = header

                     -- DELAYS
                     if fault.delay_probability > 0.0 then
                        if fault.delay_probability < 1.0 then
                           if math.random() < fault.delay_probability then
                              ngx.sleep(fault.delay)
                           end
                        else
                           ngx.sleep(fault.delay)
                        end
                     end

                     -- ABORTS
                     if fault.abort_probability > 0.0 then
                        if fault.abort_probability < 1.0 then
                           if math.random() < fault.abort_probability then
                              return ngx.exit(fault.return_code)
                           end
                        else
                           return ngx.exit(fault.return_code)
                        end
                     end
                  end
               end
            end
         end
      end
   end
end

return {
   inject_faults = inject_faults
}
