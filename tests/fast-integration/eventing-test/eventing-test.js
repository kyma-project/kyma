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
  orderReceivedSubName, getRandomEventId,
} = require('../test/fixtures/commerce-mock');
const {
  getEventingBackend,
  waitForNamespace,
  switchEventingBackend,
  printAllSubscriptions,
  debug,
  isDebugEnabled,
  createK8sConfigMap,
  waitForFunction,
  waitForEndpoint,
  waitForPodWithLabelAndCondition,
  createApiRuleForService,
  getConfigMap,
  deleteApiRule, k8sApply, waitForSubscription, eventingSubscriptionV1Alpha2,
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
  isUpgradeJob,
  testCompassFlow,
  testSubscriptionV1Alpha2,
  subCRDVersion,
  isStreamCreationTimeMissing,
  isConsumerCreationTimeMissing,
  getJetStreamStreamData,
  testDataConfigMapName,
  eventingNatsSvcName,
  eventingNatsApiRuleAName,
  getSubscriptionConsumerName,
  getJetStreamConsumerData,
  eventingSinkName,
  v1alpha1SubscriptionsTypes,
  subscriptionsTypes,
  subscriptionsExactTypeMatching,
  getClusterHost,
  checkFunctionReachable,
  checkEventDelivery,
  waitForV1Alpha1Subscriptions,
  waitForV1Alpha2Subscriptions,
  checkEventTracing,
  getTimeStampsWithZeroMilliSeconds,
  saveJetStreamDataForRecreateTest,
  jetStreamTestConfigMapName,
  getConfigMapWithRetries,
  checkStreamNotReCreated,
  checkConsumerNotReCreated,
  subscriptionNames,
  deployEventingSinkFunction,
  eventingUpgradeSinkName,
  waitForEventingSinkFunction,
  publishBinaryCEToEventingSink,
  ensureEventReceivedWithRetry,
  undeployEventingFunction,
  checkFunctionUnreachable,
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

    it('Cloud Events with subscription (exact) should be delivered to eventing-sink ', async function() {
      if (!isEventingSinkDeployed || !testSubscriptionV1Alpha2 || backend !== bebBackend) {
        this.skip();
      }

      let eventSource = 'eventing-test';
      if (backend === bebBackend) {
        eventSource = getEventMeshNamespace();
      }
      for (let i=0; i < subscriptionsExactTypeMatching.length; i++) {
        await checkEventDelivery(clusterHost, 'binary', subscriptionsExactTypeMatching[i].type, eventSource);
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

    if (backend === natsBackend) {
      testStreamNotReCreated();
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

  function jsSaveStreamAndConsumersDataSuite() {
    context('save stream and consumer data', function() {
      it('saving JetStream data to test stream and consumers will not be re-created by kyma upgrade', async function() {
        if (!isEventingSinkDeployed || !testSubscriptionV1Alpha2) {
          debug(`Skipping saving JetStream test data into a configMap: ${jetStreamTestConfigMapName}`);
          this.skip();
        }
        debug(`Using NATS host: ${natsApiRuleVSHost}`);
        await saveJetStreamDataForRecreateTest(natsApiRuleVSHost, jetStreamTestConfigMapName);
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
        debug(`fetch configMap: ${jetStreamTestConfigMapName}...`);
        const cm = await getConfigMapWithRetries(jetStreamTestConfigMapName, testNamespace);
        if (!cm || !isEventingSinkDeployed || !testSubscriptionV1Alpha2) {
          debug(`Skipping stream and consumers not re-created check`);
          this.skip();
          return;
        }

        // if configMap is found, then check for re-creation
        expect(cm.data).to.have.nested.property('stream');
        expect(cm.data).to.have.nested.property('consumers');
        preUpgradeStreamData = JSON.parse(cm.data.stream);
        preUpgradeConsumersData = JSON.parse(cm.data.consumers);
      });

      it('upgrade should not have re-created stream', async function() {
        debug('Verifying that the stream was not re-created by the kyma upgrade...');
        await checkStreamNotReCreated(natsApiRuleVSHost, preUpgradeStreamData);
      });

      it('upgrade should not have re-created consumers', async function() {
        debug('Verifying that the consumers were not re-created by the kyma upgrade...');
        await checkConsumerNotReCreated(natsApiRuleVSHost, preUpgradeConsumersData);
      });
    });
  }

  // testStreamNotReCreated - compares the stream creation timestamp before and after upgrade
  // and if the timestamp is the same, we conclude that the stream is not re-created.
  function testStreamNotReCreated() {
    context('stream and consumer check with JetStream backend', function() {
      let wantStreamName = null;
      let wantStreamCreationTime = null;
      let gotStreamData = null;
      let cm;
      before('check if stream creation timestamp is available', async function() {
        try {
          cm = await getConfigMap(testDataConfigMapName);
          if (isStreamCreationTimeMissing(cm.data)) {
            debug('Skipping the stream recreation check as the stream creation timestamp is missing!');
            this.skip();
          }
          wantStreamName = cm.data.streamName;
          wantStreamCreationTime = cm.data.streamCreationTime;
        } catch (err) {
          if (err.statusCode === 404) {
            debug('Skipping the stream recreation check due to missing eventing test data configmap!');
            this.skip();
          } else {
            throw err;
          }
        }
      });

      it('Get the current stream creation timestamp', async function() {
        gotStreamData = await getJetStreamStreamData(natsApiRuleVSHost);
      });

      it('Compare the stream creation timestamp', async function() {
        assert.equal(gotStreamData.streamName, wantStreamName);
        assert.equal(
            getTimeStampsWithZeroMilliSeconds(gotStreamData.streamCreationTime),
            getTimeStampsWithZeroMilliSeconds(wantStreamCreationTime),
        );
      });
    });
  }

  // testConsumerNotReCreated - compares the consumer creation timestamp before and after upgrade
  // and if the timestamp is the same, we conclude that the consumer is not re-created.
  function testConsumerNotReCreated() {
    context('consumer check with JetStream backend', function() {
      let wantConsumerName = null;
      let wantConsumerCreationTime = null;
      let gotConsumerData = null;
      let cm;
      before('check if consumer creation timestamp is available', async function() {
        try {
          cm = await getConfigMap(testDataConfigMapName);
          if (isConsumerCreationTimeMissing(cm.data)) {
            debug('Skipping the consumer recreation check as the consumer creation timestamp is missing!');
            this.skip();
          }
          wantConsumerName = cm.data.consumerName;
          wantConsumerCreationTime = cm.data.consumerCreationTime;
        } catch (err) {
          if (err.statusCode === 404) {
            debug('Skipping the consumer recreation check due to missing eventing test data configmap!');
            this.skip();
          } else {
            throw err;
          }
        }
      });

      it('Get the current consumer creation timestamp', async function() {
        const consumerName = await getSubscriptionConsumerName(orderReceivedSubName, testNamespace, subCRDVersion);
        gotConsumerData = await getJetStreamConsumerData(consumerName, natsApiRuleVSHost);
      });

      it('Compare the consumer creation timestamp', async function() {
        assert.equal(gotConsumerData.consumerName, wantConsumerName);
        assert.equal(
            getTimeStampsWithZeroMilliSeconds(gotConsumerData.consumerCreationTime),
            getTimeStampsWithZeroMilliSeconds(wantConsumerCreationTime),
        );
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

  function jsTestAtLeastOnceDelivery() {
    context('test jetStream at-least once delivery during upgrade', function() {
      const [encoding, eventType, eventSource] = ['binary', 'order.created.v1', 'upgrade'];
      const sink = `http://${eventingUpgradeSinkName}.${testNamespace}.svc.cluster.local`;
      const subscriptions = [
        eventingSubscriptionV1Alpha2(eventType, eventSource,
            sink, subscriptionNames.orderCreatedUpgrade, testNamespace),
      ];
      let eventIdBinary = getRandomEventId(encoding);
      let host; let eventID; let cm;

      before('Check if this an upgrade job and get the eventID if kyma is already upgraded', async function() {
        cm = await getConfigMapWithRetries(testDataConfigMapName, 'default');
        if (!cm || !isUpgradeJob || !isEventingSinkDeployed) {
          debug(`Skipping jetStream at-least once delivery test`);
          this.skip();
        }
        eventID = cm.data.upgradeEventID;
      });

      it('Create upgrade sink', async function() {
        await deployEventingSinkFunction(eventingUpgradeSinkName);
        await waitForEventingSinkFunction(eventingUpgradeSinkName);
      });

      it('ensure upgrade sink is reachable through the api rule', async function() {
        this.test.retries(5);

        host = await getClusterHost(eventingUpgradeSinkName, testNamespace);
        expect(host).to.not.empty;
        debug('host fetched, now checking if eventing-upgrade-sink function is reachable...');
        await checkFunctionReachable(eventingUpgradeSinkName, testNamespace, host);
      });

      it('Check for events sent before upgrade', async function() {
        if (eventID !== undefined) {
          await ensureEventReceivedWithRetry(eventingUpgradeSinkName, clusterHost,
              encoding, eventID, eventType, eventSource);
        }
      });

      it('Create subscriptions during pre-upgrade phase', async function() {
        if (eventID === undefined) {
          await k8sApply(subscriptions);
          await waitForSubscription(subscriptionNames.orderCreatedUpgrade, testNamespace, 'v1alpha2');
        }
      });

      it('Publish events', async function() {
        await publishBinaryCEToEventingSink(clusterHost, eventIdBinary, eventType, eventSource);
      });

      it('Wait for events to be delivered', async function() {
        await ensureEventReceivedWithRetry(eventingUpgradeSinkName, clusterHost,
            encoding, eventIdBinary, eventType, eventSource);
      });

      it('Delete the upgrade sink', async function() {
        await undeployEventingFunction(eventingUpgradeSinkName);
        await checkFunctionUnreachable(eventingUpgradeSinkName, testNamespace, host);
      });

      it('Generate new eventId, save the id and publish events', async function() {
        eventIdBinary = getRandomEventId(encoding);
        await createK8sConfigMap(
            {
              ...cm.data,
              upgradeEventID: eventIdBinary,
            },
            testDataConfigMapName,
        );
        await publishBinaryCEToEventingSink(clusterHost, eventIdBinary, eventType, eventSource);
      });
    });
  }

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

    // Checking subscription consumers are not recreated
    testConsumerNotReCreated();

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

    // Running JetStream At-least Once delivery Test during Upgrade
    jsTestAtLeastOnceDelivery();
  });

  // this is record consumer creation time to compare after the Kyma upgrade
  after('Record order.received.v1 consumer data to ConfigMap', async () => {
    const currentBackend = await getEventingBackend();
    if (currentBackend && currentBackend.toLowerCase() !== natsBackend) {
      debug('Skipping the recording consumer data for non NATS backend!');
      return;
    }

    let testDataConfigMap;
    try {
      testDataConfigMap = await getConfigMap(testDataConfigMapName);
    } catch (err) {
      if (err.statusCode === 404) {
        debug('Skipping the recording consumer data due to missing configmap!');
        return;
      }
      throw err;
    }

    debug('Adding JetStream consumer info to eventing test data configmap');
    const consumerName = await getSubscriptionConsumerName(orderReceivedSubName, testNamespace, subCRDVersion);
    const consumerInfo = await getJetStreamConsumerData(consumerName, natsApiRuleVSHost);
    if (consumerInfo) {
      await createK8sConfigMap(
          {
            ...testDataConfigMap.data,
            ...consumerInfo,
          },
          testDataConfigMapName,
      );
    } else {
      throw Error(`Couldn't add consumer info to the eventing data CM due to` +
          `missing consumer ${consumerName} in NATS JetStream`);
    }
  });

  after('Delete the created APIRule', async function() {
    await deleteApiRule(eventingNatsApiRuleAName, kymaSystem);
  });

  after('Unexpose Grafana', async function() {
    await unexposeGrafana(isSKR);
    this.test.retries(3);
  });
});
