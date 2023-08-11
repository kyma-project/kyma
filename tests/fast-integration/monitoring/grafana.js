module.exports = {
  assertPodsExist,
  assertGrafanaRedirectsExist,
  setGrafanaProxy,
  resetGrafanaProxy,
  restartProxyPod,
};

const {
  assert,
  expect,
} = require('chai');

const {
  toBase64,
  k8sApply,
  k8sDelete,
  patchDeployment,
  k8sAppsApi,
  retryPromise,
  sleep,
  waitForDeployment,
  waitForPodWithLabel,
  error, waitForPodWithLabelAndCondition,
} = require('../utils');

const {
  checkIfGrafanaIsReachable,
} = require('./client');

const kymaNs = 'kyma-system';
const kymaProxyDeployment = 'monitoring-auth-proxy-grafana';
const grafanaLabel = {
  key: 'app',
  value: 'grafana',
};
const conditionReady = {
  condition: 'Ready',
  status: 'True',
};
const proxySecret = {
  apiVersion: 'v1',
  kind: 'Secret',
  metadata: {
    name: 'monitoring-auth-proxy-grafana-user',
    namespace: kymaNs,
  },
  type: 'Opaque',
  data: {
    OAUTH2_PROXY_SKIP_PROVIDER_BUTTON: toBase64('true'),
  },
};

async function assertPodsExist() {
  await waitForPodWithLabel('app', 'grafana', kymaNs);
}

async function assertGrafanaRedirectsExist() {
  await assertGrafanaRedirects();
}

async function assertGrafanaRedirects() {
  let res = await checkGrafanaRedirect('https://kyma-project.io', 403);
  assert.isTrue(res, 'Grafana redirect to kyma docs does not work!');

  await createBasicProxySecret();
  await restartProxyPod();

  res = await checkGrafanaRedirect('https://accounts.google.com/signin/oauth', 200);
  assert.isTrue(res, 'Grafana redirect to google does not work!');

  await createProxySecretWithIPAllowlisting();
  // Remove the --reverse-proxy flag from the deployment to make the whitelisting also working for old deployment
  // versions in the upgrade tests
  await patchProxyDeployment('--reverse-proxy=true');
  await restartProxyPod();

  res = await checkGrafanaRedirect('https://grafana.', 200);
  assert.isTrue(res, 'Grafana redirect to grafana landing page does not work!');
}

async function setGrafanaProxy() {
  await createProxySecretWithIPAllowlisting();
  // Remove the --reverse-proxy flag from the deployment to make the whitelisting also working for old deployment
  // versions in the upgrade tests
  await restartProxyPod();
  await patchProxyDeployment('--reverse-proxy=true');

  const res = await checkGrafanaRedirect('https://grafana.', 200);
  assert.isTrue(res, 'Grafana redirect to grafana landing page does not work!');
}

async function resetGrafanaProxy(isSkr) {
  await deleteProxySecret();
  await restartProxyPod();

  const docsUrl = (isSkr ? 'https://help.sap.com/docs/BTP/' : 'https://kyma-project.io/docs');
  const res = await checkGrafanaRedirect(docsUrl, 403);
  assert.isTrue(res, 'Authproxy reset was not successful. Grafana is not redirected to kyma docs!');
}


async function patchProxyDeployment(toRemove) {
  const deployment = await retryPromise(
      async () => {
        return k8sAppsApi.readNamespacedDeployment(kymaProxyDeployment, kymaNs);
      },
      12,
      5000,
  ).catch((err) => {
    error(err);
    throw new Error(`Timeout: ${kymaProxyDeployment} is not found`);
  });

  const argPosFrom = deployment.body.spec.template.spec.containers[0].args.findIndex(
      (arg) => arg.toString().includes(toRemove),
  );

  if (argPosFrom === -1) {
    return;
  }

  const patch = [
    {
      op: 'remove',
      path: `/spec/template/spec/containers/0/args/${argPosFrom}`,
    },
  ];

  await patchDeployment(kymaProxyDeployment, kymaNs, patch);
  const patchedDeployment = await k8sAppsApi.readNamespacedDeployment(kymaProxyDeployment, kymaNs);
  expect(patchedDeployment.body.spec.template.spec.containers[0].args.findIndex(
      (arg) => arg.toString().includes(toRemove),
  )).to.equal(-1);

  // We have to wait for the deployment to redeploy the actual pod.
  await sleep(1000);
  await waitForDeployment(kymaProxyDeployment, kymaNs);
}

async function createBasicProxySecret() {
  await k8sApply([proxySecret], kymaNs);
}

async function createProxySecretWithIPAllowlisting() {
  const secret = proxySecret;
  secret.data.OAUTH2_PROXY_TRUSTED_IPS = toBase64('0.0.0.0/0');
  await k8sApply([secret], kymaNs);
}

async function deleteProxySecret() {
  await k8sDelete([proxySecret], kymaNs);
}

async function restartProxyPod() {
  const patchRep0 = [
    {
      op: 'replace',
      path: '/spec/replicas',
      value: 0,
    },
  ];
  await patchDeployment(kymaProxyDeployment, kymaNs, patchRep0);
  const patchedDeploymentRep0 = await k8sAppsApi.readNamespacedDeployment(kymaProxyDeployment, kymaNs);
  expect(patchedDeploymentRep0.body.spec.replicas).to.be.equal(0);

  const patchRep1 = [
    {
      op: 'replace',
      path: '/spec/replicas',
      value: 1,
    },
  ];
  await patchDeployment(kymaProxyDeployment, kymaNs, patchRep1);
  const patchedDeploymentRep1 = await k8sAppsApi.readNamespacedDeployment(kymaProxyDeployment, kymaNs);
  expect(patchedDeploymentRep1.body.spec.replicas).to.be.equal(1);

  // We have to wait for the deployment to redeploy the actual pod.
  await sleep(1_000);
  await waitForDeployment(kymaProxyDeployment, kymaNs);
  await waitForPodWithLabelAndCondition(grafanaLabel.key, grafanaLabel.value, kymaNs, conditionReady.condition,
      conditionReady.status, 60_000);
}

async function checkGrafanaRedirect(redirectURL, httpStatus) {
  let retries = 0;
  while (retries < 20) {
    const isReachable = await checkIfGrafanaIsReachable(redirectURL, httpStatus);
    if (isReachable === true) {
      return true;
    }
    await sleep(5 * 1000);
    retries++;
  }
  return false;
}
