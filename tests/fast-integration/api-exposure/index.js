const {
  ensureApiExposureFixture,
  testHttpbinOAuthResponse,
  testHttpbinAllowResponse,
  testHttpbinOAuthMethod,
  cleanApiExposureFixture,
} = require('./fixtures');

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

function apiExposureTests() {
  describe('API Exposure Tests:', function() {
    this.timeout(10 * 60 * 1000);
    this.slow(5000);

    it('Httpbin deployment should be ready', async function() {
      await retryWithDelay((r)=> ensureApiExposureFixture(), defaultRetryDelayMs, defaultRetries);
    });

    it('Testing unsecured httpbin API Rule', async function() {
      await testHttpbinAllowResponse();
    });

    it('Testing secured httpbin API Rule', async function() {
      await testHttpbinOAuthResponse();
    });

    it('Secured httpbin should fail on disallowed method call', async function() {
      await testHttpbinOAuthMethod();
    });

    it('Namespace should be deleted', async function() {
      await cleanApiExposureFixture(false);
    });
  });
}

module.exports = {
  apiExposureTests,
};
