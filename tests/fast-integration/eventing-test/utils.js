const suffix = 'eventing';
const appName = `app-${suffix}`;
const scenarioName = `test-${suffix}`;
const testNamespace = 'test';
const mockNamespace = process.env.MOCK_NAMESPACE || 'mocks';
const isSKR = process.env.KYMA_TYPE === 'SKR';
const backendK8sSecretName = process.env.BACKEND_SECRET_NAME || 'eventing-backend';
const backendK8sSecretNamespace = process.env.BACKEND_SECRET_NAMESPACE || 'default';
const eventMeshSecretFilePath = process.env.EVENTMESH_SECRET_FILE || '';
const DEBUG_MODE = process.env.DEBUG;
const timeoutTime = 10 * 60 * 1000;
const slowTime = 5000;

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
};
