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

const apiRuleAllowYaml = fs.readFileSync(
    path.join(__dirname, './apirule-allow.yaml'),
    {
      encoding: 'utf8',
    },
);
const apiRuleOAuthYaml = fs.readFileSync(
    path.join(__dirname, './apirule-oauth.yaml'),
    {
      encoding: 'utf8',
    },
);

const {config, expect} = require('chai');
config.truncateThreshold = 0; // more verbose errors

const {
  getSecretData,
  retryPromise,
  waitForDeployment,
  waitForK8sObject,
  waitForVirtualService,
  k8sApply,
  deleteAllK8sResources,
  deleteNamespaces,
  convertAxiosError,
} = require('../../utils');

const {
  OAuthToken,
  OAuthCredentials,
} = require('../../lib/oauth');

const https = require('https');
const axios = require('axios').default;

const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;


const httpbinAllowService = 'httpbin-allow';
const httpbinOAuthService = 'httpbin-oauth2';
const httpbinNamespace = 'istio-connectivity-test';

const istioTestNamespaceObj = k8s.loadYaml(istioTestNamespaceYaml);
const httpbinDeploymentObj = k8s.loadAllYaml(
    httpbinDeploymentYaml,
);
const apiRuleAllowObj = k8s.loadYaml(apiRuleAllowYaml);
const apiRuleOAuthObj = k8s.loadAllYaml(
    apiRuleOAuthYaml,
);

async function checkHttpbinAllowResponse() {
  const vs = await waitForVirtualService(httpbinNamespace, httpbinAllowService);
  const host = vs.spec.hosts[0];
  const res = await retryPromise(
      () => getHttpbin(host),
      30,
      2000,
  ).catch((err) => {
    throw convertAxiosError(err, 'Httpbin call responded with error');
  });

  expect(res.status).to.be.equal(200);
}

async function checkHttpbinOAuthResponse() {
  const vs = await waitForVirtualService(httpbinNamespace, httpbinOAuthService);
  const host = vs.spec.hosts[0];
  const domain = host.split('.').slice(1).join('.');
  const accessToken = await getOAuthToken(domain);

  // expect error when unauthorized
  let errorOccurred = false;
  try {
    res = await axios.get(`https://${host}/headers`);
  } catch (err) {
    errorOccurred = true;
    expect(err.response.status).to.be.equal(401);
  }
  expect(errorOccurred).to.be.equal(true);

  let res = await retryPromise(
      () => getHttpbinOauth2(host, accessToken),
      30,
      2000,
  ).catch((err) => {
    throw convertAxiosError(err, 'Httpbin Oauth2 call responded with error');
  });
  expect(res.status).to.be.equal(200);
}

async function getHttpbin(host) {
  return axios.get(`https://${host}/headers`).catch((err) => {
    throw convertAxiosError(err, 'Httpbin call responded with error');
  });
}

async function getHttpbinOauth2(host, accessToken) {
  return await axios.get(`https://${host}/headers`, {
    headers: {
      Authorization: `bearer ${accessToken}`,
    },
  }).catch((err) => {
    throw convertAxiosError(err, 'Httpbin call responded with error');
  });
}

async function getOAuthToken(domain) {
  const oAuthSecretData = await getSecretData('httpbin-client', httpbinNamespace);
  const oAuthTokenGetter = new OAuthToken(
      `https://oauth2.${domain}/oauth2/token`,
      new OAuthCredentials(oAuthSecretData['client_id'], oAuthSecretData['client_secret']),
  );
  const accessToken = oAuthTokenGetter.getToken(['read', 'write']);

  return accessToken;
}

async function ensureIstioConnectivityFixture() {
  await k8sApply([istioTestNamespaceObj]);
  await k8sApply(httpbinDeploymentObj, httpbinNamespace);
  await k8sApply([apiRuleAllowObj], httpbinNamespace);
  await k8sApply(apiRuleOAuthObj, httpbinNamespace);
  await waitForDeployment('httpbin', httpbinNamespace);
  const apiRulePath = `/apis/gateway.kyma-project.io/v1alpha1/namespaces/${httpbinNamespace}/apirules`;
  await waitForK8sObject(apiRulePath, {}, (_type, _apiObj, watchObj) => {
    return (watchObj.object.metadata.name == httpbinAllowService && watchObj.object.status.APIRuleStatus.code == 'OK');
  }, 60 * 1000, 'Waiting for APIRule to be ready timeout');
  await waitForK8sObject(apiRulePath, {}, (_type, _apiObj, watchObj) => {
    return (watchObj.object.metadata.name == httpbinOAuthService && watchObj.object.status.APIRuleStatus.code == 'OK');
  }, 60 * 1000, 'Waiting for APIRule to be ready timeout');
  await waitForVirtualService(httpbinNamespace, httpbinAllowService);
  await waitForVirtualService(httpbinNamespace, httpbinOAuthService);
}

function getResourcePaths(namespace) {
  return [
    `/apis/gateway.kyma-project.io/v1alpha1/namespaces/${namespace}/apirules`,
    `/apis/hydra.ory.sh/v1alpha1/namespaces/${namespace}/oauth2clients`,
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
  checkHttpbinOAuthResponse,
  checkHttpbinAllowResponse,
  cleanIstioConnectivityFixture,
};
