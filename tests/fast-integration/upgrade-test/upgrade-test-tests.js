const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
  isDebugEnabled,
  printAllSubscriptions,
  printEventingControllerLogs,
  printEventingPublisherProxyLogs,
} = require('../utils');
const {loggingTests} = require('../logging');
const {
  monitoringTests,
  unexposeGrafana,
} = require('../monitoring');
const {tracingTests} = require('../tracing');
const {
  checkInClusterEventDelivery,
  checkFunctionResponse,
} = require('../test/fixtures/commerce-mock');
const {testNamespace} = require('../eventing-test/utils');

describe('Upgrade test tests', function() {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  let initialRestarts = null;
  const testNamespace = 'test';

  it('Listing all pods in cluster', async function() {
    initialRestarts = await getContainerRestartsForAllNamespaces();
  });

  it('in-cluster event should be delivered (legacy events, structured and binary cloud events)', async function() {
    await checkInClusterEventDelivery(testNamespace);
  });

  it('function should be reachable through secured API Rule', async function() {
    await checkFunctionResponse(testNamespace);
  });

  it('Should print report of restarted containers, skipped if no crashes happened', async function() {
    const afterTestRestarts = await getContainerRestartsForAllNamespaces();
    printRestartReport(initialRestarts, afterTestRestarts);
  });

  monitoringTests();
  loggingTests();
  tracingTests(testNamespace);

  // runs after each test in every block
  afterEach(async function() {
    // if the test is failed, then printing some debug logs
    if (this.currentTest.state === 'failed' && isDebugEnabled()) {
      await printAllSubscriptions(testNamespace);
      await printEventingControllerLogs();
      await printEventingPublisherProxyLogs();
    }
  });

  after('Unexpose Grafana', async () => {
    await unexposeGrafana();
  });
});
