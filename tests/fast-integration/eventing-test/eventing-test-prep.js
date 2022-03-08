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
  eventMeshSecretFilePath,
  timeoutTime,
  slowTime,
  gardener,
  director,
  shootName,
  cleanupTestingResources,
} = require('./utils');
const {
  ensureCommerceMockLocalTestFixture,
  setEventMeshSourceNamespace,
  ensureCommerceMockWithCompassTestFixture,
} = require('../test/fixtures/commerce-mock');
const {
  debug,
  createEventingBackendK8sSecret,
} = require('../utils');
const {
  addScenarioInCompass,
  assignRuntimeToScenario,
  scenarioExistsInCompass,
} = require('../compass');


describe('Eventing tests preparation', function() {
  this.timeout(timeoutTime);
  this.slow(slowTime);

  it('Prepare the test assets', async function() {
    // runs once before the first test in this block
    console.log('Running with mockNamespace =', mockNamespace);

    // If eventMeshSecretFilePath is specified then create a k8s secret for eventing-backend
    // else use existing k8s secret as specified in backendK8sSecretName & backendK8sSecretNamespace
    if (eventMeshSecretFilePath) {
      console.log('Creating Event Mesh secret');
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
    console.log('Preparing CommerceMock test fixture on Kyma OSS');
    await ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace).catch((err) => {
      console.dir(err); // first error is logged
      return ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace);
    });
  }

  // prepareAssetsForSKRTests - Sets up CommerceMost for the SKR
  async function prepareAssetsForSKRTests() {
    console.log('Preparing for tests on SKR');

    const compassScenarioAlreadyExist = await scenarioExistsInCompass(director, scenarioName);
    if (compassScenarioAlreadyExist) {
      console.log(`Compass scenario with the name ${scenarioName} already exist, do not register it again`);
    } else {
      await setupCompassScenario();
    }

    await ensureCommerceMockWithCompassTestFixture(
        director,
        appName,
        scenarioName,
        mockNamespace,
        testNamespace,
        false,
        compassScenarioAlreadyExist,
    );
  }

  // setupCompassScenario adds a compass scenario and assigns a runtime to it
  async function setupCompassScenario() {
    // Get shoot info from gardener to get compassID for this shoot
    console.log(`Fetching SKR info for shoot: ${shootName}`);

    const skrInfo = await gardener.getShoot(shootName);
    debug(
        `appName: ${appName},
         scenarioName: ${scenarioName},
         testNamespace: ${testNamespace},
         compassID: ${skrInfo.compassID}`,
    );

    console.log('Assigning SKR to scenario in Compass');
    // Create a new scenario (systems/formations) in compass for this test
    await addScenarioInCompass(director, scenarioName);
    // map scenario to target SKR
    await assignRuntimeToScenario(director, skrInfo.compassID, scenarioName);
  }

  afterEach(async function() {
    // if the test preparation failed, perform the cleanup
    if (this.currentTest.state === 'failed') {
      await cleanupTestingResources();
    }
  });
});
