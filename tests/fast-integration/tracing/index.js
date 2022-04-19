const {
  waitForNamespace,
  getEnvOrDefault,
} = require('../utils');
const {
  sendLegacyEventAndCheckTracing,
  sendCloudEventStructuredModeAndCheckTracing,
  sendCloudEventBinaryModeAndCheckTracing,
  waitForSubscriptionsTillReady,
} = require('../test/fixtures/commerce-mock');

function tracingTests(mockNamespace, testNamespace) {
  if (getEnvOrDefault('KYMA_MAJOR_UPGRADE', 'false') === 'true') {
    console.log('Skipping tracing tests for Kyma 1 to Kyma 2 upgrade scenario');
    return;
  }

  describe('Tracing Tests:', function() {
    this.timeout(5 * 60 * 1000); // 5 min
    this.slow(5000);

    before('Ensure the test and mock namespaces exist', async function() {
      await waitForNamespace(testNamespace);
      await waitForNamespace(mockNamespace);
    });

    context('with Nats backend', function() {
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
}
module.exports = {
  tracingTests,
};
