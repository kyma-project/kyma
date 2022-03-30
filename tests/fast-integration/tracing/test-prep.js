const {
  testNamespace,
  mockNamespace,
  backendK8sSecretName,
  backendK8sSecretNamespace,
  eventMeshSecretFilePath,
  timeoutTime,
  slowTime,
  cleanupTestingResources,
} = require('./utils');
const {
  ensureCommerceMockLocalTestFixture,
  setEventMeshSourceNamespace,
} = require('../test/fixtures/commerce-mock');
const {createEventingBackendK8sSecret} = require('../utils');

async function testPrep() {
  describe('Eventing tests preparation', function() {
    this.timeout(timeoutTime);
    this.slow(slowTime);

    it('Prepare the test assets', async function() {
      // run once before the first test in this block
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
      // await prepareAssetsForOSSTests(); TODO
    });

    // prepareAssetsForOSSTests - Set up CommerceMock for the OSS TODO this is already done in upgrade-test-prep.js
    // async function prepareAssetsForOSSTests() {
    //   console.log('Preparing CommerceMock test fixture on Kyma OSS');
    //   await ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace).catch((err) => {
    //     console.dir(err);
    //     return ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace);
    //   });
    // }

    afterEach(async function() {
      // if the test preparation failed, perform the cleanup
      if (this.currentTest.state === 'failed') {
        await cleanupTestingResources();
      }
    });
  });
}

module.exports = {
  testPrep,
};
