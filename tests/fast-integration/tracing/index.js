const {
  getEventingBackend,
  waitForNamespace,
  switchEventingBackend,
} = require('../utils');
const {
  testNamespace,
  backendK8sSecretName,
  backendK8sSecretNamespace,
  mockNamespace,
  natsBackend,
  isSKR,
} = require('./utils');
const {
  sendLegacyEventAndCheckTracing,
  sendCloudEventStructuredModeAndCheckTracing,
  sendCloudEventBinaryModeAndCheckTracing,
  waitForSubscriptionsTillReady,
} = require('../test/fixtures/commerce-mock');
const {prometheusPortForward} = require('../monitoring/client');
const {testPrep} = require('./test-prep');
const {testCleanup} = require('./test-cleanup');

async function tracingTests() {
  if (isSKR) {
    console.log('Skipping eventing tracing tests on SKR...');
    return;
  }
  await testPrep();
  describe('Tracing Tests:', function() {
    this.timeout(5 * 60 * 1000); // 5 min
    let cancelPrometheusPortForward = null;

    before('Ensure the test and mock namespaces exist', async function() {
      await waitForNamespace(testNamespace);
      await waitForNamespace(mockNamespace);
      cancelPrometheusPortForward = prometheusPortForward();
    });

    after(async function() {
      cancelPrometheusPortForward();
    });

    // Run Eventing tracing tests
    context('with Nats backend', function() {
      it('Switch Eventing Backend to Nats', async function() {
        const currentBackend = await getEventingBackend();
        if (currentBackend && currentBackend.toLowerCase() === natsBackend) {
          this.skip();
        }
        await switchEventingBackend(backendK8sSecretName, backendK8sSecretNamespace, natsBackend);
      });

      it('Wait until subscriptions are ready', async () => {
        await waitForSubscriptionsTillReady(testNamespace);
      });

      it('order.created.v1 event from CommerceMock should have correct tracing spans', async () => {
        await sendLegacyEventAndCheckTracing(testNamespace, mockNamespace);
      });

      it('order.created.v1 structured cloud from CommerceMock should have correct tracing spans', async function() {
        await sendCloudEventStructuredModeAndCheckTracing(testNamespace, mockNamespace);
      });

      it('order.created.v1 binary cloud event from CommerceMock should have correct tracing spans', async function() {
        await sendCloudEventBinaryModeAndCheckTracing(testNamespace, mockNamespace);
      });
    });
  });
  await testCleanup();
}
module.exports = {
  tracingTests,
};
