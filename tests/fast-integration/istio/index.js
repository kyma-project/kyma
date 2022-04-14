const {
  ensureIstioConnectivityFixture,
  checkHttpbinOAuthResponse,
  checkHttpbinAllowResponse,
  cleanIstioConnectivityFixture,
} = require('./fixtures');
const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
} = require('../utils');


function istioConnectivityTests() {
  describe('Istio Connectivity Tests:', function() {
    this.timeout(10 * 60 * 1000);
    this.slow(5000);

    let initialRestarts = null;

    it('Listing all pods in cluster', async function() {
      initialRestarts = await getContainerRestartsForAllNamespaces();
    });

    it('Httpbin deployment should be ready', async function() {
      await ensureIstioConnectivityFixture().catch((err) => {
        console.dir(err); // first error is logged
        return ensureIstioConnectivityFixture();
      });
    });

    it('Httpbin call should return 200', async function() {
      await checkHttpbinAllowResponse().catch((err) => {
        console.dir(err); // first error is logged
        return checkHttpbinAllowResponse();
      });
    });

    it('Httpbin OAuth2 call should return 200', async function() {
      await checkHttpbinOAuthResponse().catch((err) => {
        console.dir(err); // first error is logged
        return checkHttpbinOAuthResponse();
      });
    });

    it('Should print report of restarted containers, skipped if no crashes happened', async function() {
      const afterTestRestarts = await getContainerRestartsForAllNamespaces();
      printRestartReport(initialRestarts, afterTestRestarts);
    });

    it('Namespace should be deleted', async function() {
      await cleanIstioConnectivityFixture(false);
    });
  });
}

module.exports = {
  istioConnectivityTests,
};
