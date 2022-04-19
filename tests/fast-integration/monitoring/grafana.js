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


const {queryGrafana, getGrafanaUrl} = require('./client');

async function assertPodsExist() {
  const namespace = 'kyma-system';
  await waitForPodWithLabel('app', 'grafana', namespace);
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

async function checkGrafanaRedirect(redirectURL, httpStatus) {
  const url = await getGrafanaUrl();
  let ignoreSSL = false;
  if (url.includes('local.kyma.dev')) {
    ignoreSSL = true;
  }
  return await retryUrl(url, redirectURL, ignoreSSL, httpStatus);
}

async function manageSecret(action) {
  const secret = {
    apiVersion: 'v1',
    kind: 'Secret',
    metadata: {
      name: 'monitoring-auth-proxy-grafana-user',
      namespace: 'kyma-system',
    },
    type: 'Opaque',
    data: {
      OAUTH2_PROXY_SKIP_PROVIDER_BUTTON: toBase64('true'),
    },
  };
  if (action === 'create') {
    info('Creating secret: monitoring-auth-proxy-grafana-user ');
    await k8sApply([secret], 'kyma-system');
  } else if (action === 'delete') {
    info('Deleting secret: monitoring-auth-proxy-grafana-user ');
    await k8sDelete([secret], 'kyma-system');
  }
}

async function restartProxyPod() {
  const name = 'monitoring-auth-proxy-grafana';
  const ns = 'kyma-system';

  const patchRep0 = [
    {
      op: 'replace',
      path: '/spec/replicas',
      value: 0,
    },
  ];
  await patchDeployment(name, ns, patchRep0);
  const patchedDeploymentRep0 = await k8sAppsApi.readNamespacedDeployment(name, ns);
  expect(patchedDeploymentRep0.body.spec.replicas).to.be.equal(0);

  const patchRep1 = [
    {
      op: 'replace',
      path: '/spec/replicas',
      value: 1,
    },
  ];
  await patchDeployment(name, ns, patchRep1);
  const patchedDeploymentRep1 = await k8sAppsApi.readNamespacedDeployment(name, ns);
  expect(patchedDeploymentRep1.body.spec.replicas).to.be.equal(1);

  // We have to wait for the deployment to redeploy the actual pod.
  await sleep(1000);
  await waitForDeployment(name, ns);
}

async function updateProxyDeployment(fromArg, toArg) {
  const name = 'monitoring-auth-proxy-grafana';
  const ns = 'kyma-system';

  const deployment = await retryPromise(
      async () => {
        return k8sAppsApi.readNamespacedDeployment(name, ns);
      },
      12,
      5000,
  ).catch((err) => {
    error(err);
    throw new Error(`Timeout: ${name} is not found`);
  });

  const argPos = deployment.body.spec.template.spec.containers[0].args.findIndex(
      (arg) => arg.toString().includes(fromArg),
  );
  expect(argPos).to.not.equal(-1);

  const patch = [
    {
      op: 'replace',
      path: `/spec/template/spec/containers/0/args/${argPos}`,
      value: toArg,
    },
  ];

  await patchDeployment(name, ns, patch);
  const patchedDeployment = await k8sAppsApi.readNamespacedDeployment(name, ns);
  expect(patchedDeployment.body.spec.template.spec.containers[0].args.findIndex(
      (arg) => arg.toString().includes(toArg),
  )).to.not.equal(-1);

  // We have to wait for the deployment to redeploy the actual pod.
  await sleep(1000);
  await waitForDeployment(name, ns);
}

async function resetGrafanaProxy() {
  if (getEnvOrDefault('KYMA_MAJOR_VERSION', '2') === '2') {
    // delete secret
    await manageSecret('delete');
    // remove add reverse proxy
    await updateProxyDeployment('--trusted-ip=0.0.0.0/0', '--reverse-proxy=true');
    // Check if the redirect works like again after reset
    const res = await assertGrafanaRedirect('https://kyma-project.io/docs');
    assert.isTrue(res, 'Authproxy reset was not successful. Grafana is not redirected to kyma docs!');
  }
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
  resetGrafanaProxy,
};
