-- Called on the request path.
function envoy_on_request(request_handle) 
  -- rewriteCookie(request_handle)
end

function rewriteCookie(request_handle)
  auth = request_handle:headers():get("Authorization")
  cookie = request_handle:headers():get('Cookie')
  
  -- if authorization header is not already present rewrite it from cookie
  if (auth == nil or auth == '') then 
    -- if Cookie exist rewrite it to header
    if (cookie ~= nil ) then 
      token = parse(cookie)  
      if (token ~= nil) then
        -- Add authorization header
        request_handle:headers():add("Authorization", 'Bearer ' .. token)      
      end  
    end
  end  
end

-- Called on the response path.
function envoy_on_response(response_handle)	
end

function parse(token)
	-- match cookie with 'KYMA_TOKEN=' prefix and ';' suffix
  pair = string.match(token, "(KYMA_TOKEN=[A-z0-9.-]*[^;?])")
  if (pair ~= nil) then
	  -- get value from pair
    i = string.find(pair, "=")
    if (i ~= nil) then
      return string.sub(pair, i+1)
    end  
  end
end

