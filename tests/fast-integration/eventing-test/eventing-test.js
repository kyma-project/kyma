const axios = require('axios');
const https = require('https');
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  checkFunctionResponse,
  sendEventAndCheckResponse,
  checkInClusterEventDelivery,
  waitForSubscriptionsTillReady,
  sendLegacyEventAndCheckTracing,
  checkInClusterEventTracing,
} = require('../test/fixtures/commerce-mock');
const {
  waitForNamespace,
  switchEventingBackend,
  printAllSubscriptions,
  printEventingControllerLogs,
  printEventingPublisherProxyLogs,
} = require('../utils');
const {
  testNamespace,
  backendK8sSecretName,
  backendK8sSecretNamespace,
  DEBUG_MODE,
  timeoutTime,
  slowTime,
  mockNamespace,
  fatalErrCode,
  cleanupTestingResources,
  isSKR,
} = require('./utils');
const {prometheusPortForward} = require("../monitoring/client")
const {eventingMonitoringTest} = require("./metric-test")

describe('Eventing tests', function() {
  this.timeout(timeoutTime);
  this.slow(slowTime);
  let cancelPrometheusPortForward = null;

  before('Ensure the test namespaces exists', async function() {
    await waitForNamespace(testNamespace);
    await waitForNamespace(mockNamespace);

    // Set port-forward to prometheus pod
    cancelPrometheusPortForward = prometheusPortForward();
  });

  after('Ensure the Prometheus port-forward is terminated', async function (){
    if (cancelPrometheusPortForward) {
      cancelPrometheusPortForward();
    }
  })

  // eventingE2ETestSuite - Runs Eventing end-to-end tests
  function eventingE2ETestSuite() {
    it('In-cluster event should be delivered (structured and binary mode)', async function() {
      await checkInClusterEventDelivery(testNamespace);
    });

    it('lastorder function should be reachable through secured API Rule', async function() {
      await checkFunctionResponse(testNamespace);
    });

    it('order.created.v1 event from CommerceMock should trigger the lastorder function', async function() {
      await sendEventAndCheckResponse();
    });
  }

  // eventingTracingTestSuite - Runs Eventing tracing tests
  function eventingTracingTestSuite () {
    // Only run tracing tests on OSS
    if (isSKR) {
      console.log("Skipping eventing tracing tests on SKR...")
      return
    }

    it("order.created.v1 event from CommerceMock should have correct tracing spans", async function () {
      await sendLegacyEventAndCheckTracing(testNamespace, mockNamespace);
    });

    it("In-cluster event should have correct tracing spans", async function () {
      await checkInClusterEventTracing(testNamespace);
    });
  }

  // runs after each test in every block
  afterEach(async function() {
    // if there was a fatal error, perform the cleanup
    if (this.currentTest.err && this.currentTest.err.code === fatalErrCode) {
      console.log("Cleaning up test resources...")
      await cleanupTestingResources();
    }

    // if the test is failed, then printing some debug logs
    if (this.currentTest.state === 'failed' && DEBUG_MODE) {
      await printAllSubscriptions(testNamespace);
      await printEventingControllerLogs();
      await printEventingPublisherProxyLogs();
    }
  });

  // Tests
  context('with Nats backend', function() {
    // Running Eventing end-to-end tests
    eventingE2ETestSuite();
    // Running Eventing tracing tests
    eventingTracingTestSuite();
    // Running Eventing Monitoring tests
    eventingMonitoringTest('nats');
  });

  context('with BEB backend', function() {
    it('Switch Eventing Backend to BEB', async function() {
      await switchEventingBackend(backendK8sSecretName, backendK8sSecretNamespace, 'beb');
      await waitForSubscriptionsTillReady(testNamespace);

      // print subscriptions status when debugLogs is enabled
      if (DEBUG_MODE) {
        await printAllSubscriptions(testNamespace);
      }
    });

    // Running Eventing end-to-end tests
    eventingE2ETestSuite();
    // Running Eventing Monitoring tests
    eventingMonitoringTest('beb');
  });

  context('with Nats backend switched back from BEB', function() {
    it('Switch Eventing Backend to Nats', async function() {
      await switchEventingBackend(backendK8sSecretName, backendK8sSecretNamespace, 'nats');
      await waitForSubscriptionsTillReady(testNamespace);
    });

    // Running Eventing end-to-end tests
    eventingE2ETestSuite();
    // Running Eventing tracing tests
    eventingTracingTestSuite();
    // Running Eventing Monitoring tests
    eventingMonitoringTest('nats');
  });
});
