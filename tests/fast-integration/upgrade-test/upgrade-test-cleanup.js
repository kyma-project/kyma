const {cleanMockTestFixture} = require('../test/fixtures/commerce-mock');
const {resetGrafanaProxy} = require('../monitoring');

describe('Upgrade test cleanup', function() {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  const testNamespace = 'test';

  it('Test namespaces should be deleted', async function() {
    await cleanMockTestFixture('mocks', testNamespace, true);
  });

  resetGrafanaProxy();
});
