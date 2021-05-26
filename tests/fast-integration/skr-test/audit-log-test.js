const { 
    Creds,
    OauthClient,
  } = require("../audit-log");
var moment = require('moment');

const {
  ensureCommerceMockLocalTestFixture,
  checkAppGatewayResponse,
  sendEventAndCheckResponse,
  cleanMockTestFixture,
} = require("../test/fixtures/commerce-mock");

const {
  printRestartReport,
  getContainerRestartsForAllNamespaces,
} = require("../utils");

function runTest(name) {
  describe(name, function () {
    for(var i = 1; i<=1; i++){
      console.log("new test run")
      
      this.timeout(10 * 60 * 1000);
      this.slow(5000);
      const testNamespace = "test";
      let initialRestarts = null;

      it("Listing all pods in cluster", async function () {
        initialRestarts = await getContainerRestartsForAllNamespaces();
      });

      it("CommerceMock test fixture should be ready", async function () {
        await ensureCommerceMockLocalTestFixture("mocks", testNamespace).catch((err) => {
          console.dir(err); // first error is logged
          return ensureCommerceMockLocalTestFixture("mocks", testNamespace);
        });
      });

      it("function should reach Commerce mock API through app gateway", async function () {
        await checkAppGatewayResponse();
      });

      it("order.created.v1 event should trigger the lastorder function", async function () {
        await sendEventAndCheckResponse();
      });

      it("Should print report of restarted containers, skipped if no crashes happened", async function () {
        const afterTestRestarts = await getContainerRestartsForAllNamespaces();
        printRestartReport(initialRestarts, afterTestRestarts);
      });

      it("Test namespaces should be deleted", async function () {
        await cleanMockTestFixture("mocks", testNamespace, true);
      });          
    }
  });
}

  const cred = new OauthClient(Creds.fromEnv())
  before(async function() {
    this.timeout(10 * 60 * 1000);
    this.slow(5000);
    await runTest("foo")
  })

  before(async function() {
    await cred.fetchLogs()
  })

    // let timeStampStart = moment().utcOffset(0, false).format();
    // console.log("timestamp start: " + timeStampStart)
    

      

      // let timeStampEnd = moment().utcOffset(0, false).format();
      // console.log("timestamp end: " + timeStampEnd)

      describe("Check for audit logs", function(){
        it(`checks serverless group is logged`, async function() {
          await cred.parseLogs("networking.istio.io", "create")
        })
        it(`checks serverless group is logged`, async function() {
          await cred.parseLogs("rbac.authorization.k8s.io", "create")
        })
        it(`checks serverless group is logged`, async function() {
          await cred.parseLogs("applicationconnector.kyma-project.io", "create")
        })
        it(`checks serverless group is logged`, async function() {
          await cred.parseLogs("applicationconnector.kyma-project.io", "delete")
        })
        it(`checks serverless group is logged`, async function() {
          await cred.parseLogs("applicationconnector.kyma-project.io", "update")
        })
        it(`checks serverless group is logged`, async function() {
          await cred.parseLogs("serverless.kyma-project.io", "create")
        })
        it(`checks serverless group is logged`, async function() {
          await cred.parseLogs("eventing.kyma-project.io", "create")
        })
        it(`checks serverless group is logged`, async function() {
          await cred.parseLogs("eventing.kyma-project.io", "update")
        })
    })