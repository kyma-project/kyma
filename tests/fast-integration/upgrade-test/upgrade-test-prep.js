const {ensureCommerceMockLocalTestFixture} = require('../test/fixtures/commerce-mock');
const {error} = require('../utils');

describe('Upgrade test preparation', function() {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  const testNamespace = 'test';

  it('CommerceMock test fixture should be ready', async function() {
    await ensureCommerceMockLocalTestFixture('mocks', testNamespace).catch((err) => {
      error(err);
      return ensureCommerceMockLocalTestFixture('mocks', testNamespace);
    });
  });
});
