const {
  addScenarioInCompass,
  removeScenarioFromCompass,
  queryRuntimesForScenario,
  queryApplicationsForScenario,
  removeApplicationFromScenario,
  unassignRuntimeFromScenario
} = require('./helpers');

const {
  debug,
  toBase64,
  k8sApply,
  waitForCompassConnection,
  deleteAllK8sResources,
} = require('../utils');

//const UNREGISTER = process.env.UNREGISTER_FROM_COMPASS === "true" ;

async function registerKymaInCompass(client, runtimeName, scenarioName) {
  await addScenarioInCompass(client, scenarioName);
  const runtimeID = await client.registerRuntime(runtimeName, scenarioName);
  debug(`Runtime ID in Compass ${runtimeID}`);

  const pairingData = await client.requestOneTimeTokenForRuntime(runtimeID);
  const compassAgentCfg = {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: 'compass-agent-configuration',
    },
    data: {
      CONNECTOR_URL: toBase64(pairingData.connectorURL),
      RUNTIME_ID: toBase64(runtimeID),
      TENANT: toBase64(client.tenantID),
      TOKEN: toBase64(pairingData.token),
    },
  };
  await k8sApply([compassAgentCfg], 'compass-system');
  await waitForCompassConnection('compass-connection');
}

async function unregisterScenarioFromCompass(client, scenarioName) {
  // Cleanup Compass
  const applications = await queryApplicationsForScenario(client, scenarioName);
  for (const application of applications) {
    await removeApplicationFromScenario(client, application.id, scenarioName);
    await client.unregisterApplication(application.id);
  }

  // TODO: refactor this step to cover runtime agent deleting the application from Kyma
  // and then remove the runtime from compass

  deleteAllK8sResources('/api/v1/namespaces/compass-system/secrets/compass-agent-configuration');
  deleteAllK8sResources('/apis/compass.kyma-project.io/v1alpha1/compassconnections/compass-connection');

  const runtimes = await queryRuntimesForScenario(client, scenarioName);
  for (const runtime of runtimes) {
    await unassignRuntimeFromScenario(client, runtime.id, scenarioName)
  }

  await removeScenarioFromCompass(client, scenarioName);
}

async function unregisterRuntimeFromCompass(client, scenarioName) {
  const runtimes = await queryRuntimesForScenario(client, scenarioName);
  for (const runtime of runtimes) {
    await client.unregisterRuntime(runtime.id);
  }
}

async function unregisterKymaFromCompass(client, scenarioName) {
  // Cleanup Compass
  const applications = await queryApplicationsForScenario(client, scenarioName);
  for (const application of applications) {
    await removeApplicationFromScenario(client, application.id, scenarioName);
    await client.unregisterApplication(application.id);
  }

  // TODO: refactor this step to cover runtime agent deleting the application from Kyma
  // and then remove the runtime from compass

  deleteAllK8sResources('/api/v1/namespaces/compass-system/secrets/compass-agent-configuration');
  deleteAllK8sResources('/apis/compass.kyma-project.io/v1alpha1/compassconnections/compass-connection');

  const runtimes = await queryRuntimesForScenario(client, scenarioName);
  for (const runtime of runtimes) {
      await client.unregisterRuntime(runtime.id);
  }
  await removeScenarioFromCompass(client, scenarioName);
}



module.exports = {
  registerKymaInCompass,
  unregisterKymaFromCompass,
  unregisterScenarioFromCompass,
  unregisterRuntimeFromCompass
};