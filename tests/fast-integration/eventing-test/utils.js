const suffix = 'eventing';
const appName = `app-${suffix}`;
const scenarioName = `test-${suffix}`;
const testNamespace = 'test';
const mockNamespace = process.env.MOCK_NAMESPACE || 'mocks'
const isSKR = process.env.KYMA_TYPE === "SKR";
const backendK8sSecretName = process.env.BACKEND_SECRET_NAME || "eventing-backend";
const backendK8sSecretNamespace = process.env.BACKEND_SECRET_NAMESPACE || "default";
const eventMeshSecretFilePath = process.env.EVENTMESH_SECRET_FILE || "";
const skrInstanceId = process.env.SKR_INSTANCE_ID || "";
const DEBUG_MODE = process.env.DEBUG;
const timeoutTime = 10 * 60 * 1000;
const slowTime = 5000;
const eventingScenarioName = 'kyma-eventing-e2e-tests';

module.exports = {
    appName,
    scenarioName,
    testNamespace,
    mockNamespace,
    isSKR,
    backendK8sSecretName,
    backendK8sSecretNamespace,
    eventMeshSecretFilePath,
    skrInstanceId,
    DEBUG_MODE,
    timeoutTime,
    slowTime,
    eventingScenarioName
}