const { assert } = require("chai");

const {
  waitForPodWithLabel,
} = require("../utils");

const {
  prometheusPortForward,
  getPrometheusActiveTargets,
  getPrometheusAlerts,
  getPrometheusRuleGroups,
} = require("../monitoring/client");

const {
  shouldIgnoreTarget,
  shouldIgnoreAlert,
  buildScrapePoolSet,
  assertTimeSeriesExist,
  getNotRegisteredPrometheusRuleNames,
  retry,
} = require("../monitoring/helpers");

describe("Monitoring test", function () {
  this.timeout(30 * 60 * 1000); // 30 min
  this.slow(5 * 1000);

  var cancelPortForward;

  before(async () => {
    cancelPortForward = prometheusPortForward();
  });

  after(async () => {
    cancelPortForward();
  });

  it("All monitoring pods should be ready", async () => {
    let namespace = "kyma-system";
    await waitForPodWithLabel("app", "prometheus", namespace);
    await waitForPodWithLabel("app", "grafana", namespace);
    await waitForPodWithLabel("app", "prometheus-node-exporter", namespace);
    await waitForPodWithLabel(
      "app.kubernetes.io/name",
      "kube-state-metrics",
      namespace
    );
  });

  it("All Prometheus targets should be healthy", async () => {
    let unhealthyTargets = await retry(async () => {
      let activeTargets = await getPrometheusActiveTargets();
      return activeTargets
        .filter((t) => !shouldIgnoreTarget(t) && t.health != "up")
        .map((t) => `${t.labels.job}: ${t.lastError}`);
    });

    assert.isEmpty(
      unhealthyTargets,
      `Following targets are unhealthy: ${unhealthyTargets.join(", ")}`
    );
  });

  it("There should be no firing critical Prometheus alerts", async () => {
    let firingAlerts = await retry(async () => {
      let allAlerts = await getPrometheusAlerts();
      return allAlerts
        .filter((a) => !shouldIgnoreAlert(a) && a.state == "firing")
        .map((a) => a.labels.alertname);
    });

    assert.isEmpty(
      firingAlerts,
      `Following alerts are firing: ${firingAlerts.join(", ")}`
    );
  });

  it("Each Prometheus scrape pool should have a healthy target", async () => {
    let emptyScrapePools = await retry(async () => {
      let scrapePools = await buildScrapePoolSet();
      let activeTargets = await getPrometheusActiveTargets();
  
      for (const target of activeTargets) {
        scrapePools.delete(target.scrapePool);
      }
      return Array.from(scrapePools);
    });

    assert.isEmpty(
      emptyScrapePools,
      `Following scrape pools have no targets: ${emptyScrapePools.join(", ")}`);
  });

  it("All Prometheus rules should be healthy", async () => {
    let unhealthyRules = await retry(async () => {
      let ruleGroups = await getPrometheusRuleGroups();
      let allRules = ruleGroups.flatMap((g) => g.rules);
      return allRules
        .filter((r) => r.health != "ok")
        .map((r) => r.name);
    });

    assert.isEmpty(
      unhealthyRules,
      `Following rules are unhealthy: ${unhealthyRules.join(", ")}`
    );
  });

  it("Metrics used by the Kyma/Function dashboard shoud exist", async () => {
    await assertTimeSeriesExist("kube_deployment_status_replicas_available", [
      "deployment",
      "namespace",
    ]);
    await assertTimeSeriesExist("istio_requests_total", [
      "destination_service",
      "response_code",
      "source_workload",
    ]);
    await assertTimeSeriesExist("container_memory_usage_bytes", [
      "pod",
      "container",
    ]);
    await assertTimeSeriesExist(
      "kube_pod_container_resource_limits_memory_bytes",
      ["pod", "container"]
    );
    await assertTimeSeriesExist("container_cpu_usage_seconds_total", [
      "container",
      "pod",
      "namespace",
    ]);
    await assertTimeSeriesExist("kube_service_labels", ["namespace"]);
  });

  it("All prometheus rules are registered", async function () {
    let notRegisteredRules = await retry(
      getNotRegisteredPrometheusRuleNames
    );

    assert.isEmpty(
      notRegisteredRules,
      `Following rules are not picked up by Prometheus: ${notRegisteredRules.join(", ")}`
    );
  });
});
