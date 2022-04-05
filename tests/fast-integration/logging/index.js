const logging = require('./helpers');
const {lokiPortForward} = require('./client');

function loggingTests() {
  const testStartTimestamp = new Date().toISOString();
  describe('Logging Tests:', function() {
    this.timeout(5 * 60 * 1000); // 5 min
    this.slow(5 * 1000);

    let cancelPortForward = null;

    before(async () => {
      cancelPortForward = lokiPortForward();
    });

    after(async () => {
      cancelPortForward();
    });

    it('Check Loki logs from kyma-system and kyma-integration namespaces', async () => {
      await logging.checkLokiLogsInKymaNamespaces(testStartTimestamp);
    });

    it('Check Loki logs from kyma-system and kyma-integration namespaces 2', async () => {
      await logging.checkKymaLogsInLoki(testStartTimestamp);
    });

    it('Retention Period and Max look-back Period should be 120h', async () => {
      await logging.checkRetentionPeriod();
    });

    it('Persistent Volume Claim Size should be 30Gi', async () => {
      await logging.checkPersistentVolumeClaimSize();
    });

    it('Loki should not be exposed through Virtual Service', async () => {
      await logging.checkVirtualServicePresence();
    });

    it('Loki should not be exposed through Virtual Service 2', async () =>{
      await logging.checkIfLokiVirutalServiceIsPresence();
    });
  });
}
module.exports = {
  loggingTests,
  ...require('./helpers'),
  ...require('./client'),
};
