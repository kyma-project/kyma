const commerceMock = require('../test/fixtures/commerce-mock')
const gettingStartedGuide = require('../test/fixtures/getting-started-guides')
const installer = require('../installer')

const kymaVersion = process.env.INSTALL_KYMA_VERSION || "1.19.1";


describe("Kyma end to end upgrade tests", function () {

  this.timeout(10 * 60 * 1000);
  this.slow(5000);
  const testNamespace = "test"
  const skipComponents = ["dex","tracing","monitoring","console","kiali","logging"];

  it(`Install Kyma ${kymaVersion}`, async function () {
    const resourcesPath = await installer.downloadCharts({ source: kymaVersion })
    await installer.installKyma({ resourcesPath, skipComponents })
  });

  it("CommerceMock test fixture should be ready", async function () {
    await commerceMock.ensureCommerceMockTestFixture("mocks", testNamespace);
  });

  it("Getting started guide fixture should be ready", async function () {
    await gettingStartedGuide.ensureGettingStartedTestFixture()
  });

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

  it("Kyma should be upgraded to Kyma 2.0 (master branch)", async function () {
    await installer.installKyma({isUpgrade: true, skipComponents, newEventing: true});    
  })

  it.skip("Test fixtures should be deleted", async function () {
    await commerceMock.cleanMockTestFixture("mocks", testNamespace)
    await gettingStartedGuide.cleanGettingStartedTestFixture();
  })

});
