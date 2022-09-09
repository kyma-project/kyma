/* eslint-disable max-len */
const axios = require('axios');
const https = require('https');
const {expect} = require('chai');
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;

const {
  debug,
  retryPromise,
  testCompassFlow,
  replaceAllInString,
} = require('../utils');

const {queryPrometheus} = require('../monitoring/client');
const {subscriptionNames} = require('./utils');

const dashboards = {
  // The delivery dashboard
  delivery_publisherProxy: {
    title: 'Requests to publisher proxy',
    query: 'sum by (destination_service) (rate(istio_requests_total{destination_service=~"event.*-publisher-proxy.kyma-system.svc.cluster.local", response_code=~"2.*"}[5m]))',
    backends: ['nats', 'beb'],
    // The assert function receives the `data.result` section of the query result:
    // https://prometheus.io/docs/prometheus/latest/querying/api/#instant-queries
    assert: function(result) {
      const foundMetric = result.find((res) => res.metric.destination_service.startsWith('eventing-event-publisher-proxy'));
      expect(foundMetric).to.be.not.undefined;
    },
  },
  delivery_subscribers: {
    title: 'Requests to subscribers',
    query: `
          sum (rate(
            istio_requests_total{
              source_workload=~"eventing.*controller",
              destination_workload!~"istio-.+|dex|unknown|[a-z0-9-]+-dispatcher", 
              response_code=~"2.*"
            }[5m])
          ) by (le,source_workload_namespace,source_workload,destination_workload_namespace,destination_workload,response_code)`,
    backends: ['nats'],
    assert: function(result) {
      const foundMetric = result.find((res) => res.metric.destination_workload.startsWith('lastorder'));
      expect(foundMetric).to.be.not.undefined;
    },
  },
  delivery_per_subscription: {
    title: 'Delivery per Subscription',
    query: `
         sum (delivery_per_subscription{response_code=~"[245].*"}) 
          by (namespace, subscription_name,event_type,sink,response_code)`,
    backends: ['nats'],
    assert: function(result) {
      const foundMetric = result.find((res) =>
        res.metric.subscription_name &&
          res.metric.subscription_name === subscriptionNames.orderReceived &&
          res.metric.response_code === '200',
      );
      expect(foundMetric).to.be.not.undefined;
    },
  },
  latency_eventPublisherToMessagingServer: {
    title: 'Latency of Event Publisher -> Messaging Server',
    query: 'histogram_quantile(0.99999, sum(rate(event_publish_to_messaging_server_latency_bucket{namespace="kyma-system"}[5m])) by (le,pod,namespace,service))',
    backends: ['nats', 'beb'],
    assert: function(result) {
      const foundMetric = result.find((res) =>
        res.metric.namespace.toLowerCase() === 'kyma-system' &&
        res.metric.pod.toLowerCase().startsWith('eventing-publisher-proxy'));
      expect(foundMetric).to.be.not.undefined;
    },
  },
  latency_eventDispatcherToSubscribers: {
    title: 'Latency of Event Dispatcher -> Subscribers',
    query: `
          histogram_quantile(
          0.99999, 
          sum(rate(
            istio_request_duration_milliseconds_bucket{
              source_workload=~"eventing.*controller",
              destination_workload!~"istio-.+|dex|unknown|[a-z0-9-]+-dispatcher"
            }[5m])
          ) by (le,source_workload_namespace,source_workload,destination_workload_namespace,destination_workload))
        `,
    backends: ['nats'],
    assert: function(result) {
      const foundMetric = result.find((res) =>
        res.metric.source_workload === 'eventing-controller' &&
        res.metric.destination_workload.toLowerCase().startsWith('lastorder'));
      expect(foundMetric).to.be.not.undefined;
    },
  },
  // The pods dashboard
  pods_memoryUsage: {
    title: 'Memory usage',
    // This is not the exact query used in Grafana, but it ensures memory usage of eventing components are visible
    query: `
      sum by(container, pod) 
        (container_memory_usage_bytes{job="kubelet", container!="POD", container !=""}) * on(pod) 
        group_left() 
        kube_pod_labels{label_kyma_project_io_dashboard="eventing", namespace="kyma-system"}
      `,
    backends: ['nats', 'beb'],
    assert: ensureEventingPodsArePresent,
  },
  pods_cpuUsage: {
    title: 'CPU usage',
    // This is not the exact query used in Grafana, but it ensures CPU usage of eventing components are visible
    query: `
      sum by (container, pod) 
        (irate(container_cpu_usage_seconds_total{job="kubelet", image!="", container!="POD"}[4m])) * on(pod) 
        group_left() 
        kube_pod_labels{label_kyma_project_io_dashboard="eventing", namespace="kyma-system"}
      `,
    backends: ['nats', 'beb'],
    assert: ensureEventingPodsArePresent,
  },
  pods_networkReceive: {
    title: 'Network receive',
    query: `
      sum by (pod) 
        (irate(container_network_receive_bytes_total{job="kubelet"}[4m])) * on(pod) 
        kube_pod_labels{label_kyma_project_io_dashboard="eventing", namespace="kyma-system"}
      `,
    backends: ['nats', 'beb'],
    assert: ensureEventingPodsArePresent,
  },
  pods_networkTransmit: {
    title: 'Network transmit',
    query: `
      sum by (pod) 
        (irate(container_network_transmit_bytes_total{job="kubelet"}[4m])) * on(pod) 
        kube_pod_labels{label_kyma_project_io_dashboard="eventing", namespace="kyma-system"}
      `,
    backends: ['nats', 'beb'],
    assert: ensureEventingPodsArePresent,
  },
};

const skrDashboards = {
  // The delivery dashboard
  delivery_applicationConnectivityValidator: {
    title: 'Requests to application connectivity validator',
    query: 'sum by(destination_service) (rate(istio_requests_total{destination_service="central-application-connectivity-validator.kyma-system.svc.cluster.local", response_code=~"2.*"}[5m]))',
    backends: ['nats', 'beb'],
    assert: function(result) {
      const foundMetric = result.find((res) => res.metric.destination_service.startsWith('central-application-connectivity-validator'));
      expect(foundMetric).to.be.not.undefined;
    },
  },
  // The latency dashboard
  latency_connectivityValidatorToPublisherProxy: {
    title: 'Latency of Connectivity Validator -> Event Publisher',
    query: `
        histogram_quantile(
          0.99999, 
          sum(rate(
            istio_request_duration_milliseconds_bucket{
              source_workload_namespace="kyma-system",
              source_workload="central-application-connectivity-validator",
              destination_workload_namespace="kyma-system",
              destination_workload="eventing-publisher-proxy"
            }[5m])
          ) by (le,source_workload_namespace,source_workload,destination_workload_namespace,destination_workload))
        `,
    backends: ['nats', 'beb'],
    assert: function(result) {
      const foundMetric = result.find((res) =>
        res.metric.source_workload.toLowerCase() === 'central-application-connectivity-validator' &&
          res.metric.destination_workload.toLowerCase() === 'eventing-publisher-proxy');
      expect(foundMetric).to.be.not.undefined;
    },
  },
};

// A generic assertion for the pod dashboards
function ensureEventingPodsArePresent(result) {
  let controllerFound = false; let publisherProxyFound = false; let natsFound = false;
  result.forEach((res) => {
    if (controllerFound && publisherProxyFound && natsFound) return;
    if (res.metric.pod.startsWith('eventing-nats')) natsFound = true;
    if (res.metric.pod.startsWith('eventing-controller')) controllerFound = true;
    if (res.metric.pod.startsWith('eventing-publisher-proxy')) publisherProxyFound = true;
  });
  expect(controllerFound).to.be.true;
  expect(publisherProxyFound).to.be.true;
  expect(natsFound).to.be.true;
}

function runDashboardTestCase(dashboardName, test) {
  return retryPromise(async () => {
    await queryPrometheus(test.query).then((result) => {
      debug(dashboardName + ' result: ' + JSON.stringify(result, null, 2));
      test.assert(result);
    }).catch((reason) => {
      throw new Error(reason);
    });
  }, 120, 5000);
}

async function eventingMonitoringTest(backend, isSkr, isJetStreamEnabled = false) {
  let allDashboards = dashboards;
  if (isJetStreamEnabled) {
    allDashboards = Object.assign(allDashboards, getJetStreamDashboardTests());
  }
  if (isSkr && testCompassFlow) {
    allDashboards = Object.assign(allDashboards, skrDashboards);
  }

  for (const [dashboardName, test] of Object.entries(allDashboards)) {
    if (test.backends.includes(backend)) {
      console.log('Testing dashboard: ' + test.title);
      await runDashboardTestCase(dashboardName, test);
    }
  }
}

function getJetStreamDashboardTests() {
  const dashboardVariables = {
    '$server': 'eventing-nats-0',
    '$stream': 'sap',
    '$consumer': '.*',
    '$__rate_interval': '5m',
  };

  // load JetStream dashboard queries from JSON file
  const natsDashboardQueries = require('./nats-dashboard-queries.json');

  const dashboardTests = {};
  for (let i = 0; i < natsDashboardQueries.length; i++) {
    let finalQuery = natsDashboardQueries[i]['query'];
    // replace variables in queries with values
    for (const [variableName, value] of Object.entries(dashboardVariables)) {
      finalQuery = replaceAllInString(finalQuery, variableName, value);
    }

    dashboardTests[`jetstream_dash_${i}`] = {
      title: natsDashboardQueries[i]['title'],
      query: finalQuery,
      backends: ['nats'],
      // The assert function receives the `data.result` section of the query result:
      // https://prometheus.io/docs/prometheus/latest/querying/api/#instant-queries
      assert: function(result) {
        // checking if the length of the result contains something
        expect(result).to.have.length.greaterThan(0);
      },
    };
  }
  return dashboardTests;
}

module.exports = {
  eventingMonitoringTest,
};
