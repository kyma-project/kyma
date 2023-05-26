const {
  commerceMockTests,
  // gettingStartedGuideTests,
} = require('./');

const {apiExposureTests} = require('../api-exposure');
const {monitoringTests, unexposeGrafana} = require('../monitoring');
const {loggingTests} = require('../logging');
const {createIstioAccessLogResource} = require('../logging/client.js');
const {cleanMockTestFixture} = require('./fixtures/commerce-mock');
const {ensureCommerceMockLocalTestFixture} = require('../test/fixtures/commerce-mock');
const {tracingTests} = require('../tracing');
const {error, sleep} = require('../utils');

describe('Executing Standard Testsuite:', function() {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);

  const mockNamespace = process.env.MOCK_NAMESPACE || 'mocks';
  const testNamespace = 'test';

  before('CommerceMock test fixture should be ready', async function() {
    await sleep(30000);
    await ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace).catch((err) => {
      error(err);
      return ensureCommerceMockLocalTestFixture(mockNamespace, testNamespace);
    });
  });

  before('Istio Accesslog Resource should be deployed', async function() {
    await createIstioAccessLogResource();
  });

  after('Test Cleanup: Test namespaces should be deleted', async function() {
    await cleanMockTestFixture(mockNamespace, testNamespace, true);
  });

  after('Unexpose Grafana', async function() {
    await unexposeGrafana();
  });

  monitoringTests();

  apiExposureTests();
  commerceMockTests(testNamespace);
  // unusuble because of redis dependency that is not usable in the current form after SC migration
  // gettingStartedGuideTests();

  loggingTests();
  tracingTests(testNamespace);
});
