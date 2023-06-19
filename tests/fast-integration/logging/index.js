module.exports = {
  loggingTests,
  ...require('./loki'),
  ...require('./client'),
};

const loki = require('./loki');
const {
  k8sDelete,
} = require('../utils');
const {loadResourceFromFile} = require('./client');

function loggingTests() {
  const testStartTimestamp = new Date().toISOString();
  console.log('testStartTimestamp', testStartTimestamp);
  describe('Logging Tests:', function() {
    this.timeout(5 * 60 * 1000);
    this.slow(5000);

    it('Check Loki logs from kyma-system namespace', async () => {
      await loki.checkKymaLogs(testStartTimestamp);
    });

    it('Should exclude fluent-bit logs', async () => {
      await loki.checkFluentBitLogs(testStartTimestamp);
    });

    it('Retention Period and Max look-back Period should be 120h', async () => {
      await loki.checkRetentionPeriod();
    });

    it('Persistent Volume Claim Size should be 30Gi', async () => {
      await loki.checkPersistentVolumeClaimSize();
    });
    istioAccessLogsTests(testStartTimestamp);
  });
}

function istioAccessLogsTests(startTimestamp) {
  describe('Istio Access Logs tests', function() {
    const istioAccessLogsResource = loadResourceFromFile('./istio_access_logs.yaml');
    const namespace = 'kyma-system';

    after('Should delete the Istio Access Logs resource', async () => {
      await k8sDelete(istioAccessLogsResource, namespace);
    });

    it('Should query Loki and verify format of Istio Access Logs', async () => {
      await loki.verifyIstioAccessLogFormat(startTimestamp);
    });
  });
}

