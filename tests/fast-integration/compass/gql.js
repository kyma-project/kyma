function registerApplication(appName, scenarioName) {
    return `mutation { result: registerApplication(in: { name: \\"${appName}\\" labels: { scenarios: [\\"${scenarioName}\\"] } }){id} }`;
}

function unregisterApplication(applicationID) {
    return `mutation { result: unregisterApplication(id: \\"${applicationID}\\") { id } }`;
}

function registerRuntime(runtimeName, scenarioName) {
    return `mutation { result: registerRuntime(in: { name: \\"${runtimeName}\\" labels: { scenarios: [\\"${scenarioName}\\"] } }){id} }`;
}

function unregisterRuntime(runtimeID) {
    return `mutation { result: unregisterRuntime(id: \\"${runtimeID}\\") { id } }`
}

function requestOneTimeTokenForApplication(appID) {
    return `mutation { result: requestOneTimeTokenForApplication(id: \\"${appID}\\"){ token connectorURL } }`;
}

function requestOneTimeTokenForRuntime(runtimeID) {
    return `mutation { result: requestOneTimeTokenForRuntime(id: \\"${runtimeID}\\"){ token connectorURL } }`;
}

function queryLabelDefinition(key) {
    return `query { result: labelDefinition(key: \\"${key}\\") { key schema } }`
}

function updateLabelDefinition(key, schema) {
    const schemaSerialized = escapeForGQL(JSON.stringify(schema));
    return `mutation { result: updateLabelDefinition(in: { key: \\"${key}\\" schema: \\"${schemaSerialized}\\" } ) { key schema } }`
}

function queryRuntimesWithFilter(filter) {
    const querySerialized = escapeForGQL(filter.query);
    return `query { result: runtimes(filter: { key: \\"${filter.key}\\", query: \\"${querySerialized}\\" }) { data { id name } } }`;
}

function queryApplicationsWithFilter(filter) {
    const querySerialized = escapeForGQL(filter.query);
    return `query { result: applications(filter: { key: \\"${filter.key}\\", query: \\"${querySerialized}\\" }) { data { id name } } }`;
}

function escapeForGQL(str) {
    return str.split('"').join(`\\\\\\"`);
}

module.exports = {
    registerApplication,
    unregisterApplication,
    registerRuntime,
    unregisterRuntime,
    requestOneTimeTokenForApplication,
    requestOneTimeTokenForRuntime,
    queryLabelDefinition,
    updateLabelDefinition,
    queryRuntimesWithFilter,
    queryApplicationsWithFilter,
}