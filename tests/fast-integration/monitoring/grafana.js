const {
  assert,
  expect,
} = require('chai');

const {
  getEnvOrDefault,
  toBase64,
  k8sApply,
  k8sDelete,
  patchDeployment,
  k8sAppsApi,
  retryPromise,
  sleep,
  waitForDeployment,
  waitForPodWithLabel,
  info,
  error,
} = require('../utils');

const {
  queryGrafana,
  getGrafanaUrl,
} = require('./client');

const kymaNs = 'kyma-system';
const kymaProxyDeployment = 'monitoring-auth-proxy-grafana';


async function assertPodsExist() {
  await waitForPodWithLabel('app', 'grafana', kymaNs);
}

async function assertGrafanaRedirectsExist() {
  if (getEnvOrDefault('KYMA_MAJOR_VERSION', '2') === '2') {
    await assertGrafanaRedirectsInKyma2();
  } else {
    await assertGrafanaRedirectsInKyma1();
  }
}

async function assertGrafanaRedirectsInKyma2() {
  info('Checking grafana redirect for kyma docs');
  let res = await checkGrafanaRedirect('https://kyma-project.io/docs', 403);
  assert.isTrue(res, 'Grafana redirect to kyma docs does not work!');

  // Creating secret for auth proxy redirect
  await manageSecret('create');
  await restartProxyPod();

  info('Checking grafana redirect for google as OIDC provider');
  res = await checkGrafanaRedirect('https://accounts.google.com/signin/oauth', 200);
  assert.isTrue(res, 'Grafana redirect to google does not work!');

  await updateProxyDeployment('--reverse-proxy=true', '--trusted-ip=0.0.0.0/0');

  info('Checking grafana redirect to grafana URL');
  res = await checkGrafanaRedirect('https://grafana.', 200);
  assert.isTrue(res, 'Grafana redirect to grafana landing page does not work!');
}

async function assertGrafanaRedirectsInKyma1() {
  info('Checking grafana redirect for dex');
  const res = await checkGrafanaRedirect('https://dex.', 200);
  assert.isTrue(res, 'Grafana redirect to dex does not work!');
}

async function setGrafanaProxy() {
  if (getEnvOrDefault('KYMA_MAJOR_VERSION', '2') === '2') {
    await manageSecret('create');
    await restartProxyPod();
    await updateProxyDeployment('--reverse-proxy=true', '--trusted-ip=0.0.0.0/0');
    info('Checking grafana redirect to grafana URL');
    const res = await checkGrafanaRedirect('https://grafana.', 200);
    assert.isTrue(res, 'Grafana redirect to grafana landing page does not work!');
  }
}

async function resetGrafanaProxy() {
  if (getEnvOrDefault('KYMA_MAJOR_VERSION', '2') === '2') {
    await manageSecret('delete');
    await updateProxyDeployment('--trusted-ip=0.0.0.0/0', '--reverse-proxy=true');

    info('Checking grafana redirect to kyma docs');
    const res = await checkGrafanaRedirect('https://kyma-project.io/docs', 403);
    assert.isTrue(res, 'Authproxy reset was not successful. Grafana is not redirected to kyma docs!');
  }
}

async function manageSecret(action) {
  const secret = {
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
  if (action === 'create') {
    info('Creating secret: monitoring-auth-proxy-grafana-user');
    await k8sApply([secret], kymaNs);
  } else if (action === 'delete') {
    info('Deleting secret: monitoring-auth-proxy-grafana-user');
    await k8sDelete([secret], kymaNs);
  }
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
  await sleep(1000);
  await waitForDeployment(kymaProxyDeployment, kymaNs);
}

async function updateProxyDeployment(fromArg, toArg) {
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
      (arg) => arg.toString().includes(fromArg),
  );

  const argPosTo = deployment.body.spec.template.spec.containers[0].args.findIndex(
      (arg) => arg.toString().includes(toArg),
  );

  if (argPosFrom === -1 && argPosTo !== -1) {
    info(`Skipping updating Proxy Deployment as it is already in desired state`);
    return;
  }

  const patch = [
    {
      op: 'replace',
      path: `/spec/template/spec/containers/0/args/${argPosFrom}`,
      value: toArg,
    },
  ];

  await patchDeployment(kymaProxyDeployment, kymaNs, patch);
  const patchedDeployment = await k8sAppsApi.readNamespacedDeployment(kymaProxyDeployment, kymaNs);
  expect(patchedDeployment.body.spec.template.spec.containers[0].args.findIndex(
      (arg) => arg.toString().includes(toArg),
  )).to.not.equal(-1);

  // We have to wait for the deployment to redeploy the actual pod.
  await sleep(1000);
  await waitForDeployment(kymaProxyDeployment, kymaNs);
}

async function checkGrafanaRedirect(redirectURL, httpStatus) {
  const url = await getGrafanaUrl();
  let ignoreSSL = false;
  if (url.includes('local.kyma.dev')) {
    ignoreSSL = true;
  }
  return await retryUrl(url, redirectURL, ignoreSSL, httpStatus);
}

async function retryUrl(url, redirectURL, ignoreSSL, httpStatus) {
  let retries = 0;
  while (retries < 20) {
    const res = await queryGrafana(url, redirectURL, ignoreSSL, httpStatus);
    if (res === true) {
      return res;
    }
    await sleep(5 * 1000);
    retries++;
  }
  return false;
}

module.exports = {
  assertPodsExist,
  assertGrafanaRedirectsExist,
  setGrafanaProxy,
  resetGrafanaProxy,
};
