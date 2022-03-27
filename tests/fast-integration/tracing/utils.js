const {
  cleanMockTestFixture,
} = require('../test/fixtures/commerce-mock');

const {
  debug,
  deleteEventingBackendK8sSecret,
} = require('../utils');

const testNamespace = `test-tracing`;
const mockNamespace = process.env.MOCK_NAMESPACE || 'mocks';
const backendK8sSecretName = process.env.BACKEND_SECRET_NAME || 'tracing-backend';
const backendK8sSecretNamespace = process.env.BACKEND_SECRET_NAMESPACE || 'default';
const eventMeshSecretFilePath = process.env.EVENTMESH_SECRET_FILE || '';
const isSKR = process.env.KYMA_TYPE === 'SKR';
const natsBackend = 'nats';
const timeoutTime = 10 * 60 * 1000;
const slowTime = 5000;

// cleans up all the test resources including the compass scenario
async function cleanupTestingResources() {
  if (eventMeshSecretFilePath) {
    debug('Removing Event Mesh secret');
    await deleteEventingBackendK8sSecret(backendK8sSecretName, backendK8sSecretNamespace);
  }

  debug(`Removing ${testNamespace} and ${mockNamespace} namespaces`);
  await cleanMockTestFixture(mockNamespace, testNamespace, true);
}

module.exports = {
  testNamespace,
  mockNamespace,
  isSKR,
  backendK8sSecretName,
  backendK8sSecretNamespace,
  eventMeshSecretFilePath,
  timeoutTime,
  slowTime,
  cleanupTestingResources,
  natsBackend,
};

