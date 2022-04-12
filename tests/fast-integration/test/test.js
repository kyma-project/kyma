// const {
//   commerceMockTests,
//   gettingStartedGuideTests,
// } = require('./');

const {monitoringTests} = require('../monitoring');
// const {loggingTests} = require('../logging');
// const {tracingTests} = require('../tracing');
// const {cleanMockTestFixture} = require('./fixtures/commerce-mock');
// const {ensureCommerceMockLocalTestFixture} = require('../test/fixtures/commerce-mock');


describe('Executing Standard Testsuite:', function() {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);

  // const withCentralAppConnectivity = (process.env.WITH_CENTRAL_APP_CONNECTIVITY === 'true');
  // const mockNamespace = process.env.MOCK_NAMESPACE || 'mocks';
  // const testNamespace = 'test';
  //
  // before('CommerceMock test fixture should be ready', async function() {
  //   await ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace, withCentralAppConnectivity).catch((err) => {
  //     console.dir(err); // first error is logged
  //     return ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace, withCentralAppConnectivity);
  //   });
  // });
  //
  // after('Test Cleanup: Test namespaces should be deleted', async function() {
  //   await cleanMockTestFixture(mockNamespace, testNamespace, true);
  // });

  // commerceMockTests(testNamespace);
  // gettingStartedGuideTests();

  monitoringTests();
  // loggingTests();
  // tracingTests(mockNamespace, testNamespace);
});
