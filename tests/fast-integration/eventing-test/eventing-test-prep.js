const axios = require('axios');
const {expect} = require('chai');
const https = require('https');
const httpsAgent = new https.Agent({
  rejectUnauthorized: false, // curl -k
});
axios.defaults.httpsAgent = httpsAgent;
const {
  appName,
  scenarioName,
  testNamespace,
  mockNamespace,
  isSKR,
  testCompassFlow,
  skrInstanceId,
  backendK8sSecretName,
  backendK8sSecretNamespace,
  timeoutTime,
  slowTime,
  gardener,
  director,
  shootName,
  cleanupTestingResources,
  updateShootName,
} = require('./utils');
const {eventMeshSecretFilePath} = require('./common/common');
const {
  ensureCommerceMockLocalTestFixture,
  setEventMeshSourceNamespace,
  ensureCommerceMockWithCompassTestFixture,
} = require('../test/fixtures/commerce-mock');
const {
  info,
  error,
  debug,
  createEventingBackendK8sSecret,
  isK8sClientInitialized,
} = require('../utils');
const {
  addScenarioInCompass,
  assignRuntimeToScenario,
  scenarioExistsInCompass,
  isRuntimeAssignedToScenario,
} = require('../compass');

describe('Eventing tests preparation', function() {
  this.timeout(timeoutTime);
  this.slow(slowTime);

  it('Print test initial configs', async function() {
    debug(`Mock namespace: ${mockNamespace}`);
    debug(`Test namespace: ${testNamespace}`);
    debug(`Is SKR cluster: ${isSKR}`);
    debug(`SKR instance Id: ${skrInstanceId}`);
    debug(`Test Compass flow enabled: ${testCompassFlow}`);
  });

  it('Prepare EventMesh secret', async function() {
    // If eventMeshSecretFilePath is specified then create a k8s secret for eventing-backend
    // else skip this step and use existing k8s secret as specified in backendK8sSecretName & backendK8sSecretNamespace
    if (!eventMeshSecretFilePath) {
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

  it('Prepare SKR Kubeconfig if needed', async function() {
    // Skip this step if it is not a SKR cluster
    if (!isSKR) {
      this.skip();
    }

    if (isK8sClientInitialized()) {
      info(`Skipping fetching SKR kubeconfig because k8s client is already initialized.`);
      this.skip();
    }

    // check if skrInstanceId is provided
    expect(skrInstanceId).to.not.be.empty;

    // 'skr-test/helpers' initializes KEB clients on import, that is why it is imported only if needed
    const {getSKRConfig} = require('../skr-test/helpers');
    const {initK8sConfig} = require('../skr-test/helpers');

    debug(`Fetching SKR config for Instance Id: ${skrInstanceId}`);
    const shoot = await getSKRConfig(skrInstanceId);

    debug('Initiating SKR K8s config...');
    await initK8sConfig(shoot);

    debug(`Setting shoot name to: ${shoot.name}`);
    updateShootName(shoot.name);
  });

  it('Prepare assets without Compass flow', async function() {
    // Skip this step if compass flow is enabled
    if (testCompassFlow) {
      this.skip();
    }

    // Deploy Commerce mock application, function and subscriptions for tests
    await prepareAssetsForOSSTests();
  });

  it('Prepare assets with Compass flow', async function() {
    // Skip this step if compass flow is disabled
    if (!testCompassFlow) {
      this.skip();
    }

    // Deploy Commerce mock application, function and subscriptions for tests (includes compass flow)
    await prepareAssetsForSKRTests();
  });

  afterEach(async function() {
    // if the test preparation failed, perform the cleanup
    if (this.currentTest.state === 'failed') {
      await cleanupTestingResources();
    }
  });

  // // **** Helper functions ****
  // prepareAssetsForOSSTests - Sets up CommerceMost for the OSS
  async function prepareAssetsForOSSTests() {
    debug('Preparing CommerceMock/In-cluster test fixtures on Kyma');
    await ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace).catch((err) => {
      error(err); // first error is logged
      return ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace);
    });
  }

  // prepareAssetsForSKRTests - Sets up CommerceMost for the SKR
  async function prepareAssetsForSKRTests() {
    debug('Preparing CommerceMock/In-cluster test fixtures with compass flow on SKR');

    const skrInfo = await gardener.getShoot(shootName);

    debug(
        `appName: ${appName},
         scenarioName: ${scenarioName},
         testNamespace: ${testNamespace},
         compassID: ${skrInfo.compassID}`,
    );

    // check if compass scenario setup is needed
    const compassScenarioAlreadyExist = await scenarioExistsInCompass(director, scenarioName);
    if (compassScenarioAlreadyExist) {
      debug(`Compass scenario with the name ${scenarioName} already exist, do not register it again`);
    } else {
      await setupCompassScenario();
    }

    // check if assigning the runtime to the scenario is needed
    const runtimeAssignedToScenario = await isRuntimeAssignedToScenario(director, skrInfo.compassID, scenarioName);
    if (!runtimeAssignedToScenario) {
      debug('Assigning Runtime to a compass scenario');
      // map scenario to target SKR
      await assignRuntimeToScenario(director, skrInfo.compassID, scenarioName);
    }

    await ensureCommerceMockWithCompassTestFixture(
        director,
        appName,
        scenarioName,
        mockNamespace,
        testNamespace,
        compassScenarioAlreadyExist,
    );
  }

  // setupCompassScenario adds a compass scenario
  async function setupCompassScenario() {
    debug('Assigning SKR to scenario in Compass');
    // Create a new scenario (systems/formations) in compass for this test
    await addScenarioInCompass(director, scenarioName);
  }
});
