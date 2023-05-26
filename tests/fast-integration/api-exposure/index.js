const {
  ensureApiExposureFixture,
  testHttpbinOAuthResponse,
  testHttpbinAllowResponse,
  testHttpbinOAuthMethod,
  cleanApiExposureFixture,
} = require('./fixtures');

function apiExposureTests() {
  describe('API Exposure Tests:', function() {
    this.timeout(10 * 60 * 1000);
    this.slow(5000);

    it('Httpbin deployment should be ready', async function() {
      await ensureApiExposureFixture().catch((err) => {
        console.dir(err); // first error is logged
        return ensureApiExposureFixture();
      });
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
