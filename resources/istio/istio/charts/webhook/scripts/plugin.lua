-- Called on the request path.
function envoy_on_request(request_handle)
    GraphQlAuthPlugin_OnRequest(request_handle)
end

-- Called on the response path.
function envoy_on_response(response_handle)
end

function GraphQlAuthPlugin_OnRequest(request_handle)

    -- it does not work - i'm not able to read host header
    -- local authority = request_handle:headers():get(":authority")
    -- if (authority == "core-ui-api.kyma-system.svc.cluster.local") then
    local path = request_handle:headers():get(":path")
    if (path ~= nil) and (path == "/graphql") then

        request_handle:headers():replace("kyma-graphql-parsed", "true")

        local query = request_handle:body():getBytes(0, request_handle:body():length())
        local queryItems = parseGraphQlQuery(query)
        if (queryItems.resources ~= "") then
            request_handle:headers():replace("kyma-graphql-resources", queryItems.resources)
        end
    end
end

function parseGraphQlQuery(payload)

    local result = {
        name = "",
        resources = "{"
    }

    if (payload ~= nil) and (payload ~= "") then

        local payloadItems = {
            name = "",
            resourcesArr = {}
        }

        local newPayload = string.gsub(payload, "\\n", " ")

        for query in string.gmatch(newPayload, 'query .+') do

            local handleFunc = handleQueryStatement
            for w in string.gmatch(query, "[%w,{,},(,)]+") do

                if handleFunc ~= nil then
                    handleFunc = handleFunc(w, payloadItems)
                end
            end
        end

        for mutation in string.gmatch(newPayload, 'mutation .+') do

            local handleFunc = handleMutationStatement
            for w in string.gmatch(mutation, "[%w,{,},(,)]+") do

                if handleFunc ~= nil then
                    handleFunc = handleFunc(w, payloadItems)
                end
            end
        end

        table.sort(payloadItems.resourcesArr)

        result.name = payloadItems.name

        local resouresArrLen = #payloadItems.resourcesArr
        if resouresArrLen > 0 then
            result.resources = result.resources .. payloadItems.resourcesArr[1]
            for i = 2, resouresArrLen do
                result.resources = result.resources .. "," .. payloadItems.resourcesArr[i]
            end
        end
    end

    result.resources = result.resources .. "}"
    return result
end

function handleQueryStatement(item)

    if (item == "query") then
        return handleQueryStart
    end
    -- if not query statement try again until query statement found
    return handleQueryStatement
end

function handleMutationStatement(item)

    if (item == "mutation") then
        -- TODO implement mutations
        return nil
    end
    -- if not query statement try again until query statement found
    return handleMutationStatement
end

function handleQueryStart(item, queryItems)

    if (item == "{") then
        return handleQueryBodyOpen(item)
    end
    return handleQueryName(item, queryItems)
end

function handleQueryName(item, queryItems)

    local attribsStart = item:find("[(]")
    if (attribsStart ~= nil) and (attribsStart > 0) then
        queryItems.name = item:sub(0, attribsStart-1)
        return skipAttribsBlockFunc(handleQueryBodyOpen, queryItems)
    end

    queryItems.name = item
    return handleQueryBodyOpen
end

function handleQueryBodyOpen(item)

    if (item == "{") then
        return handleQueryResource
    end
    return nil
end

function handleQueryResource(item, queryItems)

    if item == "}" then
        return nil
    end

    if (item:find("^{") ~= nil) then
        if (item:find("}$") ~= nil) then
            return handleQueryResource
        else
            return skipQueryBlockFunc(handleQueryResource, queryItems)
        end
    end

    local fieldsStart = item:find("{")
    if (fieldsStart ~= nil) and (fieldsStart > 0) then

        local resourceName = item:sub(0, fieldsStart-1)

        if(resourceName ~= nil and (resourceName:len() > 0)) then
            table.insert(queryItems.resourcesArr, resourceName)
        end

        local fieldsEnd = item:find("}")
        if (fieldsEnd ~= nil) then
            return handleQueryResource(item:sub(fieldsEnd), queryItems)
        else
            return skipQueryBlockFunc(handleQueryResource, queryItems)
        end
    end

    local attribsStart = item:find("[(]")
    if (attribsStart ~= nil) and (attribsStart > 0) then
        local resourceName = item:sub(0, attribsStart-1)
        table.insert(queryItems.resourcesArr, resourceName)
        return skipAttribsBlockFunc(handleQueryResource, queryItems)
    else
        table.insert(queryItems.resourcesArr, item)
        return handleQueryResource
    end
end

function skipQueryBlockFunc(nextFunc, queryItems)

    local function skipQueryBlock(item)

        local endIndex = item:find("^}$")
        if (endIndex ~= nil) then

            if (endIndex < item:len()) then

                local nextWord = item:sub(endIndex+1)
                return nextFunc(nextWord, queryItems)
            end

            return nextFunc
        end
        return skipQueryBlock
    end

    return skipQueryBlock
end

function skipAttribsBlockFunc(nextFunc, queryItems)

    local function skipAttribsBlock(item)

        local endIndex = item:find("[)]")
        if (endIndex ~= nil) then

            if (endIndex < item:len()) then

                local nextWord = item:sub(endIndex+1)
                return nextFunc(nextWord, queryItems)
            end

            return nextFunc
        end
        return skipAttribsBlock
    end

    return skipAttribsBlock
end
