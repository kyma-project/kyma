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
const httpbinNamespace = 'api-exposure-test';

const istioTestNamespaceObj = k8s.loadYaml(istioTestNamespaceYaml);
const httpbinDeploymentObj = k8s.loadAllYaml(
    httpbinDeploymentYaml,
);
const apiRuleAllowObj = k8s.loadYaml(apiRuleAllowYaml);
const apiRuleOAuthObj = k8s.loadAllYaml(
    apiRuleOAuthYaml,
);

const defaultRetryDelayMs = 1000;
const defaultRetries = 5;

async function testHttpbinAllowResponse() {
  const vs = await waitForVirtualService(httpbinNamespace, httpbinAllowService);
  const host = vs.spec.hosts[0];

  let res = await retryPromise(
      () => getHttpbin(host, ''),
      30,
      2000,
  ).catch((err) => {
    throw convertAxiosError(err, 'Httpbin-allow GET call responded with error');
  });
  expect(res.status).to.be.equal(200);

  res = await retryPromise(
      () => postHttpbin(host, ''),
      30,
      2000,
  ).catch((err) => {
    throw convertAxiosError(err, 'Httpbin-allow POST call responded with error');
  });
  expect(res.status).to.be.equal(200);
}

async function testHttpbinOAuthResponse() {
  const vs = await waitForVirtualService(httpbinNamespace, httpbinOAuthService);
  const host = vs.spec.hosts[0];
  const domain = host.split('.').slice(1).join('.');
  const accessToken = await getOAuthToken(domain);

  // expect error when unauthorized
  let errorOccurred = false;
  try {
    res = await getHttpbin(host, '');
  } catch (err) {
    errorOccurred = true;
    expect(err.response.status).to.be.equal(401);
  }
  expect(errorOccurred).to.be.equal(true);

  let res = await retryPromise(
      () => getHttpbin(host, accessToken),
      30,
      2000,
  ).catch((err) => {
    throw convertAxiosError(err, 'Httpbin Oauth2 GET call responded with error');
  });
  expect(res.status).to.be.equal(200);
}

async function testHttpbinOAuthMethod() {
  const vs = await waitForVirtualService(httpbinNamespace, httpbinOAuthService);
  const host = vs.spec.hosts[0];
  const domain = host.split('.').slice(1).join('.');
  const accessToken = await getOAuthToken(domain);

  // expect error when using disallowed method
  let errorOccurred = false;
  try {
    res = await postHttpbin(host, accessToken);
  } catch (err) {
    errorOccurred = true;
    expect(err.response.status).to.be.equal(404);
  }
  expect(errorOccurred).to.be.equal(true);
}

async function getHttpbin(host, accessToken) {
  return await axios.get(`https://${host}/headers`, {
    headers: {
      Authorization: `bearer ${accessToken}`,
    },
  });
}

async function postHttpbin(host, accessToken) {
  return await axios.post(`https://${host}/post`, 'test data', {
    headers: {
      Authorization: `bearer ${accessToken}`,
    },
  });
}

async function getOAuthToken(domain) {
  const oAuthSecretData = await getSecretData('httpbin-client', httpbinNamespace);
  const oAuthTokenGetter = new OAuthToken(
      `https://oauth2.${domain}/oauth2/token`,
      new OAuthCredentials(oAuthSecretData['client_id'], oAuthSecretData['client_secret']),
  );
  const accessToken = oAuthTokenGetter.getToken(['read']);

  return accessToken;
}

async function ensureApiExposureFixture() {
  await retryPromise( (r)=> k8sApply([istioTestNamespaceObj]), defaultRetryDelayMs, defaultRetries);
  await retryPromise( (r)=> k8sApply(httpbinDeploymentObj, httpbinNamespace), defaultRetryDelayMs, defaultRetries);
  await retryPromise( (r)=> k8sApply([apiRuleAllowObj], httpbinNamespace), defaultRetryDelayMs, defaultRetries);
  await retryPromise( (r)=> k8sApply(apiRuleOAuthObj, httpbinNamespace), defaultRetryDelayMs, defaultRetries);
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

function cleanApiExposureFixture(wait = true) {
  for (const path of getResourcePaths(httpbinNamespace)) {
    deleteAllK8sResources(path);
  }
  return deleteNamespaces([httpbinNamespace], wait);
}

module.exports = {
  ensureApiExposureFixture,
  testHttpbinOAuthResponse,
  testHttpbinAllowResponse,
  testHttpbinOAuthMethod,
  cleanApiExposureFixture,
};
