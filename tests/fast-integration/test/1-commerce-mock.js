const axios = require('axios');
const https = require('https');
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  checkFunctionResponse,
  addService,
  updateService,
  deleteService,
  sendLegacyEventAndCheckResponse,
  sendCloudEventStructuredModeAndCheckResponse,
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
  checkCommerceMockLogsInLoki,
} = require('../logging');

function commerceMockTests(testNamespace) {
  describe('CommerceMock Tests:', function() {
    this.timeout(10 * 60 * 1000);
    this.slow(5000);
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

    it('in-cluster event should be delivered (structured and binary mode)', async function() {
      await checkInClusterEventDelivery(testNamespace);
    });

    it('function should be reachable through secured API Rule', async function() {
      await checkFunctionResponse(testNamespace);
    });

    it('order.created.v1 event should trigger the lastorder function', async function() {
      await sendLegacyEventAndCheckResponse();
    });

    it('order.created.v1 cloud event in structured mode should trigger the lastorder function', async function() {
      await sendCloudEventStructuredModeAndCheckResponse();
    });

    it('order.created.v1 cloud event in binary mode should trigger the lastorder function', async function() {
      await sendCloudEventBinaryModeAndCheckResponse();
    });

    it('should add, update and delete a service', async function() {
      const serviceId = await addService();
      await updateService(serviceId);
      await deleteService(serviceId);
    });

    it('Should print report of restarted containers, skipped if no crashes happened', async function() {
      const afterTestRestarts = await getContainerRestartsForAllNamespaces();
      printRestartReport(initialRestarts, afterTestRestarts);
    });

    it('Logs from commerce mock pod should be retrieved through Loki', async function() {
      await checkLokiLogs(testStartTimestamp);
    });

    it('Logs from commerce mock pod should be retrieved through Loki 2', async function() {
      await checkCommerceMockLogsInLoki(testStartTimestamp);
    });
  });
}

module.exports = {
  commerceMockTests,
};
