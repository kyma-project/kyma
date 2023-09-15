const axios = require('axios');
const https = require('https');
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  checkFunctionResponse,
  checkInClusterEventDelivery,
} = require('./fixtures/commerce-mock');
const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
} = require('../utils');
const loki = require('../logging');

function commerceMockTests(testNamespace) {
  describe('CommerceMock Tests:', function() {
    this.timeout(10 * 60 * 1000);
    this.slow(5000);
    const testStartTimestamp = new Date().toISOString();
    let initialRestarts = null;

    it('Listing all pods in cluster', async function() {
      initialRestarts = await getContainerRestartsForAllNamespaces();
    });

    it('in-cluster event should be delivered (structured and binary mode)', async function() {
      await checkInClusterEventDelivery(testNamespace);
    });

    it('function should be reachable through API Rule', async function() {
      await checkFunctionResponse(testNamespace);
    });

    it('Should print report of restarted containers, skipped if no crashes happened', async function() {
      const afterTestRestarts = await getContainerRestartsForAllNamespaces();
      printRestartReport(initialRestarts, afterTestRestarts);
    });

    it('Logs from commerce mock pod should be retrieved through Loki', async function() {
      await loki.checkCommerceMockLogs(testStartTimestamp);
    });
  });
}

module.exports = {
  commerceMockTests,
};
