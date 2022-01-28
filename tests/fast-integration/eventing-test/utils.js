const {
  cleanMockTestFixture,
  cleanCompassResourcesSKR,
} = require('../test/fixtures/commerce-mock');

const {
  debug,
  deleteEventingBackendK8sSecret, getShootNameFromK8sServerUrl,
} = require('../utils');

const {DirectorClient, DirectorConfig} = require('../compass');
const {GardenerClient, GardenerConfig} = require('../gardener');
const suffix = 'evnt';
const appName = `app-${suffix}`;
const scenarioName = `test-${suffix}`;
const testNamespace = `test-${suffix}`;
const mockNamespace = process.env.MOCK_NAMESPACE || 'mocks';
const isSKR = process.env.KYMA_TYPE === 'SKR';
const backendK8sSecretName = process.env.BACKEND_SECRET_NAME || 'eventing-backend';
const backendK8sSecretNamespace = process.env.BACKEND_SECRET_NAMESPACE || 'default';
const eventMeshSecretFilePath = process.env.EVENTMESH_SECRET_FILE || '';
const DEBUG_MODE = process.env.DEBUG;
const timeoutTime = 10 * 60 * 1000;
const slowTime = 5000;
const fatalErrCode = 'ERR_MOCHA_FATAL';

async function cleanupTestingResources() {
  if (isSKR) {
    debug('Cleaning SKR...');

    // director client for Compass
    const director = new DirectorClient(DirectorConfig.fromEnv());

    // create gardener clients
    const gardener = new GardenerClient(GardenerConfig.fromEnv());

    // Get shoot info from gardener to get compassID for this shoot
    const shootName = getShootNameFromK8sServerUrl();
    const skrInfo = await gardener.getShoot(shootName);
    await cleanCompassResourcesSKR(director, appName, scenarioName, skrInfo.compassID);
  }

  // Delete eventing backend secret if it was created by test
  if (eventMeshSecretFilePath) {
    debug('Removing Event Mesh secret');
    await deleteEventingBackendK8sSecret(backendK8sSecretName, backendK8sSecretNamespace);
  }

  debug('Cleaning test resources');
  await cleanMockTestFixture(mockNamespace, testNamespace, true);
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
  fatalErrCode,
  cleanupTestingResources,
};
