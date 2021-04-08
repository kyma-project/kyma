const SCENARIOS_DEFINITION_NAME = "scenarios";

async function removeScenarioFromCompass(client, scenarioName) {
    const scenariosDefinition = await client.queryLabelDefinition(SCENARIOS_DEFINITION_NAME);
    const idx = scenariosDefinition.schema.items.enum.indexOf(scenarioName);
    if (idx === -1) {
        return;
    }

    scenariosDefinition.schema.items.enum.splice(idx, 1);
    await client.updateLabelDefinition(SCENARIOS_DEFINITION_NAME, scenariosDefinition.schema);
}

async function addScenarioInCompass(client, scenarioName) {
    const scenariosDefinition = await client.queryLabelDefinition(SCENARIOS_DEFINITION_NAME);
    if(scenariosDefinition.schema.items.enum.includes(scenarioName)) {
        return;
    }

    scenariosDefinition.schema.items.enum.push(scenarioName);
    await client.updateLabelDefinition(SCENARIOS_DEFINITION_NAME, scenariosDefinition.schema);
}

async function queryRuntimesForScenario(client, scenarioName) {
    const filter = {
        key: SCENARIOS_DEFINITION_NAME,
        query: `$[*] ? (@ == "${scenarioName}" )`
    }

    return await client.queryRuntimesWithFilter(filter);
}

async function queryApplicationsForScenario(client, scenarioName) {
    const filter = {
        key: SCENARIOS_DEFINITION_NAME,
        query: `$[*] ? (@ == "${scenarioName}" )`
    }

    return await client.queryApplicationsWithFilter(filter);
}

async function registerOrReturnApplication(client, appName, scenarioName) {
    const applications = await queryApplicationsForScenario(client, scenarioName);
    const filtered = applications.filter((app) => app.name === appName);
    if (filtered.length > 0) {
        return filtered[0].id;
    }

    return await client.registerApplication(appName, scenarioName);
}

async function assignRuntimeToScenario(client, runtimeID, scenarioName) {
    const runtime = await client.getRuntime(runtimeID);
    if(!runtime.labels[SCENARIOS_DEFINITION_NAME]) {
        runtime.labels[SCENARIOS_DEFINITION_NAME] = [];
    }

    const scenarios = runtime.labels[SCENARIOS_DEFINITION_NAME];
    scenarios.push(scenarioName);

    return await client.setRuntimeLabel(runtimeID, SCENARIOS_DEFINITION_NAME, scenarios);
}

module.exports = {
  removeScenarioFromCompass,
  addScenarioInCompass,
  queryRuntimesForScenario,
  queryApplicationsForScenario,
  registerOrReturnApplication,
  assignRuntimeToScenario
};