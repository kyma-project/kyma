const { assert } = require("chai");
const util = require('util')

const {
    listResources,
    sleep,
    waitForPodWithLabel,
} = require("../utils");

const {
    getPrometheusActiveTargets,
    getPrometheusAlerts,
    queryPrometheus,
    getPrometheusRuleGroups,
} = require("./client");

async function assertPodsExist() {
    let namespace = "kyma-system";
    await waitForPodWithLabel("app", "prometheus", namespace);
    await waitForPodWithLabel("app", "prometheus-node-exporter", namespace);
    await waitForPodWithLabel(
        "app.kubernetes.io/name",
        "kube-state-metrics",
        namespace
    );
}

async function assertAllTargetsAreHealthy() {
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
}

async function assertNoCriticalAlertsExist() {
    let firingAlerts = await retry(async () => {
        let allAlerts = await getPrometheusAlerts();
        return allAlerts
            .filter((a) => !shouldIgnoreAlert(a) && a.state == "firing");
    });

    assert.isEmpty(
        firingAlerts,
        `Following alerts are firing: ${firingAlerts.map((a) => util.inspect(a, false, null, true)).join(", ")}`
    );
}

async function assertScrapePoolTargetsExist() {
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
}

async function assertAllRulesAreHealthy() {
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
}

async function assertMetricsExist() {
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
        "kube_pod_container_resource_limits",
        ["pod", "container"],
        "memory"
    );
    await assertTimeSeriesExist("container_cpu_usage_seconds_total", [
        "container",
        "pod",
        "namespace",
    ]);
    await assertTimeSeriesExist("kube_service_labels", ["namespace"]);
}

async function assertRulesAreRegistered() {
    let notRegisteredRules = await retry(
        getNotRegisteredPrometheusRuleNames
    );

    assert.isEmpty(
        notRegisteredRules,
        `Following rules are not picked up by Prometheus: ${notRegisteredRules.join(", ")}`
    );
}

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
    // List of alerts that we don't care about and should be filtered
    var alertNamesToIgnore = [
        // Watchdog is an alert meant to ensure that the entire alerting pipeline is functional
        "Watchdog",
        // Scrape limits can be exceeded on long-running clusters and can be ignored
        "ScrapeLimitForTargetExceeded",
        // Overcommitting resources is fine for e2e test scenarios
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

async function assertTimeSeriesExist(metric, labels, resource="") {
    let resultlessQueries = []
    let result = ""
    let query = ""

    for (const label of labels) {
        if (resource === "") {
            query = `topk(10,${metric}{${label}=~\"..*\"})`;
            result = await queryPrometheus(query);
        } else {
            query = `topk(10,${metric}{${label}=~\"..*\", resource=\"${resource}\"})`;
            result = await queryPrometheus(query);
        }

        if (result.length == 0) {
            resultlessQueries.push(query);
        }
    }
    assert.isEmpty(resultlessQueries, `Following queries return no results: ${resultlessQueries.join(", ")}`)
}

async function getK8sPrometheusRuleNames() {
    let path = '/apis/monitoring.coreos.com/v1/prometheusrules';
    let rules = await listResources(path);
    return rules.map((o) => o.metadata.name);
}

async function getRegisteredPrometheusRuleNames() {
    let rules = await getPrometheusRuleGroups();
    return rules.map((o) => o.name);
}

function removeNamePrefixes(ruleNames) {
    return ruleNames.map((rule) =>
        rule
            .replace("monitoring-", "")
            .replace("kyma-", "")
            .replace("logging-", "")
            .replace("fluent-bit-", "")
            .replace("loki-", "")
    );
}

async function getNotRegisteredPrometheusRuleNames() {
    let registeredRules = await getRegisteredPrometheusRuleNames();
    let k8sRuleNames = await getK8sPrometheusRuleNames();
    k8sRuleNames = removeNamePrefixes(k8sRuleNames);
    let notRegisteredRules = k8sRuleNames.filter((rule) => !registeredRules.includes(rule));
    return notRegisteredRules;
}

// Retries to execute getList() {maxRetries} times every {interval} ms until the returned list is empty
async function retry(getList, maxRetries = 20, interval = 5 * 1000) {
    let list = [];
    let retries = 0;
    while (retries < maxRetries) {
        list = await getList();
        if (list.length === 0) {
            break;
        }
        await sleep(interval);
        retries++;
    }
    return list;
}

module.exports = {
    assertPodsExist,
    assertAllTargetsAreHealthy,
    assertNoCriticalAlertsExist,
    assertScrapePoolTargetsExist,
    assertAllRulesAreHealthy,
    assertMetricsExist,
    assertRulesAreRegistered,
}
