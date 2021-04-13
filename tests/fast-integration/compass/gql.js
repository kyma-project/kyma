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

function setRuntimeLabel(runtimeID, key, value) {
    value = prepareLabelValue(value);
    return `mutation { result: setRuntimeLabel(runtimeID: \\"${runtimeID}\\" key: \\"${key}\\" value: ${value}) { key value } }`;
}

function queryRuntime(runtimeID) {
    return `query { result: runtime(id: \\"${runtimeID}\\") { id name labels status { condition } } }`;
}

function queryApplication(appID) {
    return `query { result: application(id: \\"${appID}\\") { id name labels } }`;
}

function setApplicationLabel(appID, key, value) {
    value = prepareLabelValue(value);
    return `mutation { result: setApplicationLabel(applicationID: \\"${appID}\\" key: \\"${key}\\" value: ${value}) { key value } }`;
}

function deleteApplicationLabel(appID, key) {
    return `mutation { result: deleteApplicationLabel(applicationID: \\"${appID}\\", key: \\"${key}\\") { key value } }`;
}

function escapeForGQL(str) {
    return str.split('"').join(`\\\\\\"`);
}

function escapeForGQLArray(str) {
    return str.split('"').join(`\\\"`);
}

function prepareLabelValue(value) {
    if (typeof value !== "string") {
        value = Array.isArray(value) 
            ? escapeForGQLArray(JSON.stringify(value))
            : escapeForGQL(JSON.stringify(value));
    } else {
        value = `\\"${value}\\"`
    }

    return value;
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
    setRuntimeLabel,
    queryRuntime,
    queryApplication,
    setApplicationLabel,
    deleteApplicationLabel,
}