const {
  listResources,
  getVirtualService,
  sleep,
  waitForDeployment,
  retryPromise,
  k8sApply,
  k8sDelete,
  toBase64,
  patchDeployment,
  k8sAppsApi,
} = require("../utils");
const k8s = require("@kubernetes/client-node");
const { expect } = require("chai");

const {
  queryPrometheus,
  queryGrafana,
} = require("./client");

const { assert } = require("chai");
const { V1ServerAddressByClientCIDR } = require("@kubernetes/client-node");

function shouldIgnoreTarget(target) {
  let podsToBeIgnored = [
    // Ignore the pods that are created during tests.
    "-testsuite-",
    "test",
    "nodejs12-",
    "nodejs14-",
    "upgrade",
    // Ignore the pods created by jobs which are executed after installation of control-plane.
    "compass-migration",
    "compass-director-tenant-loader-default",
    "compass-agent-configuration",
  ];

  let namespacesToBeIgnored = ["test", "e2e"];

  return podsToBeIgnored.includes(target.pod) || namespacesToBeIgnored.includes(target.namespace);
}

function shouldIgnoreAlert(alert) {
  var alertNamesToIgnore = [
    // Watchdog is an alert meant to ensure that the entire alerting pipeline is functional, so it should always be firing,
    "Watchdog",
    // Scrape limits can be exceeded on long-running clusters and can be ignored
    "ScrapeLimitForTargetExceeded",

    // Overcommitting means that a cluster with at least 1 node taken down doesn't have enough resources to run all the pods
    // It's fine in an e2e test scenario since the clusters are usually deliberately created small to save money
    "KubeCPUOvercommit",
    "KubeMemoryOvercommit",
  ]

  return alert.labels.severity == "critical" || alertNamesToIgnore.includes(alert.labels.alertname)
}

async function getServiceMonitors() {
  let path = '/apis/monitoring.coreos.com/v1/servicemonitors'

  let resources = await listResources(path);

  return resources.filter(r => !shouldIgnoreServiceMonitor(r.metadata.name));
}

async function getPodMonitors() {
  let path = '/apis/monitoring.coreos.com/v1/podmonitors'

  let resources = await listResources(path);

  return resources.filter(r => !shouldIgnorePodMonitor(r.metadata.name));
}

function shouldIgnoreServiceMonitor(serviceMonitorName) {
  var serviceMonitorsToBeIgnored = [
    // tracing-metrics is created automatically by jaeger operator and can't be disabled
    "tracing-metrics",
  ]
  return serviceMonitorsToBeIgnored.includes(serviceMonitorName);
}

function shouldIgnorePodMonitor(podMonitorName) {
  var podMonitorsToBeIgnored = [
    // The targets scraped by these podmonitors will be tested here: https://github.com/kyma-project/kyma/issues/6457
  ]
  return podMonitorsToBeIgnored.includes(podMonitorName);
}

async function buildScrapePoolSet() {
  let serviceMonitors = await getServiceMonitors();
  let podMonitors = await getPodMonitors();

  let scrapePools = new Set();

  for (const monitor of serviceMonitors) {
    let endpoints = monitor.spec.endpoints
    for (let i = 0; i < endpoints.length; i++) {
      let scrapePool = `${monitor.metadata.namespace}/${monitor.metadata.name}/${i}`
      scrapePools.add(scrapePool);
    }
  }
  for (const monitor of podMonitors) {
    let endpoints = monitor.spec.podmetricsendpoints
    for (let i = 0; i < endpoints.length; i++) {
      let scrapePool = `${monitor.metadata.namespace}/${monitor.metadata.name}/${i}`
      scrapePools.add(scrapePool);
    }
  }
  return scrapePools
}

async function assertTimeSeriesExist(metric, labels) {
  let resultlessQueries = []
  for (const label of labels) {
    let query = `topk(10,${metric}{${label}=~\"..*\"})`;
    let result = await queryPrometheus(query);

    if (result.length == 0) {
      resultlessQueries.push(query);
    }
  }
  assert.isEmpty(resultlessQueries, `Following queries return no results: ${resultlessQueries.join(", ")}`)
}

async function retryUrl(url, redirectURL, ignoreSSL, httpStatus) {
  let retries = 0
  while (retries < 20) {
    let res = await queryGrafana(url, redirectURL, ignoreSSL, httpStatus)
    if (res === true) {
      return res
    }
    await sleep(5*1000)
    retries++
  }
  return false
}

async function assertGrafanaRedirect(redirectURL) {
  let vs = await getVirtualService("kyma-system", "monitoring-grafana")
  let ignoreSSL = false
  if (vs.includes("local.kyma.dev")) {
    ignoreSSL = true
  }
  let url = "https://"+vs
  if (redirectURL.includes("https://dex.")) {
    console.log("Checking redirect for dex")
    return await retryUrl(url, redirectURL, ignoreSSL, 200)
  }

  if (redirectURL.includes("https://kyma-project.io/docs")) {
    console.log("Checking redirect for kyma docs")
    return await retryUrl(url, redirectURL, ignoreSSL, 403)
  }

  if (redirectURL.includes("https://accounts.google.com/signin/oauth")) {
    console.log("Checking redirect for google")
    return await retryUrl(url, redirectURL, ignoreSSL, 200)
  }

  if (redirectURL.includes("grafana")) {
    console.log("Checking redirect for grafana")
    return await retryUrl(url, redirectURL, ignoreSSL, 200)
  }
}

async function manageSecret(action) {
  const sec = {
    apiVersion: "v1",
    kind: "Secret",
    metadata: {
      name: "monitoring-auth-proxy-grafana-user",
      namespace: "kyma-system",
    },
    type: "Opaque",
    data: {
      OAUTH2_PROXY_SKIP_PROVIDER_BUTTON: toBase64("true")
    },
  }
  if (action === "create") {
    console.log("Creating secret: monitoring-auth-proxy-grafana-user ")
    await k8sApply([sec], "kyma-system");
  } else if (action === "delete") {
    console.log("Deleting secret: monitoring-auth-proxy-grafana-user ")
    await k8sDelete([sec], "kyma-system");
  }
}

async function updateProxyDeployment(fromArg, toArg) {
  const name = "monitoring-auth-proxy-grafana"
  const ns = "kyma-system"

  const deployment = await retryPromise(
    async () => {
      return k8sAppsApi.readNamespacedDeployment(name, ns);
    },
    12,
    5000
  ).catch((err) => {
    throw new Error(`Timeout: ${name} is not found`);
  });

  const argPos = deployment.body.spec.template.spec.containers[0].args.findIndex(
      arg => arg.toString().includes(fromArg)
  );
  expect(argPos).to.not.equal(-1);

  const patch = [
    {
      op: "replace",
      path: `/spec/template/spec/containers/0/args/${argPos}`,
      value: toArg,
    },
  ];

  await patchDeployment(name, ns, patch)
  const patchedDeployment = await k8sAppsApi.readNamespacedDeployment(name, ns);
  expect(patchedDeployment.body.spec.template.spec.containers[0].args.findIndex(
      arg => arg.toString().includes(toArg)
  )).to.not.equal(-1);

  // We have to wait for the deployment to redeploy the actual pod.
  await sleep(1000);
  await waitForDeployment(name, ns);
}

async function restartProxyPod() {
  const name = "monitoring-auth-proxy-grafana"
  const ns = "kyma-system"

  const patchRep0 = [
    {
      op: 'replace',
      path: '/spec/replicas',
      value: 0,
    },
  ];
  await patchDeployment(name, ns, patchRep0)
  const patchedDeploymentRep0 = await k8sAppsApi.readNamespacedDeployment(name, ns);
  expect(patchedDeploymentRep0.body.spec.replicas).to.be.equal(0);

  const patchRep1 = [
    {
      op: 'replace',
      path: '/spec/replicas',
      value: 1,
    },
  ];
  await patchDeployment(name, ns, patchRep1)
  const patchedDeploymentRep1 = await k8sAppsApi.readNamespacedDeployment(name, ns);
  expect(patchedDeploymentRep1.body.spec.replicas).to.be.equal(1);

  // We have to wait for the deployment to redeploy the actual pod.
  await sleep(1000);
  await waitForDeployment(name, ns);
}

async function resetProxy() {
  // delete secret
  manageSecret("delete")
  // remove add reverse proxy
  updateProxyDeployment("--trusted-ip=0.0.0.0/0","--reverse-proxy=true")
  // Check if the redirect works like again after reset
  let res = await assertGrafanaRedirect("https://kyma-project.io/docs");
  assert.isTrue(res, "Grafana redirect to kyma docs does not work!");

  return res
}

async function checkGrafanaRedirectsInKyma1() {
  let res = await assertGrafanaRedirect("https://dex.")
  assert.isTrue(res, "Grafana redirect to dex does not work!");
}

async function checkGrafanaRedirectsInKyma2() {
  // Checking grafana redirect to kyma docs
  let res = await assertGrafanaRedirect("https://kyma-project.io/docs")
  assert.isTrue(res, "Grafana redirect to kyma docs does not work!");

  // Creating secret for auth proxy redirect
  await manageSecret("create");
  await restartProxyPod();
  // Checking grafana redirect to OIDC provider
  res = await assertGrafanaRedirect("https://accounts.google.com/signin/oauth");
  assert.isTrue(res, "Grafana redirect to google does not work!");

  await updateProxyDeployment("--reverse-proxy=true", "--trusted-ip=0.0.0.0/0");
  // Checking that authentication works and redirects to grafana URL
  res = await assertGrafanaRedirect("https://grafana.");
  assert.isTrue(res, "Grafana redirect to grafana landing page does not work!");

  res = await resetProxy() 
  assert.isTrue(res, "Grafana Authproxy is not reset successfully!  ")
}

module.exports = {
  shouldIgnoreTarget,
  shouldIgnoreAlert,
  buildScrapePoolSet,
  assertTimeSeriesExist,
  assertGrafanaRedirect,
  restartProxyPod,
  updateProxyDeployment,
  checkGrafanaRedirectsInKyma1,
  checkGrafanaRedirectsInKyma2
};
