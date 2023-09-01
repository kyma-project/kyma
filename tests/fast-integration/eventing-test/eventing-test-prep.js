const axios = require('axios');
const https = require('https');
const fs = require('fs');
const path = require('path');
const k8s = require('@kubernetes/client-node');

const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  testNamespace,
  kymaVersion,
  isSKR,
  skrInstanceId,
  backendK8sSecretName,
  backendK8sSecretNamespace,
  timeoutTime,
  slowTime,
  shootName,
  cleanupTestingResources,
  eventingSinkName,
  getClusterHost,
  checkFunctionReachable,
  deployEventingSinkFunction,
  waitForEventingSinkFunction,
  deployV1Alpha1Subscriptions,
  deployV1Alpha2Subscriptions,
  createK8sNamespace,
  isUpgradeJob,
} = require('./utils');
const {
  eventMeshSecretFilePath,
} = require('./common/common');
const {
  setEventMeshSourceNamespace,
} = require('../test/fixtures/commerce-mock');
const {
  info,
  debug,
  createEventingBackendK8sSecret,
  deployJaeger,
} = require('../utils');
const {expect} = require('chai');

const jaegerYaml = fs.readFileSync(
    path.join(__dirname, '../test/fixtures/jaeger/jaeger.yaml'),
    {
      encoding: 'utf8',
    },
);


describe('Eventing tests preparation', function() {
  this.timeout(timeoutTime);
  this.slow(slowTime);

  it('Print test initial configs', async function() {
    debug(`Test namespace: ${testNamespace}`);
    debug(`Kyma version: ${kymaVersion}`);
    debug(`Is SKR cluster: ${isSKR}`);
    debug(`Is upgrade job: ${isUpgradeJob}`);
    debug(`SKR instance Id: ${skrInstanceId}`);
    debug(`SKR shoot name: ${shootName}`);
  });

  it('Prepare SKR Kubeconfig if needed', async function() {
    // Skip this step if it is not a SKR cluster
    if (!isSKR) {
      this.skip();
    }

    if (!skrInstanceId) {
      info(`Skipping fetching SKR kubeconfig because skrInstanceId is not set.`);
      this.skip();
    }

    // 'skr-test/helpers' initializes KEB clients on import, that is why it is imported only if needed
    const {getSKRConfig} = require('./skr-helpers/helpers');
    const {initK8sConfig} = require('./skr-helpers/helpers');

    debug(`Fetching SKR config for Instance Id: ${skrInstanceId}`);
    const shoot = await getSKRConfig(skrInstanceId);

    debug('Initiating SKR K8s config...');
    await initK8sConfig(shoot);
  });

  it('Prepare EventMesh secret', async function() {
    // If eventMeshSecretFilePath is specified then create a k8s secret for eventing-backend
    // else skip this step and use existing k8s secret as specified in backendK8sSecretName & backendK8sSecretNamespace
    // For upgrade tests we do not need eventMesh; all tests are done with NATS.
    if (isUpgradeJob || !eventMeshSecretFilePath) {
      this.skip();
    }

    debug('Creating Event Mesh secret');
    const eventMeshInfo = await createEventingBackendK8sSecret(
        eventMeshSecretFilePath,
        backendK8sSecretName,
        backendK8sSecretNamespace,
    );
    setEventMeshSourceNamespace(eventMeshInfo['namespace']);
  });

  it('Create test namespace', async function() {
    await createK8sNamespace(testNamespace);
  });

  it('Prepare eventing-sink function', async function() {
    debug('Preparing EventingSinkFunction');
    await deployEventingSinkFunction(eventingSinkName);
    await waitForEventingSinkFunction(eventingSinkName);
  });

  it('Eventing-sink function should be reachable through API Rule', async function() {
    const host = await getClusterHost(eventingSinkName, testNamespace);
    expect(host).to.not.empty;
    debug('host fetched, now checking if eventing-sink function is reachable...');
    await checkFunctionReachable(eventingSinkName, testNamespace, host);
  });

  it('Prepare v1alpha1 subscriptions', async function() {
    await deployV1Alpha1Subscriptions();
  });

  it('Prepare v1alpha2 subscriptions', async function() {
    await deployV1Alpha2Subscriptions();
  });

  it('Should deploy jaeger', async function() {
    if (isSKR || isUpgradeJob) {
      this.skip();
    }
    await deployJaeger(k8s.loadAllYaml(jaegerYaml));
  });

  // afterEach(async function() {
  // if the test preparation failed, perform the cleanup
  // if (this.currentTest.state === 'failed') {
  //    await cleanupTestingResources();
  //  }
  //  });
});
