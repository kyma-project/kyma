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
  waitForSubscriptions,
  checkInClusterEventTracing,
  getRandomEventId,
  getVirtualServiceHost,
  sendInClusterEventWithRetry,
  ensureInClusterEventReceivedWithRetry,
} = require('../test/fixtures/commerce-mock');
const {
  getEventingBackend,
  waitForNamespace,
  switchEventingBackend,
  waitForEventingBackendToReady,
  printAllSubscriptions,
  printEventingControllerLogs,
  printEventingPublisherProxyLogs,
  k8sDelete,
  debug,
  isDebugEnabled,
  k8sApply,
  deleteK8sPod,
  eventingSubscription,
  waitForPodStatusWithLabel,
} = require('../utils');
const {
  eventingMonitoringTest,
} = require('./metric-test');
const {
  testNamespace,
  backendK8sSecretName,
  backendK8sSecretNamespace,
  timeoutTime,
  slowTime,
  mockNamespace,
  isSKR,
  isJetStreamEnabled,
  isFileStorage,
  getNatsPods,
} = require('./utils');
const {
  bebBackend,
  natsBackend,
  eventMeshNamespace,
} = require('./common/common');
const {
  assert,
} = require('chai');
const {
  exposeGrafana,
  unexposeGrafana,
} = require('../monitoring');

describe('Eventing tests', function() {
  this.timeout(timeoutTime);
  this.slow(slowTime);

  before('Ensure the test and mock namespaces exist', async function() {
    await waitForNamespace(testNamespace);
    await waitForNamespace(mockNamespace);
  });

  before('Expose Grafana', async function() {
    await exposeGrafana();
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

    if (backend === natsBackend && isJetStreamEnabled && isFileStorage) {
      testJetStreamFileStorage();
    }
  }

  function testJetStreamFileStorage() {
    context('with JetStream file storage', function() {
      const minute = 60 * 1000;
      const funcName = 'lastorder';
      const encodingBinary = 'binary';
      const encodingStructured = 'structured';
      const eventIdBinary = getRandomEventId(encodingBinary);
      const eventIdStructured = getRandomEventId(encodingStructured);
      const sink = `http://lastorder.${testNamespace}.svc.cluster.local`;
      const subscriptions = [
        eventingSubscription(`sap.kyma.custom.inapp.order.received.v1`, sink, 'order-received', testNamespace),
        eventingSubscription(`sap.kyma.custom.commerce.order.created.v1`, sink, 'order-created', testNamespace),
      ];

      it('Delete subscriptions', async function() {
        await k8sDelete(subscriptions);
      });

      it('Publish events', async function() {
        const host = await getVirtualServiceHost(testNamespace, funcName);
        assert.isNotEmpty(host);

        await sendInClusterEventWithRetry(host, eventIdBinary, encodingBinary);
        await sendInClusterEventWithRetry(host, eventIdStructured, encodingStructured);
      });

      it('Delete all Nats pods', async function() {
        const natsPods = await getNatsPods();
        for (let i = 0; i < natsPods.body.items.length; i++) {
          const pod = natsPods.body.items[i];
          await deleteK8sPod(pod);
        }
      });

      it('Wait until all Nats pods are deleted', async function() {
        // Assuming that Nats pods had the status.phase equals to "Running", so if it changed to "Pending"
        // this means that they were successfully deleted and recreated.
        await waitForPodStatusWithLabel('app.kubernetes.io/name', 'nats', 'kyma-system', 'Pending', 5 * minute);
      });

      it('Wait until any Nats pod is ready', async function() {
        // When the status.phase changes from "Pending" to "Running" this means that Nats pod containers are starting.
        await waitForPodStatusWithLabel('app.kubernetes.io/name', 'nats', 'kyma-system', 'Running', 5 * minute);
      });

      it('Wait until eventing backend is ready', async function() {
        await waitForEventingBackendToReady(natsBackend, 'eventing-backend', 'kyma-system', 5 * minute);
      });

      it('Recreate subscriptions', async function() {
        await k8sApply(subscriptions);
        await waitForSubscriptions(subscriptions);
      });

      it('Wait for events to be delivered', async function() {
        const host = await getVirtualServiceHost(testNamespace, funcName);
        assert.isNotEmpty(host);

        await ensureInClusterEventReceivedWithRetry(host, eventIdBinary);
        await ensureInClusterEventReceivedWithRetry(host, eventIdStructured);
      });
    });
  }

  // eventingTracingTestSuite - Runs Eventing tracing tests
  function eventingTracingTestSuite() {
    // Only run tracing tests on OSS
    if (isSKR) {
      debug('Skipping eventing tracing tests on SKR');
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
    if (this.currentTest.state === 'failed' && isDebugEnabled()) {
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
    eventingMonitoringTest(natsBackend, isJetStreamEnabled);
  });

  context('with BEB backend', function() {
    // skip publishing cloud events for beb backend when event mesh credentials file is missing
    if (eventMeshNamespace === undefined) {
      debug('Skipping E2E eventing tests for BEB backend due to missing EVENTMESH_SECRET_FILE');
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
      if (isDebugEnabled()) {
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
    eventingE2ETestSuite(natsBackend);
    // Running Eventing tracing tests
    eventingTracingTestSuite();
    // Running Eventing Monitoring tests
    eventingMonitoringTest(natsBackend, isJetStreamEnabled);
  });

  after('Unexpose Grafana', async function() {
    await unexposeGrafana(isSKR);
  });
});
