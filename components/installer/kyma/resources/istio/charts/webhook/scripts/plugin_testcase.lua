require "plugin"
lu = require "luaunit"

TestParseGraphQlQuery = {}

-- Tests for sample queries:

function TestParseGraphQlQuery:test_should_not_fail_on_nil()
    -- given
    local input
    -- when
    local actual = parseGraphQlQuery(input)
    -- then
    local expected = {name = "", resources = "{}"}
    lu.assertEquals(actual, expected)
end

function TestParseGraphQlQuery:test_should_not_fail_on_empty_string()
    -- given
    local input = ""
    -- when
    local actual = parseGraphQlQuery(input)
    -- then
    local expected = {name = "", resources = "{}"}
    lu.assertEquals(actual, expected)
end

function TestParseGraphQlQuery:test_should_parse_simple_query()
    -- given
    local input = [[ {"query":"query {\n  apis(environment: \"default\") {\n    name\n  }\n} ","variables":null,"operationName":""} ]]
    -- when
    local actual = parseGraphQlQuery(input)
    -- then
    local expected = {
        name = "",
        resources = "{apis}"
    }
    lu.assertEquals(actual, expected)
end

function TestParseGraphQlQuery:test_should_parse_named_query()
    -- given
    local input = [[ {"query":"query a {\n  apis(environment: \"default\") {\n    name\n  }\n} ","variables":null,"operationName":"a"} ]]
    -- when
    local actual = parseGraphQlQuery(input)
    -- then
    local expected = {
        name = "a",
        resources = "{apis}"
    }
    lu.assertEquals(actual, expected)
end

function TestParseGraphQlQuery:test_should_parse_with_no_whitechars_in_attribs()
    -- given
    local input = [[ {"query":"query deployments { groups {name} clusterroles {id} }"} ]]
    -- when
    local actual = parseGraphQlQuery(input)
    -- then
    local expected = {
        name = "deployments",
        resources = "{clusterroles,groups}"
    }
    lu.assertEquals(actual, expected)
end

function TestParseGraphQlQuery:test_should_parse_with_no_whitechars_in_attribs_start()
    -- given
    local input = [[ {"query":"query deployments { groups {name } clusterroles {id } }"} ]]
    -- when
    local actual = parseGraphQlQuery(input)
    -- then
    local expected = {
        name = "deployments",
        resources = "{clusterroles,groups}"
    }
    lu.assertEquals(actual, expected)
end

-- Tests for real Kyma queries:

function TestParseGraphQlQuery:test_should_parse_get_remote_environments()
    -- given
    local input = [[ {"query":"query {\n  remoteEnvironments {\n    name        \n    status\n    enabledInEnvironments\n    source {\n      type\n    }\n  }\n}","variables":null,"operationName":"a"} ]]
    --local input = [[ {"query":"query {\n      remoteEnvironments{\n        name\n        status\n        enabledInEnvironments\n        source {\n          type\n        }\n      }\n    }"} ]]
    -- when
    local actual = parseGraphQlQuery(input)
    -- then
    local expected = {
        name = "",
        resources = "{remoteEnvironments}"
    }
    lu.assertEquals(actual, expected)
end

function TestParseGraphQlQuery:test_should_parse_get_remote_environments_when_name_with_attribs()
    -- given
    local input = [[ {"query":"query RemoteEnvironment($environment: String!){\n      remoteEnvironments(environment: $environment) {\n        name\n      }\n    }","variables":{"environment":"stage"}} ]]
    -- when
    local actual = parseGraphQlQuery(input)
    -- then
    local expected = {
        name = "RemoteEnvironment",
        resources = "{remoteEnvironments}"
    }
    lu.assertEquals(actual, expected)
end

function TestParseGraphQlQuery:test_should_parse_get_deployments()
    -- given
    local input = [[ {"query":"query Deployments($environment: String!) {\n      deployments(environment: $environment) {\n        name\n        boundServiceInstanceNames\n        labels\n        creationTimestamp\n        status {\n          replicas\n          updatedReplicas\n          readyReplicas\n          availableReplicas\n          conditions {\n            status\n            type\n            lastTransitionTimestamp\n            lastUpdateTimestamp\n            message\n            reason\n          }\n        }\n        containers {\n          name\n          image\n        }\n      }\n    }","variables":{"environment":"stage"}} ]]
    -- when
    local actual = parseGraphQlQuery(input)
    -- then
    local expected = {
        name = "Deployments",
        resources = "{deployments}"
    }
    lu.assertEquals(actual, expected)
end

os.exit(lu.LuaUnit.run())