module.exports = {
  loadTestData,
  patchSecret,
  waitForLogPipelineStatusRunning,
  waitForTracePipeline,
  waitForPodWithLabel,
};

const k8s = require('@kubernetes/client-node');
const fs = require('fs');
const path = require('path');
const {
  k8sCoreV1Api,
  waitForK8sObject,
} = require('../utils');

function loadTestData(fileName) {
  return loadResourceFromFile(`./testdata/${fileName}`);
}

function waitForLogPipelineStatusRunning(name) {
  return waitForLogPipelineStatusCondition(name, 'Running', 180000);
}

function loadResourceFromFile(file) {
  const yaml = fs.readFileSync(path.join(__dirname, file), {
    encoding: 'utf8',
  });
  return k8s.loadAllYaml(yaml);
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

function checkLastCondition(logPipeline, conditionType) {
  const conditions = logPipeline.status.conditions;
  if (conditions.length === 0) {
    return false;
  }
  const lastCondition = conditions[conditions.length - 1];
  return lastCondition.type === conditionType;
}

function waitForTracePipeline(name) {
  return waitForK8sObject(
      '/apis/telemetry.kyma-project.io/v1alpha1/tracepipelines',
      {},
      (_type, watchObj, _) => {
        return (watchObj.metadata.name === name);
      },
      18000,
      `Waiting for trace pipeline ${name} timeout 18000 ms)`,
  );
}

function waitForPodWithLabel(
    labelKey,
    labelValue,
    namespace = 'default',
    timeout = 90000,
) {
  const query = {
    labelSelector: `${labelKey}=${labelValue}`,
  };
  return waitForK8sObject(
      `/api/v1/namespaces/${namespace}/pods`,
      query,
      (_type, _apiObj, watchObj) => {
        return (
          watchObj.object.status.phase === 'Running' &&
            watchObj.object.status.containerStatuses.every((cs) => cs.ready)
        );
      },
      timeout,
      `Waiting for pod with label ${labelKey}=${labelValue} timeout (${timeout} ms)`,
  );
}

async function patchSecret(secretName, namespace, patch) {
  const options = {'headers': {'Content-type': k8s.PatchUtils.PATCH_FORMAT_JSON_PATCH}};

  await k8sCoreV1Api.patchNamespacedSecret(
      secretName,
      namespace,
      patch,
      undefined,
      undefined,
      undefined,
      undefined,
      options,
  );
}
