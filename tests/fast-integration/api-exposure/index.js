const {
  ensureApiExposureFixture,
  testHttpbinOAuthResponse,
  testHttpbinAllowResponse,
  cleanApiExposureFixture,
} = require('./fixtures');
const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
} = require('../utils');


function apiExposureTests() {
  describe('API Exposure Tests:', function() {
    this.timeout(10 * 60 * 1000);
    this.slow(5000);

    let initialRestarts = null;

    it('Listing all pods in cluster', async function() {
      initialRestarts = await getContainerRestartsForAllNamespaces();
    });

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

    it('Should print report of restarted containers, skipped if no crashes happened', async function() {
      const afterTestRestarts = await getContainerRestartsForAllNamespaces();
      printRestartReport(initialRestarts, afterTestRestarts);
    });

    it('Namespace should be deleted', async function() {
      await cleanApiExposureFixture(false);
    });
  });
}

module.exports = {
  apiExposureTests,
};
