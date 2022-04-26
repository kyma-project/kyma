const axios = require('axios');
const https = require('https');
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  ensureCommerceMockLocalTestFixture,
  checkFunctionResponse,
  addService,
  updateService,
  deleteService,
  sendCloudEventStructuredModeAndCheckResponse,
  cleanMockTestFixture,
  checkInClusterEventDelivery,
  sendCloudEventBinaryModeAndCheckResponse,
} = require('./fixtures/commerce-mock');
const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
} = require('../utils');
const {
  checkLokiLogs,
  lokiPortForward,
} = require('../logging');

function commerceMockTests() {
  describe('CommerceMocko Tests:', function() {
    this.timeout(10 * 60 * 1000);
    this.slow(5000);
    const withCentralAppConnectivity = (process.env.WITH_CENTRAL_APP_CONNECTIVITY === 'true');
    const testNamespace = 'test';
    const testStartTimestamp = new Date().toISOString();
    let initialRestarts = null;
    let cancelPortForward = null;

    before(() => {
      cancelPortForward = lokiPortForward();
    });

    after(() => {
      cancelPortForward();
    });

    it('Listing all pods in cluster', async function() {
      initialRestarts = await getContainerRestartsForAllNamespaces();
    });

    it('CommerceMock test fixture should be ready', async function() {
      await ensureCommerceMockLocalTestFixture('mocks', testNamespace, withCentralAppConnectivity).catch((err) => {
        console.dir(err); // first error is logged
        return ensureCommerceMockLocalTestFixture('mocks', testNamespace, withCentralAppConnectivity);
      });
    });

    it('function should be reachable through secured API Rule', async function() {
      await checkFunctionResponse(testNamespace);
    });

    // this test fails. We need to find out why
    it('should add, update and delete a service', async function() {
      const serviceId = await addService();
      await updateService(serviceId);
      await deleteService(serviceId);
    });
    //
    it('Should print report of restarted containers, skipped if no crashes happened', async function() {
      const afterTestRestarts = await getContainerRestartsForAllNamespaces();
      printRestartReport(initialRestarts, afterTestRestarts);
    });

    it('Logs from commerce mock pod should be retrieved through Loki', async function() {
      await checkLokiLogs(testStartTimestamp);
    });

    // it('Test namespaces should be deleted', async function() {
    //   await cleanMockTestFixture('mocks', testNamespace, true);
    // });
  });
}

module.exports = {
  commerceMockTests,
};
