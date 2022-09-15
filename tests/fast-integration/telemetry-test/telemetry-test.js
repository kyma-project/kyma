const k8s = require('@kubernetes/client-node');
const {assert, expect} = require('chai');
const fs = require('fs');
const path = require('path');
const {
  k8sCoreV1Api,
  k8sApply,
  k8sDelete,
  waitForK8sObject,
} = require('../utils');
const {logsPresentInLoki, queryLoki} = require('../logging');
const {
  exposeGrafana,
  unexposeGrafana,
} = require('../monitoring');

const telemetryNamespace = 'kyma-system';
const defaultNamespace = 'default';
const mockserverNamespace = 'mockserver';
const testStartTimestamp = new Date().toISOString();

// Load Deployments
const regexFilterDeployment = loadResourceFromFile('./resources/deployments/regex-filter-deployment.yaml');
const mockserverDeployment = loadResourceFromFile('./resources/deployments/mockserver.yaml');
const spammerWorkloadPod = loadResourceFromFile('./resources/deployments/logs-workload.yaml');

// Load Telemetry CR's
const httpLogPipelineCR = loadResourceFromFile(
    './resources/telemetry-custom-resources/http-logpipeline.yaml');
const unknownPluginLogPipelineCR = loadResourceFromFile(
    './resources/telemetry-custom-resources/unknown-plugin-logpipeline.yaml');
const dropLabelsLogPipelineCR = loadResourceFromFile(
    './resources/telemetry-custom-resources/loki-metadata-filter-drop-labels-logpipeline.yaml');
const keepLabelsLogPipelineCR = loadResourceFromFile(
    './resources/telemetry-custom-resources/loki-metadata-filter-keep-labels-logpipeline.yaml');
const kubernetesCustomFilterLogPipelineCR = loadResourceFromFile(
    './resources/telemetry-custom-resources/kubernetes-custom-filter-logpipeline.yaml');
const excludeIstioProxyLogPipelineCR = loadResourceFromFile(
    './resources/telemetry-custom-resources/loki-exclude-istio-proxy-logpipeline.yaml');
const regexParser = loadResourceFromFile(
    './resources/telemetry-custom-resources/regex-logparser.yaml');

// CR names
const httpLogPipelineName = 'http-mockserver';
const dropLabelsLogPipelineName = 'loki-keep-annotations-drop-labels';
const keepLabelsLogPipelineName = 'loki-drop-annotations-keep-labels';
const excludeIstioProxyLogPipelineName = 'exclude-istio-proxy';

function loadResourceFromFile(file) {
  const yaml = fs.readFileSync(path.join(__dirname, file), {
    encoding: 'utf8',
  });
  return k8s.loadAllYaml(yaml);
}

function checkLastCondition(logPipeline, conditionType) {
  const conditions = logPipeline.status.conditions;
  if (conditions.length === 0) {
    return false;
  }
  const lastCondition = conditions[conditions.length - 1];
  return lastCondition.type === conditionType;
}

function waitForLogPipelineStatusCondition(name, lastConditionType, timeout) {
  return waitForK8sObject(
      '/apis/telemetry.kyma-project.io/v1alpha1/logpipelines',
      {},
      (_type, watchObj, _) => {
        return (
          watchObj.metadata.name === name && checkLastCondition(watchObj, lastConditionType)
        );
      },
      timeout,
      `Waiting for log pipeline ${name} timeout (${timeout} ms)`,
  );
}

async function prepareEnvironment() {
  await k8sApply(regexParser, telemetryNamespace);
  await k8sApply(mockserverDeployment, mockserverNamespace);
  await k8sApply(regexFilterDeployment, defaultNamespace);
  await k8sApply(spammerWorkloadPod, defaultNamespace);
}

async function cleanEnvironment() {
  await k8sDelete(regexParser, telemetryNamespace);
  await k8sDelete(mockserverDeployment, mockserverNamespace);
  await k8sDelete(regexFilterDeployment, defaultNamespace);
  await k8sDelete(spammerWorkloadPod, defaultNamespace);
}

describe('Telemetry Operator tests', function() {
  before('Prepare environment, expose Grafana', async () => {
    await prepareEnvironment();
    await exposeGrafana();
  });

  after('Clean environment, unexpose Grafana', async () => {
    await cleanEnvironment();
    await unexposeGrafana();
  });

  it('Operator should be ready', async () => {
    const res = await k8sCoreV1Api.listNamespacedPod(
        telemetryNamespace,
        'true',
        undefined,
        undefined,
        undefined,
        'control-plane=telemetry-operator',
    );
    const podList = res.body.items;
    assert.equal(podList.length, 1);
  });

  it('Loki LogPipeline should have Running condition', async () => {
    await waitForLogPipelineStatusCondition('loki', 'Running', 180000);
  });

  it('Webhook should reject a LogPipeline with unknown plugin', async () => {
    try {
      await k8sApply(unknownPluginLogPipelineCR, telemetryNamespace);
      await k8sDelete(unknownPluginLogPipelineCR, telemetryNamespace);
      assert.fail('Should not be able to apply LogPipeline with unknown plugin');
    } catch (e) {
      assert.equal(e.statusCode, 403);
      expect(e.body.message).to.have.string('denied the request');
      const errMsg = 'section \'abc\' tried to instance a plugin name that don\'t exists';
      expect(e.body.message).to.have.string(errMsg);
    }
  });

  it('Webhook should reject a LogPipeline with denied custom filter', async () => {
    try {
      await k8sApply(kubernetesCustomFilterLogPipelineCR, telemetryNamespace);
      await k8sDelete(kubernetesCustomFilterLogPipelineCR, telemetryNamespace);
      assert.fail('Should not be able to apply LogPipeline with kubernetes custom filter');
    } catch (e) {
      assert.equal(e.statusCode, 403);
      expect(e.body.message).to.have.string('denied the request');
      const errMsg = 'plugin \'kubernetes\' is forbidden';
      expect(e.body.message).to.have.string(errMsg);
    }
  });

  it('Should push the logs from kyma-system namespace to the default Loki output', async () => {
    const labels = '{namespace="kyma-system", job="telemetry-fluent-bit"}';
    const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
    assert.isTrue(logsPresent, 'No logs present in Loki with namespace="kyma-system"');
  });

  it('Should parse the logs using regex', async () => {
    try {
      const labels = '{namespace="default"}|json|pass="bar"|user="foo"';
      const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
      assert.isTrue(logsPresent, 'No parsed logs present in Loki');
    } catch (e) {
      assert.fail(e);
    }
  });

  context('Should verify HTTP LogPipeline', async () => {
    it(`Should create HTTP LogPipeline '${httpLogPipelineName}'`, async () => {
      await k8sApply(httpLogPipelineCR, telemetryNamespace);
      await waitForLogPipelineStatusCondition(httpLogPipelineName, 'Running', 180000);
    });

    it('Should push logs to the HTTP mockserver', async () => {
    // The mockserver prints received logs to stdout, which should finally be pushed to Loki by the other pipeline
      const labels = '{namespace="mockserver"}';
      const logsPresent = await logsPresentInLoki(labels, testStartTimestamp);
      assert.isTrue(logsPresent, 'No logs received by mockserver present in Loki');
    });

    it(`Should delete HTTP LogPipeline '${httpLogPipelineName}'`, async () => {
      await k8sDelete(httpLogPipelineCR, telemetryNamespace);
    });
  });

  context('Should verify Kubernetes metadata scenario 1: drop annotations, keep labels', async () => {
    it(`Should create Loki LogPipeline '${keepLabelsLogPipelineName}'`, async () =>{
      await k8sApply(keepLabelsLogPipelineCR, telemetryNamespace);
      await waitForLogPipelineStatusCondition(keepLabelsLogPipelineName, 'Running', 180000);
    });

    it(`Should verify that only labels are pushed to Loki`, async () =>{
      const labels = '{namespace="kyma-system", job="drop-annotations-keep-labels-telemetry-fluent-bit"}';
      const responseBody = await queryLoki(labels, testStartTimestamp);

      assert.isTrue(responseBody.data.result.length > 0, `No logs present in Loki for labels: ${labels}`);
      const entry = JSON.parse(responseBody.data.result[0].values[0][1]);
      assert.isTrue('kubernetes' in entry, `No kubernetes metadata present in log entry: ${entry} `);

      expect(entry['kubernetes']).not.to.have.property('annotations');
      expect(entry['kubernetes']).to.have.property('labels');
    });

    it(`Should delete Loki LogPipeline '${keepLabelsLogPipelineName}'`, async () =>{
      await k8sDelete(keepLabelsLogPipelineCR, telemetryNamespace);
    });
  });

  context('Should verify Kubernetes metadata scenario 2: keep annotations, drop labels', async () => {
    it(`Should create Loki LogPipeline '${dropLabelsLogPipelineName}'`, async () =>{
      await k8sApply(dropLabelsLogPipelineCR, telemetryNamespace);
      await waitForLogPipelineStatusCondition(dropLabelsLogPipelineName, 'Running', 180000);
    });

    it(`Should verify that only annotations are pushed to Loki`, async () =>{
      const labels = '{namespace="kyma-system", job="keep-annotations-drop-labels-telemetry-fluent-bit"}';
      const responseBody = await queryLoki(labels, testStartTimestamp);

      assert.isTrue(responseBody.data.result.length > 0, `No logs present in Loki for labels: ${labels}`);
      const entry = JSON.parse(responseBody.data.result[0].values[0][1]);
      assert.isTrue('kubernetes' in entry, `No kubernetes metadata present in log entry: ${entry} `);

      expect(entry['kubernetes']).not.to.have.property('labels');
      expect(entry['kubernetes']).to.have.property('annotations');
    });

    it(`Should delete Loki LogPipeline '${dropLabelsLogPipelineName}'`, async () =>{
      await k8sDelete(dropLabelsLogPipelineCR, telemetryNamespace);
    });
  });

  context('Should verify istio-proxy container and system logs are excluded', async () => {
    it(`Should create Loki LogPipeline '${keepLabelsLogPipelineName}'`, async () =>{
      await k8sApply(excludeIstioProxyLogPipelineCR, telemetryNamespace);
      await waitForLogPipelineStatusCondition(excludeIstioProxyLogPipelineName, 'Running', 180000);
    });

    it(`Should verify no system logs are pushed to Loki`, async () =>{
      const labels = '{namespace="kyma-system", job="exclude-istio-proxy-telemetry-fluent-bit"}';
      const logsFound = await logsPresentInLoki(labels, testStartTimestamp, 3);
      assert.isFalse(logsFound, `No logs must present in Loki for labels: ${labels}`);
    });

    it(`Should verify no istio-proxy logs are pushed to Loki`, async () =>{
      const labels = '{container="istio-proxy", job="exclude-istio-proxy-telemetry-fluent-bit"}';
      const logsFound = await logsPresentInLoki(labels, testStartTimestamp, 3);
      assert.isFalse(logsFound, `No logs must present in Loki for labels: ${labels}`);
    });

    it(`Should delete Loki LogPipeline '${keepLabelsLogPipelineName}'`, async () =>{
      await k8sDelete(excludeIstioProxyLogPipelineCR, telemetryNamespace);
    });
  });
});
