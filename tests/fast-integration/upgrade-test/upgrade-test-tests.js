const {
  checkInClusterEventDelivery,
  checkFunctionResponse,
  sendLegacyEventAndCheckResponse,
} = require('../test/fixtures/commerce-mock');
const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
} = require('../utils');
const {
  checkServiceInstanceExistence,
} = require('./fixtures/helm-broker');
const {
  loggingTests,
} = require('../logging');
const {
  monitoringTests,
} = require('../monitoring');
const {
  tracingTests,
} = require('../tracing');

describe('Upgrade test tests', function() {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  let initialRestarts = null;
  const testNamespace = 'test';

  it('Listing all pods in cluster', async function() {
    initialRestarts = await getContainerRestartsForAllNamespaces();
  });

  it('in-cluster event should be delivered', async function() {
    await checkInClusterEventDelivery(testNamespace);
  });

  it('function should be reachable through secured API Rule', async function() {
    await checkFunctionResponse(testNamespace);
  });

  it('order.created.v1 legacy event should trigger the lastorder function', async function() {
    await sendLegacyEventAndCheckResponse();
  });

  it('service instance provisioned by helm broker should be reachable', async function() {
    await checkServiceInstanceExistence(testNamespace);
  });

  it('Should print report of restarted containers, skipped if no crashes happened', async function() {
    const afterTestRestarts = await getContainerRestartsForAllNamespaces();
    printRestartReport(initialRestarts, afterTestRestarts);
  });

  monitoringTests();
  loggingTests();
  tracingTests();
});
