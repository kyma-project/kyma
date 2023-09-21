const axios = require('axios');
const https = require('https');
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  getRandomEventId,
} = require('../test/fixtures/commerce-mock');
const {
  getEventingBackend,
  waitForNamespace,
  switchEventingBackend,
  debug,
  createK8sConfigMap,
  createApiRuleForService,
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
  testDataConfigMapName,
  eventingNatsSvcName,
  eventingNatsApiRuleAName,
  eventingSinkName,
  v1alpha1SubscriptionsTypes,
  subscriptionsTypes,
  subscriptionsExactTypeMatching,
  getClusterHost,
  checkFunctionReachable,
  checkEventDelivery,
  waitForV1Alpha1Subscriptions,
  waitForV1Alpha2Subscriptions,
  saveJetStreamDataForRecreateTest,
  jsRecreatedTestConfigMapName,
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
  debugBanner,
  isSKR,
  isUpgradeJob,
  isJSRecreatedTestEnabled,
  isJSAtLeastOnceDeliveryTestEnabled,
} = require('./utils');
const {
  bebBackend,
  natsBackend,
  getEventMeshNamespace,
  kymaSystem,
} = require('./common/common');
const {
  expect,
} = require('chai');
const {
  exposeGrafana,
  unexposeGrafana,
} = require('../monitoring');

let clusterHost = '';
let isJSAtLeastOnceTested = false;

describe('Eventing tests', function() {
  let natsApiRuleVSHost;
  this.timeout(timeoutTime);
  this.slow(slowTime);

  before('Ensure the test namespace exist', async function() {
    await waitForNamespace(testNamespace);
  });

  before('Expose Grafana', async function() {
    if (isUpgradeJob) {
      return;
    }
    await exposeGrafana();
    this.test.retries(3);
  });

  before('Create an ApiRule for NATS', async () => {
    if (!isUpgradeJob) {
      debug(`Skipping creating ApiRule for NATS because it is not upgrade test job`);
      return;
    }

    const vs = await createApiRuleForService(eventingNatsApiRuleAName,
        kymaSystem,
        eventingNatsSvcName,
        8222);
    natsApiRuleVSHost = vs.spec.hosts[0];
  });

  before('Ensure eventing-sink function is ready', async function() {
    await waitForEventingSinkFunction(eventingSinkName);
  });

  before('Get cluster host name from Virtual Services', async function() {
    clusterHost = await getClusterHost(eventingSinkName, testNamespace);
    expect(clusterHost).to.not.empty;
    debug(`host name fetched: ${clusterHost}`);
  });

  // eventDeliveryTestSuite - Runs Eventing tests for event delivery
  function eventDeliveryTestSuite(backend) {
    it('Wait for subscriptions to be ready', async function() {
      // waiting for v1alpha1 subscriptions
      await waitForV1Alpha1Subscriptions();

      // waiting for v1alpha2 subscriptions
      await waitForV1Alpha2Subscriptions();
    });

    it('Eventing-sink function should be reachable through API Rule', async function() {
      await checkFunctionReachable(eventingSinkName, testNamespace, clusterHost);
    });

    it('Cloud Events [binary] with v1alpha1 subscription should be delivered to eventing-sink ', async function() {
      let eventSource = 'eventing-test';
      if (backend === bebBackend) {
        eventSource = getEventMeshNamespace();
      }
      for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
        debugBanner(`Testing Cloud Events [binary] [v1alpha1] ${v1alpha1SubscriptionsTypes[i].type}`);
        await checkEventDelivery(clusterHost, 'binary', v1alpha1SubscriptionsTypes[i].type, eventSource, true);
      }
    });

    it('Cloud Events [structured] with v1alpha1 subscription should be delivered to eventing-sink ', async function() {
      let eventSource = 'eventing-test';
      if (backend === bebBackend) {
        eventSource = getEventMeshNamespace();
      }
      for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
        debugBanner(`Testing Cloud Events [structured] [v1alpha1] ${v1alpha1SubscriptionsTypes[i].type}`);
        await checkEventDelivery(clusterHost, 'structured', v1alpha1SubscriptionsTypes[i].type, eventSource, true);
      }
    });

    it('Legacy Events with v1alpha1 subscription should be delivered to eventing-sink ', async function() {
      for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
        debugBanner(`Testing Cloud Events [Legacy] [v1alpha1] ${v1alpha1SubscriptionsTypes[i].type}`);
        await checkEventDelivery(clusterHost, 'legacy', v1alpha1SubscriptionsTypes[i].type, 'test', true);
      }
    });

    it('Cloud Events [binary] with v1alpha2 subscription should be delivered to eventing-sink ', async function() {
      for (let i=0; i < subscriptionsTypes.length; i++) {
        debugBanner(`Testing Cloud Events [binary] [v1alpha2] ${subscriptionsTypes[i].type}`);
        await checkEventDelivery(clusterHost, 'binary', subscriptionsTypes[i].type, subscriptionsTypes[i].source);
      }
    });

    it('Cloud Events [structured] with v1alpha2 subscription should be delivered to eventing-sink ', async function() {
      for (let i=0; i < subscriptionsTypes.length; i++) {
        debugBanner(`Testing Cloud Events [structured] [v1alpha2] ${subscriptionsTypes[i].type}`);
        await checkEventDelivery(clusterHost, 'structured', subscriptionsTypes[i].type, subscriptionsTypes[i].source);
      }
    });

    it('Legacy Events with v1alpha2 subscription should be delivered to eventing-sink ', async function() {
      for (let i=0; i < subscriptionsTypes.length; i++) {
        debugBanner(`Testing Cloud Events [legacy] [v1alpha2] ${subscriptionsTypes[i].type}`);
        await checkEventDelivery(clusterHost, 'legacy', subscriptionsTypes[i].type, subscriptionsTypes[i].source);
      }
    });

    it('Cloud Events with subscription (exact) should be delivered to eventing-sink ', async function() {
      let eventSource = 'eventing-test';
      if (backend === bebBackend) {
        eventSource = getEventMeshNamespace();
      }
      for (let i=0; i < subscriptionsExactTypeMatching.length; i++) {
        debugBanner(`Testing Cloud Events (type matching: exact) ${subscriptionsExactTypeMatching[i].type}`);
        await checkEventDelivery(clusterHost, 'binary', subscriptionsExactTypeMatching[i].type, eventSource);
      }
    });
  }

  // eventingMonitoringTestSuite - Runs Eventing tests for monitoring
  function eventingMonitoringTestSuite(backend, isSKR, isUpgradeJob=true) {
    if (isUpgradeJob) {
      return;
    }
    it('Run Eventing Monitoring tests', async function() {
      await eventingMonitoringTest(backend, isSKR, true);
    });
  }

  function jsTestStreamConsumerNotRecreatedTestSuite(upgradeStage='pre') {
    // The test scenario is:
    // 1. Before upgrade, save the stream and consumer creation timestamp in a configMap.
    // 2. Upgrade the Kyma cluster.
    // 3. After upgrade, fetch again the stream and consumer creation timestamp from NATS.
    // 4. Compares the stream and consumer creation timestamp before and after upgrade.
    // 5. If the timestamps are same, we conclude that the stream/consumer is not re-created.
    // Note: It only checks consumer of first subscription (i.e. subscriptionsTypes[0]).
    context('Test jetStream stream and consumer not re-created during kyma upgrade', function() {
      before('Check if this is an upgrade job and test enabled', async function() {
        if (!isUpgradeJob || !isJSRecreatedTestEnabled) {
          debug(`Skipping jetStream stream and consumer not re-created test`);
          debug(`isUpgradeJob: ${isUpgradeJob}, isJSRecreatedTestEnabled: ${isJSRecreatedTestEnabled}`);
          this.skip();
        }
      });

      context('Pre-upgrade tasks to save stream and consumer data', function() {
        before('Check if this is pre-upgrade stage', async function() {
          if (upgradeStage !== 'pre') {
            debug(`Skipping pre-upgrade tasks...`);
            this.skip();
          }
          debugBanner('Pre-upgrade tasks to save stream and consumer data');
        });

        it('saving JetStream data to test stream and consumers will not be re-created ' +
            'by kyma upgrade', async function() {
          debug(`Using NATS host: ${natsApiRuleVSHost}`);
          await saveJetStreamDataForRecreateTest(natsApiRuleVSHost, jsRecreatedTestConfigMapName);
        });
      });

      context('Post-upgrade tasks to verify that stream and consumer not re-created ' +
          'check with NATS backend', function() {
        let preUpgradeStreamData = null;
        let preUpgradeConsumersData = null;

        before('Check if this is post-upgrade stage', async function() {
          if (upgradeStage !== 'post') {
            debug(`Skipping post-upgrade tasks...`);
            this.skip();
          }
          debugBanner('Post-upgrade tasks to verify that stream and consumer not re-created check with NATS backend');
        });

        before('fetch data stream and consumer data from configMap', async function() {
          debug(`fetch configMap: ${jsRecreatedTestConfigMapName}...`);
          const cm = await getConfigMapWithRetries(jsRecreatedTestConfigMapName, testNamespace);
          if (!cm) {
            // TODO: Remove this once these changes are migrated to release-branch
            debug(`Skipping stream and consumers not re-created check because configMap was not found!`);
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

      before('Check if this is an upgrade job and test enabled', async function() {
        if (!isUpgradeJob || !isJSAtLeastOnceDeliveryTestEnabled) {
          debug(`Skipping jetStream at-least once delivery during upgrade test`);
          debug(`isUpgradeJob: ${isUpgradeJob}, 
          isJSAtLeastOnceDeliveryTestEnabled: ${isJSAtLeastOnceDeliveryTestEnabled}`);
          this.skip();
        }
      });

      context('Pre-upgrade tasks to publish event which should not be delivered', function() {
        before('Check if this is pre-upgrade stage', async function() {
          if (upgradeStage !== 'pre' || isJSAtLeastOnceTested) {
            debug(`Skipping pre-upgrade tasks...`);
            this.skip();
          }
          debugBanner('Pre-upgrade tasks to publish event which should not be delivered');
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
          debugBanner('Post-upgrade tasks to verify that event is delivered after sink is available');
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
          await ensureEventReceivedWithRetry(eventingUpgradeSinkName, clusterHost,
              encoding, eventID, subscriptionsTypes[0].type, subscriptionsTypes[0].source, 50);
          // change the flag to true we do not prepare test again
          isJSAtLeastOnceTested = true;
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

    // Running Eventing end-to-end event delivery tests
    eventDeliveryTestSuite(natsBackend);

    // Running Eventing monitoring tests.
    eventingMonitoringTestSuite(natsBackend, isSKR, isUpgradeJob);

    // Running JetStream stream and consumers not re-created by upgrade test.
    jsTestStreamConsumerNotRecreatedTestSuite('post');

    // Running JetStream At-least Once delivery Test during Upgrade
    jsTestAtLeastOnceDeliveryTestSuite('post');
  });

  context('with BEB backend', function() {
    // skip backend-switching in upgrade test
    if (isUpgradeJob) {
      debug('Skipping backend switching for upgrade test.');
      return;
    }
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

    // Running Eventing end-to-end event delivery tests
    eventDeliveryTestSuite(bebBackend);

    // Running Eventing monitoring tests.
    eventingMonitoringTestSuite(bebBackend, isSKR, isUpgradeJob);
  });

  context('with Nats backend switched back from BEB', async function() {
    // skip backend-switching in upgrade test
    if (isUpgradeJob) {
      debug('Skipping backend switching for upgrade test.');
      return;
    }
    it('Switch Eventing Backend to Nats', async function() {
      const currentBackend = await getEventingBackend();
      if (currentBackend && currentBackend.toLowerCase() === natsBackend) {
        this.skip();
      }
      await switchEventingBackend(backendK8sSecretName, backendK8sSecretNamespace, natsBackend);
    });

    // Running Eventing end-to-end event delivery tests
    eventDeliveryTestSuite(natsBackend);

    // Running Eventing monitoring tests.
    eventingMonitoringTestSuite(natsBackend, isSKR, isUpgradeJob);

    // Running stream and consumer not re-created by upgrade test
    jsTestStreamConsumerNotRecreatedTestSuite('pre');

    // Running JetStream At-least Once delivery Test during Upgrade
    jsTestAtLeastOnceDeliveryTestSuite('pre');
  });

  after('Delete the created APIRule', async function() {
    if (!isUpgradeJob) {
      debug(`Skipping deleting ApiRule for NATS because it is not upgrade test job`);
      return;
    }

    await deleteApiRule(eventingNatsApiRuleAName, kymaSystem);
  });

  after('Unexpose Grafana', async function() {
    if (isUpgradeJob) {
      return;
    }
    await unexposeGrafana(isSKR);
    this.test.retries(3);
  });
});
