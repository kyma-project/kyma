const {
  listResources,
} = require("../utils");

const {
  queryPrometheus,
} = require("./client");

const { assert } = require("chai");

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

module.exports = {
  shouldIgnoreTarget,
  shouldIgnoreAlert,
  buildScrapePoolSet,
  assertTimeSeriesExist,
};
