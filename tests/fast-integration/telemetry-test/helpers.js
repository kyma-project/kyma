const k8s = require('@kubernetes/client-node');
const fs = require('fs');
const path = require('path');
const {
  waitForK8sObject,
} = require('../utils');

function loadTestData(fileName) {
  return loadResourceFromFile(`./testdata/${fileName}`);
}

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

function waitForLogPipelineStatusRunning(name) {
  return waitForLogPipelineStatusCondition(name, 'Running', 180000);
}

module.exports = {
  loadTestData,
  waitForLogPipelineStatusRunning,
};
