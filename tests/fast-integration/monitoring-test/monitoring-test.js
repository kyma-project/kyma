const uuid = require("uuid");
const axios = require('axios');

const {
  assert
} = require("chai");

const {
  debug,
  genRandom,
  kubectlPortForward,
  waitForPodWithLabel,
} = require("../utils");

const {
  shouldIgnoreTarget,
  shouldIgnoreAlert,
  buildScrapePoolSet,
  checkMetricWithLabels,
} = require('../monitoring/helpers')

describe("Monitoring test", function () {

  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const scenarioName = `test-${suffix}`;
  const runtimeID = uuid.v4();

  debug(`RuntimeID ${runtimeID}`, `Scenario ${scenarioName}`, `Runtime ${runtimeName}`, `Application ${appName}`);

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  var cancelPortForward;
  let prometheusPort = 9090;

  before(() => {
    cancelPortForward = kubectlPortForward("kyma-system", "prometheus-monitoring-prometheus-0", prometheusPort);
  })

  after(() => {
    cancelPortForward()
  })

  it("All targets should be healthy", async () => {
    let response = await axios.get(`http://localhost:${prometheusPort}/api/v1/targets?state=active`);
    let responseBody = response.data;
    let activeTargets = responseBody.data.activeTargets;
    let unhealthyTargets = activeTargets.filter(t => !shouldIgnoreTarget(t) && t.health != "up").map(t => t.discoveredLabels.job);

    assert.isEmpty(unhealthyTargets, `Following targets are unhealthy: ${unhealthyTargets.join(", ")}`);
  });

  it("There should be no firing critical alerts", async () => {
    let response = await axios.get(`http://localhost:${prometheusPort}/api/v1/alerts`);
    let responseBody = response.data;
    let allAlerts = responseBody.data.alerts;
    let firingAlerts = allAlerts.filter(a => !shouldIgnoreAlert(a) && a.state == 'firing').map(a => a.labels.alertname);

    assert.isEmpty(firingAlerts, `Following alerts are firing: ${firingAlerts.join(", ")}`);
  });

  it("All pods should be ready", async () => {
    let namespace = "kyma-system";
    await waitForPodWithLabel("app", "alertmanager", namespace);
    await waitForPodWithLabel("app", "prometheus", namespace);
    await waitForPodWithLabel("app", "grafana", namespace);
    await waitForPodWithLabel("app", "prometheus-node-exporter", namespace);
    await waitForPodWithLabel("app.kubernetes.io/name", "kube-state-metrics1", namespace);
  });

  it("Each scrape pool should have a healthy target", async () => {
    let scrapePools = await buildScrapePoolSet();

    let response = await axios.get(`http://localhost:${prometheusPort}/api/v1/targets?state=active`);
    let responseBody = response.data;
    let activeTargets = responseBody.data.activeTargets;

    for (const target of activeTargets) {
      scrapePools.delete(target.scrapePool);
    }
    assert.isEmpty(scrapePools, `Following scrape pools have no targets: ${Array.from(scrapePools).join(", ")}`)
  });

  it("All rules should be healthy", async () => {
    let response = await axios.get(`http://localhost:${prometheusPort}/api/v1/rules`);
    let responseBody = response.data;
    let allRules = responseBody.data.groups.flatMap(g => g.rules);
    let unhealthyRules = allRules.filter(r => r.health != "ok").map(t => r.name);

    assert.isEmpty(unhealthyRules, `Following rules are unhealthy: ${unhealthyRules.join(", ")}`);
  });

  it("Lambda UI dashboard should be ready", async () => { // TODO: Maybe rename
    await checkMetricWithLabels("kube_deployment_status_replicas_available", ["deployment", "namespace"]);
    await checkMetricWithLabels("istio_requests_total", ["destination_service", "response_code", "source_workload"]);
    await checkMetricWithLabels("container_memory_usage_bytes", ["pod", "container"]);
    await checkMetricWithLabels("kube_pod_container_resource_limits_memory_bytes", ["pod", "container"]);
    await checkMetricWithLabels("container_cpu_usage_seconds_total", ["container", "pod", "namespace"]);
    await checkMetricWithLabels("kube_namespace_labels", ["label_istio_injection"]);
    await checkMetricWithLabels("kube_service_labels", ["namespace"]);
  });
});
