module.exports = {
  loggingTests,
  ...require('./loki'),
  ...require('./client'),
};

const loki = require('./loki');
const {
  // k8sApply,
  k8sDelete,
  sleep,
  // waitForService,
  // waitForTracePipeline,
} = require('../utils');
// const {
//   restartProxyPod,
// } = require('../monitoring/grafana.js');
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

    // it('Should create the Istio Access Logs resource for Loki', async () => {
    //   await k8sApply(istioAccessLogsResource, namespace);
    //   await restartProxyPod();
    //   await waitForService('telemetry-trace-collector-internal', namespace);
    //   await waitForTracePipeline('jaeger');
    // });

    it('Should query Loki and verify format of Istio Access Logs', async () => {
      // Sleep for 10 seconds to wait for logs to come into the istio-proxy container
      await sleep(10*1000);
      await loki.verifyIstioAccessLogFormat(startTimestamp);
    });
  });
}

// function loadResourceFromFile(file) {
//   const yaml = fs.readFileSync(path.join(__dirname, file), {
//     encoding: 'utf8',
//   });
//   return k8s.loadAllYaml(yaml);
// }
