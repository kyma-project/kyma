const {ensureCommerceMockLocalTestFixture} = require('../test/fixtures/commerce-mock');
const defaultRetryDelayMs = 1000;
const defaultRetries = 5;
const retryWithDelay = (operation, delay, retries) => new Promise((resolve, reject) => {
  return operation()
      .then(resolve)
      .catch((reason) => {
        if (retries > 0) {
          return wait(delay)
              .then(retryWithDelay.bind(null, operation, delay, retries - 1))
              .then(resolve)
              .catch(reject);
        }
        return reject(reason);
      });
});
describe('Upgrade test preparation', function() {
  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  const testNamespace = 'test';

  it('CommerceMock test fixture should be ready', async function() {
    await retryWithDelay( (r) => ensureCommerceMockLocalTestFixture('mocks', testNamespace),
        defaultRetryDelayMs, defaultRetries);
  });
});
