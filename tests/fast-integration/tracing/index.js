const {waitForNamespace} = require('../utils');
const {
  sendLegacyEventAndCheckTracing,
  sendCloudEventStructuredModeAndCheckTracing,
  sendCloudEventBinaryModeAndCheckTracing,
  waitForSubscriptionsTillReady,
} = require('../test/fixtures/commerce-mock');

const testNamespace = 'test';
const mockNamespace = process.env.MOCK_NAMESPACE || 'mocks';
const isSKR = process.env.KYMA_TYPE === 'SKR';

async function tracingTests() {
  if (isSKR) {
    console.log('Skipping eventing tracing tests on SKR...');
    return;
  }
  describe('Tracing Tests:', function() {
    this.timeout(5 * 60 * 1000); // 5 min

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
