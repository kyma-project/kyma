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
  orderReceivedSubName,
  getRandomEventId,
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
  deleteApiRule,
  k8sApply,
  waitForSubscription,
  eventingSubscriptionV1Alpha2,
  deleteK8sConfigMap,
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
  deployEventingSinkFunction,
  eventingUpgradeSinkName,
  waitForEventingSinkFunction,
  ensureEventReceivedWithRetry,
  undeployEventingFunction,
  checkFunctionUnreachable,
  publishEventWithRetry,
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

  function jsTestAtLeastOnceDeliveryTestSuite(upgradeStage='pre') {
    // The test scenario is:
    // 1. Create a second sink + subscription and wait for them to be healthy.
    // 2. Send an event and verify if it is received by the sink.
    // 3. Delete the sink.
    // 4. Publish an event.
    // 5. Upgrade the Kyma cluster.
    // 6. Revert/Deploy back the second sink.
    // 7. Check if the sink receives the event which was sent before the upgrade.
    context('Test jetStream at-least once delivery during kyma upgrade', function() {
      const encoding = 'binary';
      const subName = `upgrade-${subscriptionsTypes[0].name}`;
      const subscriptions = [
        eventingSubscriptionV1Alpha2(
            subscriptionsTypes[0].type,
            subscriptionsTypes[0].source,
            `http://${eventingUpgradeSinkName}.${testNamespace}.svc.cluster.local`,
            `upgrade-${subscriptionsTypes[0].name}`,
            testNamespace),
      ];

      before('Check if this is an upgrade job', async function() {
        if (!isUpgradeJob || !isEventingSinkDeployed) {
          debug(`Skipping jetStream at-least once delivery test isUpgradeJob: ${isUpgradeJob}`);
          this.skip();
        }
      });

      context('Pre-upgrade tasks to publish event which should not be delivered', function() {
        before('Check if this is pre-upgrade stage', async function() {
          if (upgradeStage !== 'pre') {
            debug(`Skipping pre-upgrade tasks...`);
            this.skip();
          }
        });

        it('Deploy eventing-upgrade-sink', async function() {
          await deployEventingSinkFunction(eventingUpgradeSinkName);
          await waitForEventingSinkFunction(eventingUpgradeSinkName);
          debug(`checking if eventing upgrade sink is reachable through the api rule`);
          await checkFunctionReachable(eventingUpgradeSinkName, testNamespace, clusterHost);
        });

        it('Create subscriptions during pre-upgrade phase', async function() {
          await k8sApply(subscriptions, testNamespace);
          await waitForSubscription(subName, testNamespace, 'v1alpha2');
        });

        it('Verify if events delivery is working', async function() {
          await checkEventDelivery(clusterHost, 'binary', subscriptionsTypes[0].type,
              subscriptionsTypes[0].source, false, eventingUpgradeSinkName);
        });

        it('Delete the eventing-upgrade-sink', async function() {
          await undeployEventingFunction(eventingUpgradeSinkName);
          debug(`checking if eventing upgrade sink is not alive anymore...`);
          await checkFunctionUnreachable(eventingUpgradeSinkName, testNamespace, clusterHost);
        });

        it('Generate new eventId, save the id and publish events', async function() {
          let existingCMData = {};
          // get existing configMap
          const cm = await getConfigMapWithRetries(testDataConfigMapName, testNamespace);
          if (cm && cm.data) {
            existingCMData = cm.data;
          }

          const eventIdBinary = getRandomEventId(encoding);
          await createK8sConfigMap(
              {
                ...existingCMData,
                upgradeEventID: eventIdBinary,
              },
              testDataConfigMapName,
          );
          // publish the event, which should be delivered after the upgrade
          await publishEventWithRetry(clusterHost, encoding, eventIdBinary,
              subscriptionsTypes[0].type, subscriptionsTypes[0].source);
        });
      });

      context('Post-upgrade tasks to verify that event is delivered after sink is available', function() {
        let eventID;
        before('Check if this is pre-upgrade stage', async function() {
          if (upgradeStage !== 'post') {
            debug(`Skipping post-upgrade tasks...`);
            this.skip();
          }

          const cm = await getConfigMapWithRetries(testDataConfigMapName, 'default');
          if (!cm || !cm.data || !cm.data.upgradeEventID) {
            debug(`Skipping post-upgrade tasks because config map is not configured...`);
            this.skip();
          }
          eventID = cm.data.upgradeEventID;
        });

        it(`deploy again and wait for function: ${eventingUpgradeSinkName}`, async function() {
          await deployEventingSinkFunction(eventingUpgradeSinkName);
          await waitForEventingSinkFunction(eventingUpgradeSinkName);
          debug(`checking if eventing upgrade sink is reachable through the api rule`);
          await checkFunctionReachable(eventingUpgradeSinkName, testNamespace, clusterHost);
        });

        it('Label subscription to trigger reconciliation', async function() {
          subscriptions[0].metadata.labels = {
            'changed': 'now',
          };
          await k8sApply(subscriptions, testNamespace);
          await waitForSubscription(subName, testNamespace, 'v1alpha2');
        });

        it('Wait for the pending event to be delivered', async function() {
          await ensureEventReceivedWithRetry(eventingSinkName, clusterHost,
              encoding, eventID, subscriptionsTypes[0].type, subscriptionsTypes[0].source);
        });

        it(`Delete configMap: ${testDataConfigMapName}`, async function() {
          await deleteK8sConfigMap(testDataConfigMapName);
        });
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

    // Running JetStream At-least Once delivery Test during Upgrade
    jsTestAtLeastOnceDeliveryTestSuite('post');
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
    jsTestAtLeastOnceDeliveryTestSuite('pre');
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
