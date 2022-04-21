const {
  commerceMockTests,
  gettingStartedGuideTests,
} = require('./');

const {monitoringTests, resetGrafanaProxy} = require('../monitoring');
const {loggingTests} = require('../logging');
const {tracingTests} = require('../tracing');
const {cleanMockTestFixture} = require('./fixtures/commerce-mock');
const {ensureCommerceMockLocalTestFixture} = require('../test/fixtures/commerce-mock');
const {error} = require('../utils');


describe('Executing Standard Testsuite:', async function() {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);

  const withCentralAppConnectivity = (process.env.WITH_CENTRAL_APP_CONNECTIVITY === 'true');
  const mockNamespace = process.env.MOCK_NAMESPACE || 'mocks';
  const testNamespace = 'test';

  before('CommerceMock test fixture should be ready', async function() {
    await ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace,
        withCentralAppConnectivity).catch((err) => {
      error(err);
      return ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace, withCentralAppConnectivity);
    });
  });

  after('Test Cleanup: Test namespaces should be deleted', async function() {
    await cleanMockTestFixture(mockNamespace, testNamespace, true);
  });

  after('Test Cleanup: Grafana', async function() {
    await resetGrafanaProxy();
  });

  monitoringTests();

  commerceMockTests(testNamespace);
  gettingStartedGuideTests();

  loggingTests();
  tracingTests(mockNamespace, testNamespace);
});
