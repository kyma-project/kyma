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
  checkFullyQualifiedTypeWithExactSub,
  waitForSubscriptionsTillReady,
  checkInClusterEventTracing,
} = require('../test/fixtures/commerce-mock');
const {
  getEventingBackend,
  waitForNamespace,
  switchEventingBackend,
  printAllSubscriptions,
  debug,
  isDebugEnabled,
  waitForFunction,
  waitForEndpoint,
  waitForPodWithLabelAndCondition,
  createApiRuleForService,
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
  getJetStreamStreamData,
  testDataConfigMapName,
  eventingNatsSvcName,
  eventingNatsApiRuleAName,
  eventingSinkName,
  v1alpha1SubscriptionsTypes,
  subscriptionsTypes,
  getClusterHost,
  checkFunctionReachable,
  checkEventDelivery,
  waitForV1Alpha1Subscriptions,
  waitForV1Alpha2Subscriptions,
  checkEventTracing,
  isEventingUpgradeTest,
  saveJetStreamDataForRecreateTest,
  getConfigMapWithRetries,
  checkStreamNotReCreated,
  checkConsumersNotReCreated,
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
  expect,
} = require('chai');
const {
  exposeGrafana,
  unexposeGrafana,
} = require('../monitoring');

let clusterHost = '';
let isEventingSinkDeployed = false;

describe('Eventing tests', function() {
  let natsApiRuleVSHost;
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
    this.test.retries(3);
    await waitForPodWithLabelAndCondition( telemetryOperatorLabel.key, telemetryOperatorLabel.value, kymaSystem,
        conditionReady.condition, conditionReady.status, 60_000);
  });

  before('Create an ApiRule for NATS', async () => {
    const vs = await createApiRuleForService(eventingNatsApiRuleAName,
        kymaSystem,
        eventingNatsSvcName,
        8222);
    natsApiRuleVSHost = vs.spec.hosts[0];
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
        await checkEventDelivery(clusterHost, 'binary', v1alpha1SubscriptionsTypes[i].type, eventSource, true);
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
        await checkEventDelivery(clusterHost, 'structured', v1alpha1SubscriptionsTypes[i].type, eventSource, true);
      }
    });

    it('Legacy Events with v1alpha1 subscription should be delivered to eventing-sink ', async function() {
      if (!isEventingSinkDeployed) {
        this.skip();
      }
      for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
        await checkEventDelivery(clusterHost, 'legacy', v1alpha1SubscriptionsTypes[i].type, 'test', true);
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

  // eventingMonitoringTestSuite - Runs Eventing tests for monitoring
  function eventingMonitoringTestSuite(backend, isSKR) {
    it('Run Eventing Monitoring tests', async function() {
      if (isEventingSinkDeployed && testSubscriptionV1Alpha2) {
        await eventingMonitoringTest(backend, isSKR, true);
        return;
      }
      // run old monitoring tests - deprecated - will be removed
      await eventingMonitoringTest(backend, isSKR);
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

    it('check subscription with full qualified event type and exact type matching', async function() {
      if (!testSubscriptionV1Alpha2) {
        this.skip();
      }
      await checkFullyQualifiedTypeWithExactSub(testNamespace);
    });

    if (isSKR && testCompassFlow) {
      eventingE2ETestSuiteWithCommerceMock(backend);
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

  function jsSaveStreamAndConsumersDataSuite() {
    context('save stream and consumer data', function() {
      it('saving JetStream data to test stream and consumers will not be re-created by kyma upgrade', async function() {
        if (!isEventingUpgradeTest) {
          debug('Skipping recording stream and consumer data because ENV: IS_EVENTING_UPGRADE_TEST is not set!');
          this.skip();
          return;
        }

        debug(`Using NATS host: ${natsApiRuleVSHost}`);
        await saveJetStreamDataForRecreateTest(natsApiRuleVSHost, testDataConfigMapName, testSubscriptionV1Alpha2);
      });
    });
  }

  // jsStreamAndConsumersNotReCreatedTestSuite - compares the stream creation timestamp before and after upgrade
  // and if the timestamp is the same, we conclude that the stream is not re-created.
  // It also compares the consumers creation timestamp before and after upgrade
  // and if the timestamp is the same, we conclude that the consumer is not re-created.
  function jsStreamAndConsumersNotReCreatedTestSuite() {
    context('stream and consumer not re-created check with NATS backend', function() {
      let preUpgradeStreamData = null;
      let preUpgradeConsumersData = null;

      before('fetch eventing-js-test data configMap', async function() {
        if (!isEventingUpgradeTest) {
          debug('Skipping stream and consumers not re-created check because ENV: IS_EVENTING_UPGRADE_TEST is not set!');
          this.skip();
          return;
        }

        debug(`fetch configMap: ${testDataConfigMapName}...`);
        let cm = await getConfigMapWithRetries(testDataConfigMapName, testNamespace);
        expect(cm).to.not.be.undefined;
        expect(cm.data).to.have.nested.property('stream')
        expect(cm.data).to.have.nested.property('consumers')
        preUpgradeStreamData = JSON.parse(cm.data.stream);
        preUpgradeConsumersData = JSON.parse(cm.data.consumers);
      });

      it('upgrade should not have re-created stream', async function() {
        debug('Verifying that the stream was not re-created by the kyma upgrade...');
        await checkStreamNotReCreated(natsApiRuleVSHost, preUpgradeStreamData);
      });

      it('upgrade should not have re-created consumers', async function() {
        debug('Verifying that the consumers were not re-created by the kyma upgrade...');
        await checkConsumersNotReCreated(natsApiRuleVSHost, preUpgradeConsumersData, testSubscriptionV1Alpha2);
      });
    });
  }

  //Tests
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

    // Running Eventing monitoring tests.
    eventingMonitoringTestSuite(natsBackend, isSKR);

    // Running JetStream stream and consumers not re-created by upgrade tests.
    jsStreamAndConsumersNotReCreatedTestSuite();
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

    // Running Eventing monitoring tests.
    eventingMonitoringTestSuite(bebBackend, isSKR);
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

    // Running Eventing monitoring tests.
    eventingMonitoringTestSuite(natsBackend, isSKR);

    // Record stream and consumer data for Kyma upgrade
    jsSaveStreamAndConsumersDataSuite();
  });

  after('Delete the created APIRule', async function() {
    await deleteApiRule(eventingNatsApiRuleAName, kymaSystem);
  });

  after('Unexpose Grafana', async function() {
    await unexposeGrafana(isSKR);
    this.test.retries(3);
  });
});
