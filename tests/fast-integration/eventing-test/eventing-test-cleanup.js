const axios = require('axios');
const https = require('https');
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  appName,
  scenarioName,
  testNamespace,
  mockNamespace,
  isSKR,
  backendK8sSecretName,
  backendK8sSecretNamespace,
  eventMeshSecretFilePath,
  timeoutTime,
  slowTime,
} = require('./utils');
const {
  cleanMockTestFixture,
  cleanCompassResourcesSKR,
} = require('../test/fixtures/commerce-mock');
const {
  debug,
  deleteEventingBackendK8sSecret,
} = require('../utils');

describe('Eventing tests cleanup', function() {
  this.timeout(timeoutTime);
  this.slow(slowTime);
  const director = null;
  const skrInfo = null;


  it('Cleaning: Test namespaces should be deleted', async function() {
    // Unregister SKR resources from Compass
    if (isSKR) {
      debug('Cleaning SKR...');
      await cleanCompassResourcesSKR(director, appName, scenarioName, skrInfo.compassID);
    }

    // Delete eventing backend secret if it was created by test
    if (eventMeshSecretFilePath) {
      debug('Removing Event Mesh secret');
      await deleteEventingBackendK8sSecret(backendK8sSecretName, backendK8sSecretNamespace);
    }

    debug('Cleaning test resources');
    await cleanMockTestFixture(mockNamespace, testNamespace, true);
  });
});
