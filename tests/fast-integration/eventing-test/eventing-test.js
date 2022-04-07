const axios = require('axios');
const https = require('https');
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  checkFunctionResponse,
  sendLegacyEventAndCheckResponse,
  sendCloudEventStructuredModeAndCheckResponse,
  sendCloudEventBinaryModeAndCheckResponse,
  sendLegacyEventAndCheckTracing,
  sendCloudEventStructuredModeAndCheckTracing,
  sendCloudEventBinaryModeAndCheckTracing,
  checkInClusterEventDelivery,
  waitForSubscriptionsTillReady,
  checkInClusterEventTracing,
} = require('../test/fixtures/commerce-mock');
const {
  getEventingBackend,
  waitForNamespace,
  switchEventingBackend,
  printAllSubscriptions,
  printEventingControllerLogs,
  printEventingPublisherProxyLogs,
} = require('../utils');
const {
  eventingMonitoringTest,
} = require('./metric-test');
const {
  testNamespace,
  backendK8sSecretName,
  backendK8sSecretNamespace,
  DEBUG_MODE,
  timeoutTime,
  slowTime,
  mockNamespace,
  natsBackend,
  bebBackend,
  isSKR,
  eventMeshNamespace,
} = require('./utils');

describe('Eventing tests', function() {
  this.timeout(timeoutTime);
  this.slow(slowTime);
  before('Ensure the test and mock namespaces exist', async function() {
    await waitForNamespace(testNamespace);
    await waitForNamespace(mockNamespace);
  });

  // eventingE2ETestSuite - Runs Eventing end-to-end tests
  function eventingE2ETestSuite(backend) {
    it('lastorder function should be reachable through secured API Rule', async function() {
      await checkFunctionResponse(testNamespace, mockNamespace);
    });

    it('In-cluster event should be delivered (structured and binary mode)', async function() {
      await checkInClusterEventDelivery(testNamespace);
    });

    it('order.created.v1 legacy event from CommerceMock should trigger the lastorder function', async function() {
      await sendLegacyEventAndCheckResponse(mockNamespace);
    });

    it('order.created.v1 cloud event from CommerceMock should trigger the lastorder function', async function() {
      await sendCloudEventStructuredModeAndCheckResponse(backend, mockNamespace);
    });

    it('order.created.v1 binary cloud event from CommerceMock should trigger the lastorder function', async function() {
      await sendCloudEventBinaryModeAndCheckResponse(backend, mockNamespace);
    });
  }

  // eventingTracingTestSuite - Runs Eventing tracing tests
  function eventingTracingTestSuite() {
    // Only run tracing tests on OSS
    if (isSKR) {
      console.log('Skipping eventing tracing tests on SKR...');
      return;
    }

    it('order.created.v1 event from CommerceMock should have correct tracing spans', async function() {
      await sendLegacyEventAndCheckTracing(testNamespace, mockNamespace);
    });
    it('order.created.v1 structured cloud event from CommerceMock should have correct tracing spans', async function() {
      await sendCloudEventStructuredModeAndCheckTracing(testNamespace, mockNamespace);
    });
    it('order.created.v1 binary cloud event from CommerceMock should have correct tracing spans', async function() {
      await sendCloudEventBinaryModeAndCheckTracing(testNamespace, mockNamespace);
    });
    it('In-cluster event should have correct tracing spans', async function() {
      await checkInClusterEventTracing(testNamespace);
    });
  }

  // runs after each test in every block
  afterEach(async function() {
    // if the test is failed, then printing some debug logs
    if (this.currentTest.state === 'failed' && DEBUG_MODE) {
      await printAllSubscriptions(testNamespace);
      await printEventingControllerLogs();
      await printEventingPublisherProxyLogs();
    }
  });

  // Tests
  context('with Nats backend', function() {
    it('Switch Eventing Backend to Nats', async function() {
      const currentBackend = await getEventingBackend();
      if (currentBackend && currentBackend.toLowerCase() === natsBackend) {
        this.skip();
      }
      await switchEventingBackend(backendK8sSecretName, backendK8sSecretNamespace, natsBackend);
    });
    it('Wait until subscriptions are ready', async function() {
      await waitForSubscriptionsTillReady(testNamespace);
    });
    // Running Eventing end-to-end tests
    eventingE2ETestSuite(natsBackend);
    // Running Eventing tracing tests
    eventingTracingTestSuite();
    // Running Eventing Monitoring tests
    eventingMonitoringTest(natsBackend);
  });

  context('with BEB backend', function() {
    // skip publishing cloud events for beb backend when event mesh credentials file is missing
    if (eventMeshNamespace === undefined) {
      console.log('Skipping E2E eventing tests for BEB backend due to missing EVENTMESH_SECRET_FILE');
      return;
    }
    it('Switch Eventing Backend to BEB', async function() {
      const currentBackend = await getEventingBackend();
      if (currentBackend && currentBackend.toLowerCase() === bebBackend) {
        this.skip();
      }
      await switchEventingBackend(backendK8sSecretName, backendK8sSecretNamespace, bebBackend);
    });
    it('Wait until subscriptions are ready', async function() {
      await waitForSubscriptionsTillReady(testNamespace); // print subscriptions status when debugLogs is enabled
      if (DEBUG_MODE) {
        await printAllSubscriptions(testNamespace);
      }
    });
    // Running Eventing end-to-end tests
    eventingE2ETestSuite(bebBackend);
    // Running Eventing Monitoring tests
    eventingMonitoringTest(bebBackend);
  });

  context('with Nats backend switched back from BEB', async function() {
    it('Switch Eventing Backend to Nats', async function() {
      const currentBackend = await getEventingBackend();
      if (currentBackend && currentBackend.toLowerCase() === natsBackend) {
        this.skip();
      }
      await switchEventingBackend(backendK8sSecretName, backendK8sSecretNamespace, natsBackend);
    });
    it('Wait until subscriptions are ready', async function() {
      await waitForSubscriptionsTillReady(testNamespace);
    });
    // Running Eventing end-to-end tests
    eventingE2ETestSuite();
    // Running Eventing tracing tests
    eventingTracingTestSuite();
    // Running Eventing Monitoring tests
    eventingMonitoringTest(natsBackend);
  });
});
