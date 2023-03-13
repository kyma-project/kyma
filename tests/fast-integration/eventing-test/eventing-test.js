const axios = require('axios');
const https = require('https');
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  getEventingBackend,
  waitForNamespace,
  switchEventingBackend,
  debug,
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
  isSKR,
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
  checkEventTracing,
  saveJetStreamDataForRecreateTest,
  jetStreamTestConfigMapName,
  getConfigMapWithRetries,
  checkStreamNotReCreated,
  checkConsumerNotReCreated,
  isUpgradeJob,
  deployEventingSinkFunction,
  waitForEventingSinkFunction,
  deployV1Alpha1Subscriptions,
  deployV1Alpha2Subscriptions,
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

describe('Eventing tests', function() {
  let natsApiRuleVSHost;
  this.timeout(timeoutTime);
  this.slow(slowTime);

  before('Ensure the test namespace exist', async function() {
    await waitForNamespace(testNamespace);
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

  before('Ensure eventing-sink function is ready', async function() {
    try {
      await waitForFunction(eventingSinkName, testNamespace, 30*1000);
    } catch (e) {
      if (!isUpgradeJob) {
        throw e;
      }

      // Deploying eventing sink with subscriptions if its upgrade tests
      // Only temporarily - will be removed
      debug('Eventing Sink is not deployed');
      debug(e);
      await deployEventingSinkFunction();
      await waitForEventingSinkFunction();
      await deployV1Alpha1Subscriptions();
      await deployV1Alpha2Subscriptions();
    }
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
        await checkEventDelivery(clusterHost, 'binary', v1alpha1SubscriptionsTypes[i].type, eventSource, true);
      }
    });

    it('Cloud Events [structured] with v1alpha1 subscription should be delivered to eventing-sink ', async function() {
      let eventSource = 'eventing-test';
      if (backend === bebBackend) {
        eventSource = getEventMeshNamespace();
      }
      for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
        await checkEventDelivery(clusterHost, 'structured', v1alpha1SubscriptionsTypes[i].type, eventSource, true);
      }
    });

    it('Legacy Events with v1alpha1 subscription should be delivered to eventing-sink ', async function() {
      for (let i=0; i < v1alpha1SubscriptionsTypes.length; i++) {
        await checkEventDelivery(clusterHost, 'legacy', v1alpha1SubscriptionsTypes[i].type, 'test', true);
      }
    });

    it('Cloud Events [binary] with v1alpha2 subscription should be delivered to eventing-sink ', async function() {
      for (let i=0; i < subscriptionsTypes.length; i++) {
        await checkEventDelivery(clusterHost, 'binary', subscriptionsTypes[i].type, subscriptionsTypes[i].source);
      }
    });

    it('Cloud Events [structured] with v1alpha2 subscription should be delivered to eventing-sink ', async function() {
      for (let i=0; i < subscriptionsTypes.length; i++) {
        await checkEventDelivery(clusterHost, 'structured', subscriptionsTypes[i].type, subscriptionsTypes[i].source);
      }
    });

    it('Legacy Events with v1alpha2 subscription should be delivered to eventing-sink ', async function() {
      for (let i=0; i < subscriptionsTypes.length; i++) {
        await checkEventDelivery(clusterHost, 'legacy', subscriptionsTypes[i].type, subscriptionsTypes[i].source);
      }
    });

    it('Cloud Events with subscription (exact) should be delivered to eventing-sink ', async function() {
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
      await eventingMonitoringTest(backend, isSKR, true);
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
      await checkEventTracing(clusterHost, subscriptionsTypes[0].type, subscriptionsTypes[0].source, testNamespace);
    });
  }

  function jsSaveStreamAndConsumersDataSuite() {
    context('save stream and consumer data', function() {
      it('saving JetStream data to test stream and consumers will not be re-created by kyma upgrade', async function() {
        if (!isUpgradeJob) {
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
        if (!cm || !isUpgradeJob) {
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

    // Running Eventing tracing tests [v2]
    eventingTracingTestSuiteV2(isSKR);

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

    // Running Eventing end-to-end event delivery tests
    eventDeliveryTestSuite(bebBackend);

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

    // Running Eventing end-to-end event delivery tests
    eventDeliveryTestSuite(natsBackend);

    // Running Eventing tracing tests [v2]
    eventingTracingTestSuiteV2(isSKR);

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
