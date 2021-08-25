const uuid = require("uuid");

const { assert, expect } = require("chai");

const {
  waitForPodWithLabel,
  k8sCoreV1Api,
  k8sApply,
  k8sDelete,
  wait,
} = require("../utils");
const k8s = require("@kubernetes/client-node");
const { getNotRegisteredPrometheusRuleNames } = require("./alert-rules");

const fs = require("fs");
const path = require("path");

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
} = require("../monitoring/helpers");

describe("Monitoring test", function () {
  this.timeout(30 * 60 * 1000); // 30 min
  this.slow(5 * 1000);

  var cancelPortForward;

  const testNamespace = "test-prometheusrules";

  before(async () => {
    cancelPortForward = prometheusPortForward();
    try {
      await k8sCoreV1Api.createNamespace({
        metadata: { name: testNamespace },
      });
    } catch (error) {
      console.log(`Namespace ${testNamespace} could not be created`, error);
    }
  });

  after(async () => {
    cancelPortForward();
    await k8sCoreV1Api.deleteNamespace(testNamespace);
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
    let activeTargets = await getPrometheusActiveTargets();
    let unhealthyTargets = activeTargets
      .filter((t) => !shouldIgnoreTarget(t) && t.health != "up")
      .map((t) => `${t.labels.job}: ${t.lastError}`);

    assert.isEmpty(
      unhealthyTargets,
      `Following targets are unhealthy: ${unhealthyTargets.join(", ")}`
    );
  });

  it("There should be no firing critical Prometheus alerts", async () => {
    let allAlerts = await getPrometheusAlerts();
    let firingAlerts = allAlerts
      .filter((a) => !shouldIgnoreAlert(a) && a.state == "firing")
      .map((a) => a.labels.alertname);

    assert.isEmpty(
      firingAlerts,
      `Following alerts are firing: ${firingAlerts.join(", ")}`
    );
  });

  it("Each Prometheus scrape pool should have a healthy target", async () => {
    let scrapePools = await buildScrapePoolSet();
    let activeTargets = await getPrometheusActiveTargets();

    for (const target of activeTargets) {
      scrapePools.delete(target.scrapePool);
    }

    assert.isEmpty(
      scrapePools,
      `Following scrape pools have no targets: ${Array.from(scrapePools).join(
        ", "
      )}`
    );
  });

  it("All Prometheus rules should be healthy", async () => {
    let ruleGroups = await getPrometheusRuleGroups();
    let allRules = ruleGroups.flatMap((g) => g.rules);
    let unhealthyRules = allRules
      .filter((r) => r.health != "ok")
      .map((r) => r.name);

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
    await assertTimeSeriesExist("kube_namespace_labels", [
      "label_istio_injection",
    ]);
    await assertTimeSeriesExist("kube_service_labels", ["namespace"]);
  });

  function loadCRD(filepath) {
    const _loggingConfigYaml = fs.readFileSync(path.join(__dirname, filepath), {
      encoding: "utf8",
    });
    return k8s.loadAllYaml(_loggingConfigYaml);
  }

  function sleep(ms) {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  it("All prometheus rules are registered", async function () {
    const promrule = loadCRD("./test-rule.yaml");
    await k8sApply(promrule, testNamespace);
    await sleep(60000); // wait for prometheus to update
    let notRegisteredRules = await getNotRegisteredPrometheusRuleNames();
    expect(notRegisteredRules, "Rules are not registered").to.be.empty;
  });
});
