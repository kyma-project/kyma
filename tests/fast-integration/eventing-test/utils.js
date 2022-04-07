const {
  cleanMockTestFixture,
  cleanCompassResourcesSKR,
} = require('../test/fixtures/commerce-mock');

const {
  debug,
  getEnvOrThrow,
  deleteEventingBackendK8sSecret,
  getShootNameFromK8sServerUrl,
} = require('../utils');

const {DirectorClient, DirectorConfig, getAlreadyAssignedScenarios} = require('../compass');
const {GardenerClient, GardenerConfig} = require('../gardener');
const fs = require('fs');
const isSKR = process.env.KYMA_TYPE === 'SKR';
const skipResourceCleanup = process.env.SKIP_CLEANUP || false;
const suffix = getSuffix(isSKR);
const appName = `app-${suffix}`;
const scenarioName = `test-${suffix}`;
const testNamespace = `test-${suffix}`;
const mockNamespace = process.env.MOCK_NAMESPACE || 'mocks';
const backendK8sSecretName = process.env.BACKEND_SECRET_NAME || 'eventing-backend';
const backendK8sSecretNamespace = process.env.BACKEND_SECRET_NAMESPACE || 'default';
const eventMeshSecretFilePath = process.env.EVENTMESH_SECRET_FILE || '';
const DEBUG_MODE = process.env.DEBUG;
const timeoutTime = 10 * 60 * 1000;
const slowTime = 5000;
const natsBackend = 'nats';
const bebBackend = 'beb';
const eventMeshNamespace = getEventMeshNamespace();

// SKR related constants
let gardener = null;
let director = null;
let shootName = null;
if (isSKR) {
  gardener = new GardenerClient(GardenerConfig.fromEnv()); // create gardener client
  director = new DirectorClient(DirectorConfig.fromEnv()); // director client for Compass
  shootName = getShootNameFromK8sServerUrl();
}

// reads the EventMesh namespace from the credentials file
function getEventMeshNamespace() {
  try {
    const eventMeshSecret = JSON.parse(fs.readFileSync(eventMeshSecretFilePath, {encoding: 'utf8'}));
    return '/' + eventMeshSecret['namespace'];
  } catch (e) {
    return undefined;
  }
}

// cleans up all the test resources including the compass scenario
async function cleanupTestingResources() {
  if (isSKR) {
    debug('Cleaning compass resources');
    // Get shoot info from gardener to get compassID for this shoot
    const skrInfo = await gardener.getShoot(shootName);
    await cleanCompassResourcesSKR(director, appName, scenarioName, skrInfo.compassID);
  }

  // skip the cluster resources cleanup if the SKIP_CLEANUP env flag is set
  if (skipResourceCleanup === 'true') {
    return;
  }

  // Delete eventing backend secret if it was created by test
  if (eventMeshSecretFilePath) {
    debug('Removing Event Mesh secret');
    await deleteEventingBackendK8sSecret(backendK8sSecretName, backendK8sSecretNamespace);
  }

  debug(`Removing ${testNamespace} and ${mockNamespace} namespaces`);
  await cleanMockTestFixture(mockNamespace, testNamespace, true);
}

// gets the suffix depending on kyma type
function getSuffix(isSKR) {
  let suffix;
  if (isSKR) {
    suffix = getEnvOrThrow('TEST_SUFFIX');
  } else {
    suffix = 'evnt';
  }
  return suffix;
}

// getRegisteredCompassScenarios lists the registered compass scenarios
async function getRegisteredCompassScenarios() {
  try {
    const skrInfo = await gardener.getShoot(shootName);
    const result = await getAlreadyAssignedScenarios(director, skrInfo.compassID);
    console.log('List of the active scenarios:');
    result.map((scenario, i) => console.log('%s)%s', i+1, scenario));
  } catch (e) {
    console.log('Cannot display the assigned scenarios');
  }
}

module.exports = {
  appName,
  scenarioName,
  testNamespace,
  mockNamespace,
  isSKR,
  backendK8sSecretName,
  backendK8sSecretNamespace,
  eventMeshSecretFilePath,
  DEBUG_MODE,
  timeoutTime,
  slowTime,
  director,
  gardener,
  shootName,
  suffix,
  cleanupTestingResources,
  natsBackend,
  bebBackend,
  eventMeshNamespace,
  getRegisteredCompassScenarios,
};
