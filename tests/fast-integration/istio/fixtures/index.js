const k8s = require('@kubernetes/client-node');
const fs = require('fs');
const path = require('path');

const istioTestNamespaceYaml = fs.readFileSync(
    path.join(__dirname, './ns.yaml'),
    {
      encoding: 'utf8',
    },
);

const httpbinDeploymentYaml = fs.readFileSync(
    path.join(__dirname, './httpbin.yaml'),
    {
      encoding: 'utf8',
    },
);

const apiRuleYaml = fs.readFileSync(
    path.join(__dirname, './apirule.yaml'),
    {
      encoding: 'utf8',
    },
);

const {config, expect} = require('chai');
config.truncateThreshold = 0; // more verbose errors

const {
  waitForDeployment,
  waitForK8sObject,
  waitForVirtualService,
  k8sApply,
  deleteAllK8sResources,
  deleteNamespaces,
  convertAxiosError,
} = require('../../utils');

const https = require('https');
const axios = require('axios').default;

const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;


const httpbinService = 'httpbin-api-rule';
const httpbinNamespace = 'istio-connectivity-test';

const istioTestNamespaceObj = k8s.loadYaml(istioTestNamespaceYaml);
const httpbinDeploymentObj = k8s.loadAllYaml(
    httpbinDeploymentYaml,
);
const apiRuleObj = k8s.loadYaml(apiRuleYaml);

async function checkHttpbinResponse() {
  const vs = await waitForVirtualService(httpbinNamespace, httpbinService);
  const host = vs.spec.hosts[0];

  const response = await axios.get(`https://${host}/ip`).catch((err)=>{
    throw convertAxiosError(err, 'Httpbin call responded with error');
  });
  expect(response.status).to.be.equal(200);
}

async function ensureIstioConnectivityFixture() {
  await k8sApply([istioTestNamespaceObj]);
  await k8sApply(httpbinDeploymentObj, httpbinNamespace);
  await k8sApply([apiRuleObj]);
  await waitForDeployment('httpbin', httpbinNamespace);
  const apiRulePath = `/apis/gateway.kyma-project.io/v1alpha1/namespaces/${httpbinNamespace}/apirules`;
  await waitForK8sObject(apiRulePath, {}, (_type, _apiObj, watchObj) => {
    return (watchObj.object.metadata.name == httpbinService && watchObj.object.status.APIRuleStatus.code == 'OK');
  }, 60 * 1000, 'Waiting for APIRule to be ready timeout');
  await waitForVirtualService(httpbinNamespace, httpbinService);
}

function getResourcePaths(namespace) {
  return [
    `/apis/gateway.kyma-project.io/v1alpha1/namespaces/${namespace}/apirules`,
    `/apis/apps/v1/namespaces/${namespace}/deployments`,
    `/api/v1/namespaces/${namespace}/services`,
  ];
}

function cleanIstioConnectivityFixture(wait = true) {
  for (const path of getResourcePaths(httpbinNamespace)) {
    deleteAllK8sResources(path);
  }
  return deleteNamespaces([httpbinNamespace], wait);
}

module.exports = {
  ensureIstioConnectivityFixture,
  checkHttpbinResponse,
  cleanIstioConnectivityFixture,
};
