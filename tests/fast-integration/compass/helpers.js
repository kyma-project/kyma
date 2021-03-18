const scenariosDefinitionName = "scenarios";

async function removeScenarioFromCompass(client, scenarioName) {
    const scenariosDefinition = await client.queryLabelDefinition(scenariosDefinitionName);
    const idx = scenariosDefinition.schema.items.enum.indexOf(scenarioName);
    if (idx === -1) {
        return;
    }

    scenariosDefinition.schema.items.enum.splice(idx, 1);
    await client.updateLabelDefinition(scenariosDefinitionName, scenariosDefinition.schema);
}

async function addScenarioInCompass(client, scenarioName) {
    const scenariosDefinition = await client.queryLabelDefinition(scenariosDefinitionName);
    if(scenariosDefinition.schema.items.enum.includes(scenarioName)) {
        return;
    }

    scenariosDefinition.schema.items.enum.push(scenarioName);
    await client.updateLabelDefinition(scenariosDefinitionName, scenariosDefinition.schema);
}

async function queryRuntimesForScenario(client, scenarioName) {
    const filter = {
        key: scenariosDefinitionName,
        query: `$[*] ? (@ == "${scenarioName}" )`
    }

    return await client.queryRuntimesWithFilter(filter);
}

async function queryApplicationsForScenario(client, scenarioName) {
    const filter = {
        key: scenariosDefinitionName,
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

module.exports = {
  removeScenarioFromCompass,
  addScenarioInCompass,
  queryRuntimesForScenario,
  queryApplicationsForScenario,
  registerOrReturnApplication
};