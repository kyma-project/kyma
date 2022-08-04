const loki = require('./loki');
const {k8sApply} = require('../utils');
const fs = require('fs');
const path = require('path');
const k8s = require('@kubernetes/client-node');

const istioAccessLogsResource = loadResourceFromFile('./istio_access_logs.yaml');
const namespace = 'kyma-system';

function loadResourceFromFile(file) {
  const yaml = fs.readFileSync(path.join(__dirname, file), {
    encoding: 'utf8',
  });
  return k8s.loadAllYaml(yaml);
}

function istioAccessLogsTests(startTimestamp) {
  describe('Istio Access Logs tests', function() {
    it('Should create Istio Access Logs resource for Loki', async () => {
      await k8sApply(istioAccessLogsResource, namespace);
    });

    it('Should query Loki and verify format of Istio access logs', async () => {
      await loki.verifyIstioAccessLogFormat(startTimestamp);
    });
  });
}

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

    istioAccessLogsTests(testStartTimestamp);
  });
}

module.exports = {
  loggingTests,
  ...require('./loki'),
  ...require('./client'),
};
