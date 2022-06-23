const loki = require('./loki');

function loggingTests() {
  const testStartTimestamp = new Date().toISOString();
  describe('Logging Tests:', function() {
    this.timeout(5 * 60 * 1000); // 5 min
    this.slow(5000);

    it('Check Loki logs from kyma-system namespace', async () => {
      await loki.checkKymaLogs(testStartTimestamp);
    });

    it('Retention Period and Max look-back Period should be 120h', async () => {
      await loki.checkRetentionPeriod();
    });

    it('Persistent Volume Claim Size should be 30Gi', async () => {
      await loki.checkPersistentVolumeClaimSize();
    });
  });
}
module.exports = {
  loggingTests,
  ...require('./loki'),
  ...require('./client'),
};
