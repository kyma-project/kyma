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
  waitForFunction,
  eventingSubscription,
  waitForEndpoint,
  waitForPodStatusWithLabel,
  waitForPodWithLabelAndCondition,
  createApiRuleForService,
  getConfigMap,
  deleteApiRule,
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
  testCompassFlow,
  testSubscriptionV1Alpha2,
  subCRDVersion,
  getNatsPods,
  getStreamConfigForJetStream,
  skipAtLeastOnceDeliveryTest,
  isStreamCreationTimeMissing,
  getJetStreamStreamData,
  streamDataConfigMapName,
  eventingNatsSvcName,
  eventingNatsApiRuleAName,
  subscriptionNames,
  eventingSinkName,
  v1alpha1SubscriptionsTypes,
  subscriptionsTypes,
  getClusterHost,
  checkFunctionReachable,
  checkEventDelivery,
  waitForV1Alpha1Subscriptions,
  waitForV1Alpha2Subscriptions,
  checkEventTracing,
} = require('./utils');
const {
  bebBackend,
  natsBackend,
  getEventMeshNamespace,
  kymaSystem,
  telemetryOperatorLabel,
  conditionReady,
  jaegerLabel,
  jaegerEndpoint,
} = require('./common/common');
const {
  assert,
  expect,
} = require('chai');
const {
  exposeGrafana,
  unexposeGrafana,
} = require('../monitoring');

let clusterHost = '';
let isEventingSinkDeployed = false;

describe('Eventing tests', function() {
  this.timeout(timeoutTime);
  this.slow(slowTime);

  before('Ensure the test and mock namespaces exist', async function() {
    await waitForNamespace(testNamespace);
    await waitForNamespace(mockNamespace);
  });

  before('Ensure tracing is ready', async function() {
    await waitForPodWithLabelAndCondition(jaegerLabel.key, jaegerLabel.value, kymaSystem, conditionReady.condition,
        conditionReady.status);
    await waitForEndpoint(jaegerEndpoint, kymaSystem);
  });

  before('Expose Grafana', async function() {
    await exposeGrafana();
    await waitForPodWithLabelAndCondition( telemetryOperatorLabel.key, telemetryOperatorLabel.value, kymaSystem,
        conditionReady.condition, conditionReady.status, 60_000);
  });

  before('Get stream config for JetStream', async function() {
    const success = await getStreamConfigForJetStream();
    assert.isTrue(success);
  });

  before('Check if eventing-sink is deployed', async function() {
    try {
      await waitForFunction(eventingSinkName, testNamespace, 30*1000);
      isEventingSinkDeployed = true;
    } catch (e) {
      debug('Eventing Sink is not deployed');
      debug(e);
      isEventingSinkDeployed = false;
    }
  });

  before('Get cluster host name from Virtual Services', async function() {
    if (!isEventingSinkDeployed) {
      return;
    }

    this.test.retries(5);

    clusterHost = await getClusterHost(eventingSinkName, testNamespace);
    expect(clusterHost).to.not.empty;
    debug(`host name fetched: ${clusterHost}`);
  });

  // eventDeliveryTestSuite - Runs Eventing tests for event delivery
  function eventDeliveryTestSuite(backend) {
    it('Wait for subscriptions to be ready', async function() {
      // important for upgrade tests
      if (!isEventingSinkDeployed) {
        return;
      }

      // waiting for v1alpha1 subscriptions
      await waitForV1Alpha1Subscriptions();

      if (testSubscriptionV1Alpha2) {
        // waiting for v1alpha2 subscriptions
        await waitForV1Alpha2Subscriptions();
      }
    });

    it('Eventing-sink function should be reachable through API Rule', async function() {
      if (!isEventingSinkDeployed) {
        this.skip();
      }
      await checkFunctionReachable(eventingSinkName, testNamespace, clusterHost);
    });

    it('Cloud Events [binary] with v1alpha1 subscription should be delivered to eventing-sink ', async function() {
      if (!isEventingSinkDeployed) {
        this.skip();
      }

      let eventSource = 'eventing-test';
      if (backend === bebBackend) {
        eventSource = getEventMeshNamespace();
      }
      for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
        await checkEventDelivery(clusterHost, 'binary', v1alpha1SubscriptionsTypes[i], eventSource, true);
      }
    });

    it('Cloud Events [structured] with v1alpha1 subscription should be delivered to eventing-sink ', async function() {
      if (!isEventingSinkDeployed) {
        this.skip();
      }

      let eventSource = 'eventing-test';
      if (backend === bebBackend) {
        eventSource = getEventMeshNamespace();
      }
      for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
        await checkEventDelivery(clusterHost, 'structured', v1alpha1SubscriptionsTypes[i], eventSource, true);
      }
    });

    it('Legacy Events with v1alpha1 subscription should be delivered to eventing-sink ', async function() {
      if (!isEventingSinkDeployed) {
        this.skip();
      }
      for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
        await checkEventDelivery(clusterHost, 'legacy', v1alpha1SubscriptionsTypes[i], 'test', true);
      }
    });

    it('Cloud Events [binary] with v1alpha2 subscription should be delivered to eventing-sink ', async function() {
      if (!isEventingSinkDeployed || !testSubscriptionV1Alpha2) {
        this.skip();
      }
      for (let i=0; i < subscriptionsTypes.length; i++) {
        await checkEventDelivery(clusterHost, 'binary', subscriptionsTypes[i].type, subscriptionsTypes[i].source);
      }
    });

    it('Cloud Events [structured] with v1alpha2 subscription should be delivered to eventing-sink ', async function() {
      if (!isEventingSinkDeployed || !testSubscriptionV1Alpha2) {
        this.skip();
      }
      for (let i=0; i < subscriptionsTypes.length; i++) {
        await checkEventDelivery(clusterHost, 'structured', subscriptionsTypes[i].type, subscriptionsTypes[i].source);
      }
    });

    it('Legacy Events with v1alpha2 subscription should be delivered to eventing-sink ', async function() {
      if (!isEventingSinkDeployed || !testSubscriptionV1Alpha2) {
        this.skip();
      }
      for (let i=0; i < subscriptionsTypes.length; i++) {
        await checkEventDelivery(clusterHost, 'legacy', subscriptionsTypes[i].type, subscriptionsTypes[i].source);
      }
    });
  }

  // eventingTracingTestSuite - Runs Eventing tracing tests
  function eventingTracingTestSuiteV2(isSKR) {
    // Only run tracing tests on OSS
    if (isSKR) {
      debug('Skipping eventing tracing tests on SKR');
      return;
    }

    it('In-cluster event should have correct tracing spans [v2]', async function() {
      if (!isEventingSinkDeployed || !testSubscriptionV1Alpha2) {
        this.skip();
      }

      await checkEventTracing(clusterHost, subscriptionsTypes[0].type, subscriptionsTypes[0].source, testNamespace);
    });
  }

  // eventingTestSuite - Runs Eventing tests
  function eventingTestSuite(backend, isSKR, testCompassFlow=false) {
    it('lastorder function should be reachable through secured API Rule', async function() {
      await checkFunctionResponse(testNamespace, mockNamespace);
    });

    it('In-cluster v1alpha1 subscription events should be delivered ' +
        '(legacy events, structured and binary cloud events)', async function() {
      await checkInClusterEventDelivery(testNamespace);
    });

    it('In-cluster v1alpha2 subscription events should be delivered ' +
        '(legacy events, structured and binary cloud events)', async function() {
      if (!testSubscriptionV1Alpha2) {
        this.skip();
      }
      await checkInClusterEventDelivery(testNamespace, true);
    });

    if (isSKR && testCompassFlow) {
      eventingE2ETestSuiteWithCommerceMock(backend);
    }

    if (backend === natsBackend) {
      testStreamNotReCreated();
      testJetStreamAtLeastOnceDelivery();
    }
  }

  // eventingE2ETestSuiteWithCommerceMock - Runs Eventing end-to-end tests with Compass
  function eventingE2ETestSuiteWithCommerceMock(backend) {
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

  // testStreamNotReCreated - compares the stream creation timestamp before and after upgrade
  // and if the timestamp is the same, we conclude that the stream is not re created.
  function testStreamNotReCreated() {
    context('stream check with JetStream backend', function() {
      let wantStreamData = null;
      let gotStreamData = null;
      before('check if stream creation timestamp is available', async function() {
        try {
          const cm = await getConfigMap(streamDataConfigMapName);
          wantStreamData = cm.data;
        } catch (err) {
          console.log('Skipping the stream recreation check due to missing configmap!');
          this.skip();
        }
        if (isStreamCreationTimeMissing(wantStreamData)) {
          console.log('Skipping the stream recreation check as the stream creation timestamp is missing!');
          this.skip();
        }
      });

      it('Get the current stream creation timestamp', async function() {
        const vs = await createApiRuleForService(eventingNatsApiRuleAName,
            kymaSystem,
            eventingNatsSvcName,
            8222);
        const vsHost = vs.spec.hosts[0];
        gotStreamData = await getJetStreamStreamData(vsHost);
      });

      it('Compare the stream creation timestamp', async function() {
        assert.equal(gotStreamData.streamName, wantStreamData.streamName);
        assert.equal(gotStreamData.streamCreationTime, wantStreamData.streamCreationTime);
      });


      it('Delete the created APIRule', async function() {
        await deleteApiRule(eventingNatsApiRuleAName, kymaSystem);
      });
    });
  }


  function testJetStreamAtLeastOnceDelivery() {
    context('with JetStream file storage', function() {
      const minute = 60 * 1000;
      const funcName = 'lastorder';
      const encodingBinary = 'binary';
      const encodingStructured = 'structured';
      const eventIdBinary = getRandomEventId(encodingBinary);
      const eventIdStructured = getRandomEventId(encodingStructured);
      const sink = `http://lastorder.${testNamespace}.svc.cluster.local`;
      const subscriptions = [
        eventingSubscription(`sap.kyma.custom.inapp.order.received.v1`,
            sink, subscriptionNames.orderReceived, testNamespace),
        eventingSubscription(`sap.kyma.custom.commerce.order.created.v1`,
            sink, subscriptionNames.orderCreated, testNamespace),
      ];

      before('check if at least once delivery tests need to be skipped', async function() {
        if (skipAtLeastOnceDeliveryTest()) {
          console.log('Skipping the at least once delivery tests for NATS JetStream');
          this.skip();
        }
      });

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
  function eventingTracingTestSuite(isSKR) {
    // Only run tracing tests on OSS
    if (isSKR) {
      debug('Skipping eventing tracing tests on SKR');
      return;
    }

    it('In-cluster event should have correct tracing spans', async function() {
      await checkInClusterEventTracing(testNamespace);
    });
  }

  // runs after each test in every block
  afterEach(async function() {
    // if the test is failed, then printing some debug logs
    if (this.currentTest.state === 'failed' && isDebugEnabled()) {
      await printAllSubscriptions(testNamespace, subCRDVersion);
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

    // Running Eventing end-to-end event delivery tests
    eventDeliveryTestSuite(natsBackend);

    // Running Eventing end-to-end tests - deprecated (will be removed)
    eventingTestSuite(natsBackend, isSKR, testCompassFlow);

    // Running Eventing tracing tests [v2]
    eventingTracingTestSuiteV2(isSKR);

    // Running Eventing tracing tests - deprecated (will be removed)
    eventingTracingTestSuite(isSKR);

    it('Run Eventing Monitoring tests', async function() {
      await eventingMonitoringTest(natsBackend, isSKR);
    });
  });

  context('with BEB backend', function() {
    // skip publishing cloud events for beb backend when event mesh credentials file is missing
    if (getEventMeshNamespace() === undefined) {
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
        await printAllSubscriptions(testNamespace, subCRDVersion);
      }
    });

    // Running Eventing end-to-end event delivery tests
    eventDeliveryTestSuite(bebBackend);

    // Running Eventing end-to-end tests - deprecated (will be removed)
    eventingTestSuite(bebBackend, isSKR, testCompassFlow);

    it('Run Eventing Monitoring tests', async function() {
      await eventingMonitoringTest(bebBackend, isSKR);
    });
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

    // Running Eventing end-to-end event delivery tests
    eventDeliveryTestSuite(natsBackend);

    // Running Eventing end-to-end tests - deprecated (will be removed)
    eventingTestSuite(natsBackend, isSKR, testCompassFlow);

    // Running Eventing tracing tests [v2]
    eventingTracingTestSuiteV2(isSKR);

    // Running Eventing tracing tests - deprecated (will be removed)
    eventingTracingTestSuite(isSKR);

    it('Run Eventing Monitoring tests', async function() {
      await eventingMonitoringTest(natsBackend, isSKR);
    });
  });

  after('Unexpose Grafana', async function() {
    await unexposeGrafana(isSKR);
  });
});
