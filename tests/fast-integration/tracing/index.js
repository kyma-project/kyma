const {
  waitForNamespace,
  getEnvOrDefault,
} = require('../utils');
const {
  waitForSubscriptionsTillReady,
  checkInClusterEventTracing,
} = require('../test/fixtures/commerce-mock');

function tracingTests(testNamespace) {
  if (getEnvOrDefault('KYMA_MAJOR_UPGRADE', 'false') === 'true') {
    console.log('Skipping tracing tests for Kyma 1 to Kyma 2 upgrade scenario');
    return;
  }

  describe('Tracing Tests:', function() {
    this.timeout(5 * 60 * 1000); // 5 min
    this.slow(5000);

    before('Ensure the test and mock namespaces exist', async function() {
      await waitForNamespace(testNamespace);
    });

    context('with Nats backend', function() {
      it('Wait until subscriptions are ready', async () => {
        await waitForSubscriptionsTillReady(testNamespace);
      });

      it('in-cluster structured event should have correct tracing spans', async () => {
        await checkInClusterEventTracing(testNamespace);
      });
    });
  });
}
module.exports = {
  tracingTests,
};
