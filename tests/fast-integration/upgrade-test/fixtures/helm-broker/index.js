const k8s = require('@kubernetes/client-node');
const fs = require('fs');
const path = require('path');

const {
  k8sApply,
} = require('../../../utils');

const sampleAddonsYaml = fs.readFileSync(
    path.join(__dirname, './sample-addons.yaml'),
    {
      encoding: 'utf8',
    },
);

const testingServiceInstanceYaml = fs.readFileSync(
    path.join(__dirname, './service-instance.yaml'),
    {
      encoding: 'utf8',
    },
);

const clusterAddonsCfgObj = k8s.loadYaml(sampleAddonsYaml);
const serviceInstanceObj = k8s.loadYaml(testingServiceInstanceYaml);

async function ensureHelmBrokerTestFixture(targetNamespace) {
  await k8sApply([clusterAddonsCfgObj]);
  // deprecated after sc migration
  // await waitForServiceClass('testing', targetNamespace);
  await k8sApply([serviceInstanceObj]);
  // deprecated after sc migration
  // await waitForServiceInstance('testing', targetNamespace);
}

async function checkServiceInstanceExistence(targetNamespace) {
  // deprecated after sc migration
  // await getServiceInstance('testing', targetNamespace);
  await waitForConfigMap('testing', targetNamespace);
}

module.exports = {
  ensureHelmBrokerTestFixture,
  checkServiceInstanceExistence,
};
