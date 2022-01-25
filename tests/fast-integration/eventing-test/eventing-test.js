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
} = require('../test/fixtures/commerce-mock');
const {
  waitForNamespace,
} = require('../utils');
const {
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
} = require('./utils');
const {cleanupTestingResources} = require('./utils');

describe('Eventing tests', function() {
  this.timeout(timeoutTime);
  this.slow(slowTime);

  before('Ensure the test namespaces exists', async function() {
    await waitForNamespace(testNamespace);
    await waitForNamespace(mockNamespace);
  });

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

  // runs after each test in every block
  afterEach(async function() {
    // if there was a fatal error, perform the cleanup
    if (this.currentTest.err && this.currentTest.err.code === fatalErrCode) {
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
  });

  context('with Nats backend switched back from BEB', function() {
    it('Switch Eventing Backend to Nats', async function() {
      await switchEventingBackend(backendK8sSecretName, backendK8sSecretNamespace, 'nats');
      await waitForSubscriptionsTillReady(testNamespace);
    });

    // Running Eventing end-to-end tests
    eventingE2ETestSuite();
  });
});
