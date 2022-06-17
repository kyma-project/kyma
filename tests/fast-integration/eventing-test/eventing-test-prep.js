const axios = require('axios');
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
  backendK8sSecretName,
  backendK8sSecretNamespace,
  timeoutTime,
  slowTime,
  gardener,
  director,
  shootName,
  cleanupTestingResources,
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

  it('Prepare the test assets', async function() {
    // runs once before the first test in this block
    debug('Running with mockNamespace =', mockNamespace);

    // If eventMeshSecretFilePath is specified then create a k8s secret for eventing-backend
    // else use existing k8s secret as specified in backendK8sSecretName & backendK8sSecretNamespace
    if (eventMeshSecretFilePath) {
      debug('Creating Event Mesh secret');
      const eventMeshInfo = await createEventingBackendK8sSecret(
          eventMeshSecretFilePath,
          backendK8sSecretName,
          backendK8sSecretNamespace,
      );
      setEventMeshSourceNamespace(eventMeshInfo['namespace']);
    }

    // Deploy Commerce mock application, function and subscriptions for tests
    if (isSKR) {
      await prepareAssetsForSKRTests();
    } else {
      await prepareAssetsForOSSTests();
    }
  });

  // prepareAssetsForOSSTests - Sets up CommerceMost for the OSS
  async function prepareAssetsForOSSTests() {
    debug('Preparing CommerceMock test fixture on Kyma OSS');
    await ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace).catch((err) => {
      error(err); // first error is logged
      return ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace);
    });
  }

  // prepareAssetsForSKRTests - Sets up CommerceMost for the SKR
  async function prepareAssetsForSKRTests() {
    info('Preparing for tests on SKR');

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
    // Get shoot info from gardener to get compassID for this shoot
    debug(`Fetching SKR info for shoot: ${shootName}`);

    debug('Assigning SKR to scenario in Compass');
    // Create a new scenario (systems/formations) in compass for this test
    await addScenarioInCompass(director, scenarioName);
  }

  afterEach(async function() {
    // if the test preparation failed, perform the cleanup
    if (this.currentTest.state === 'failed') {
      await cleanupTestingResources();
    }
  });
});
