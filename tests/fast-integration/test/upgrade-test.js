const commerceMock = require('./fixtures/commerce-mock')
const gettingStartedGuide = require('./fixtures/getting-started-guides')

const tests = async function () {
  it("function should reach Commerce mock API through app gateway", async function () {
    this.timeout(60 * 1000);
    await commerceMock.checkAppGatewayResponse()
  })

  it("order.created.v1 event should trigger the lastorder function", async function () {
    this.timeout(60 * 1000);
    await commerceMock.sendEventAndCheckResponse();
  });

  it("Getting started Guide", async function () {
    this.timeout(60 * 1000);
    await gettingStartedGuide.verifyOrderPersisted();
  })

}
const kymaVersion = process.env.INSTALL_KYMA_VERSION
const upgradeKyma = process.env.UPGRADE_KYMA;

if (upgradeKyma) {
  describe("Kyma end to end upgrade tests", function () {

    this.timeout(3 * 60 * 1000);
    this.slow(5000);
    const testNamespace = "test"


    describe(`Installing Kyma in version ${kymaVersion}`, function () {
      // Installation code here
    })

    describe("Install fixtures", function () {
      this.timeout(90 * 1000);
      it("CommerceMock test fixture should be ready", async function () {
        await commerceMock.ensureCommerceMockTestFixture("mocks", testNamespace);
      });

      it("Getting started guide fixture should be ready", async function () {
        await gettingStartedGuide.ensureGettingStartedTestFixture()
      });

    });

    describe("Run integration tests", tests);

    describe("Upgrade Kyma", function () {
      // execute upgrade here
    });

    describe("Run integrations tests after upgrade", tests);

    describe("Clean up", async function () {
      this.timeout(2 * 60 * 1000);
      it("Test fixtures should be deleted", async function () {
        await commerceMock.cleanMockTestFixture("mocks",testNamespace)
        await gettingStartedGuide.cleanGettingStartedTestFixture();
      })
    });

  })

}
