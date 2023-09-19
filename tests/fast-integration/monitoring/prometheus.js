module.exports = {
  assertPodsExist,
  assertAllTargetsAreHealthy,
  assertNoCriticalAlertsExist,
  assertScrapePoolTargetsExist,
  assertAllRulesAreHealthy,
  assertMetricsExist,
  assertRulesAreRegistered,
};

const {assert} = require('chai');
const util = require('util');
const {
  listResources,
  sleep,
  waitForPodWithLabel,
} = require('../utils');
const {
  getPrometheusActiveTargets,
  getPrometheusAlerts,
  queryPrometheus,
  getPrometheusRuleGroups,
} = require('./client');

async function assertPodsExist() {
  const namespace = 'kyma-system';
  await waitForPodWithLabel('app', 'prometheus', namespace);
  await waitForPodWithLabel('app', 'prometheus-node-exporter', namespace);
  await waitForPodWithLabel(
      'app.kubernetes.io/name',
      'kube-state-metrics',
      namespace,
  );
}

async function assertNoCriticalAlertsExist() {
  const firingAlerts = await retry(async () => {
    const allAlerts = await getPrometheusAlerts();
    return allAlerts
        .filter((a) => !shouldIgnoreAlert(a) && a.state === 'firing');
  });

  assert.isEmpty(
      firingAlerts,
      `Following alerts are firing: ${firingAlerts.map((a) => util.inspect(a, false, null, true)).join(', ')}`,
  );
}

async function assertAllTargetsAreHealthy() {
  const unhealthyTargets = await retry(async () => {
    const activeTargets = await getPrometheusActiveTargets();
    return activeTargets
        .filter((t) => !shouldIgnoreTarget(t) && t.health !== 'up')
        .map((t) => `${t.labels.job}: ${t.lastError}`);
  });

  assert.isEmpty(
      unhealthyTargets,
      `Following targets are unhealthy: ${unhealthyTargets.join(', ')}`,
  );
}

async function assertScrapePoolTargetsExist() {
  const emptyScrapePools = await retry(async () => {
    const scrapePools = await buildScrapePoolSet();
    const activeTargets = await getPrometheusActiveTargets();

    for (const target of activeTargets) {
      scrapePools.delete(target.scrapePool.replace('serviceMonitor/', '').replace('podMonitor/', ''));
    }
    return Array.from(scrapePools);
  });

  assert.isEmpty(
      emptyScrapePools,
      `Following scrape pools have no targets: ${emptyScrapePools.join(', ')}`);
}

async function assertAllRulesAreHealthy() {
  const unhealthyRules = await retry(async () => {
    const ruleGroups = await getPrometheusRuleGroups();
    const allRules = ruleGroups.flatMap((g) => g.rules);
    return allRules
        .filter((r) => r.health !== 'ok')
        .map((r) => r.name);
  });

  assert.isEmpty(
      unhealthyRules,
      `Following rules are unhealthy: ${unhealthyRules.join(', ')}`,
  );
}

async function assertMetricsExist() {
  // Object with exporter, its corressponding metrics followed by labels and resources.
  const metricsList = [
    {
      'monitoring-kubelet': [
        {'container_memory_usage_bytes': [['pod', 'container']]},
        {'kubelet_pod_start_duration_seconds_count': [[]]}],
    },

    {
      'monitoring-apiserver': [
        {'apiserver_request_duration_seconds_bucket': [[]]},
        {'apiserver_audit_event_total': [[]]}],
    },

    {
      'monitoring-kube-state-metrics': [
        {'kube_deployment_status_replicas_available': [['deployment', 'namespace']]},
        {'kube_pod_container_resource_limits': [['pod', 'container'], ['memory']]}],
    },

    {
      'monitoring-node-exporter': [
        {'process_cpu_seconds_total': [['pod']]},
        {'go_memstats_heap_inuse_bytes': [['pod']]}],
    },

    {
      'istio-component-monitor': [
        {'istio_requests_total': [['destination_service', 'source_workload', 'response_code']]}],
    },

    {
      'monitoring-grafana': [
        {'grafana_stat_totals_dashboard': [[]]},
        {'grafana_api_dataproxy_request_all_milliseconds_sum ': [['pod']]}],
    },

    {
      'telemetry-fluent-bit': [
        {'telemetry_fsbuffer_usage_bytes': [[]]},
        {'fluentbit_input_bytes_total ': [['pod']]}],
    },
  ];

  for (let index = 0; index < metricsList.length; index++) {
    for (const [exporter, object] of Object.entries(metricsList[index])) {
      for (const [, obj] of Object.entries(object)) {
        await assertTimeSeriesExist(exporter,
            Object.keys(obj)[0],
            obj[Object.keys(obj)[0]][0],
            obj[Object.keys(obj)[0]][1]);
      }
    }
  }
}

async function assertRulesAreRegistered() {
  const notRegisteredRules = await retry(
      getNotRegisteredPrometheusRuleNames,
  );

  assert.isEmpty(
      notRegisteredRules,
      `Following rules are not picked up by Prometheus: ${notRegisteredRules.join(', ')}`,
  );
}

function shouldIgnoreTarget(target) {
  const podsToBeIgnored = [
    // Ignore the pods that are created during tests.
    '-testsuite-',
    'test',
    'nodejs16-',
    'nodejs12-',
    'upgrade',
    // Ignore the pods created by jobs which are executed after installation of control-plane.
    'compass-migration',
    'compass-director-tenant-loader-default',
    'compass-agent-configuration',
  ];

  const namespacesToBeIgnored = ['test', 'e2e'];

  return podsToBeIgnored.includes(target.pod) || namespacesToBeIgnored.includes(target.namespace);
}

function shouldIgnoreAlert(alert) {
  // List of critical alerts we care about but which we have to ignore due to Workarounds
  const criticalAlertNamesToIgnore = [
    // Sometimes the Gardener API Server has an expiring Cert Warning shortly before Rotation
    // which causes the alert to fire as false positive
    'K8sCertificateExpirationNotice',
  ];

  // List of alerts that we don't care about and should be filtered
  const alertNamesToIgnore = [
    // Watchdog is an alert meant to ensure functioning of the entire alerting pipeline
    'Watchdog',
    // Scrape limits can be exceeded on long-running clusters and can be ignored
    'ScrapeLimitForTargetExceeded',
    // Resource overcommitment is fine for e2e test scenarios
    'KubeCPUOvercommit',
    'KubeMemoryOvercommit',
  ];

  const ignoreCriticalAlert = criticalAlertNamesToIgnore.includes(alert.labels.alertname);

  const ignoreAlert =
    alert.labels.severity !== 'critical' ||
    alertNamesToIgnore.includes(alert.labels.alertname);

  return ignoreCriticalAlert || ignoreAlert;
}

async function getServiceMonitors() {
  const path = '/apis/monitoring.coreos.com/v1/servicemonitors';

  const resources = await listResources(path);

  return resources.filter((r) => !shouldIgnoreServiceMonitor(r.metadata.name));
}

function shouldIgnoreServiceMonitor(serviceMonitorName) {
  const serviceMonitorsToBeIgnored = [
    // tracing-metrics is created automatically by jaeger operator and can't be disabled
    'tracing-metrics',
  ];
  return serviceMonitorsToBeIgnored.includes(serviceMonitorName) || !serviceMonitorName.startsWith('monitoring');
}

async function buildScrapePoolSet() {
  const serviceMonitors = await getServiceMonitors();
  const scrapePools = new Set();

  for (const monitor of serviceMonitors) {
    const endpoints = monitor.spec.endpoints;
    for (let i = 0; i < endpoints.length; i++) {
      const scrapePool = `${monitor.metadata.namespace}/${monitor.metadata.name}/${i}`;
      scrapePools.add(scrapePool);
    }
  }
  return scrapePools;
}

async function assertTimeSeriesExist(exporter, metric, labels, resource = '') {
  const resultlessQueries = [];
  let result = '';
  let query = '';

  for (const label of labels) {
    if (resource === '') {
      query = `topk(10,${metric}{${label}=~\"..*\"})`;
      result = await queryPrometheus(query);
    } else {
      query = `topk(10,${metric}{${label}=~\"..*\", resource=\"${resource}\"})`;
      result = await queryPrometheus(query);
    }

    if (result.length === 0) {
      resultlessQueries.push(query.concat('metric from service monitor: '.concat(exporter)));
    }
  }
  assert.isEmpty(resultlessQueries, `Following queries return no results: ${resultlessQueries.join(', ')}`);
}

async function getK8sPrometheusRuleNames() {
  const path = '/apis/monitoring.coreos.com/v1/prometheusrules';
  const rules = await listResources(path);
  return rules.map((o) => o.metadata.name);
}

async function getRegisteredPrometheusRuleNames() {
  const rules = await getPrometheusRuleGroups();
  return rules.map((o) => o.name);
}

function removeNamePrefixes(ruleNames) {
  return ruleNames.map((rule) =>
    rule
        .replace('monitoring-', '')
        .replace('kyma-', ''),
  );
}

async function getNotRegisteredPrometheusRuleNames() {
  const registeredRules = await getRegisteredPrometheusRuleNames();
  let k8sRuleNames = await getK8sPrometheusRuleNames();
  k8sRuleNames = removeNamePrefixes(k8sRuleNames);
  return k8sRuleNames.filter((rule) => !registeredRules.includes(rule));
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
