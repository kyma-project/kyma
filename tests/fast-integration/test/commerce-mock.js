const { ensureCommerceMockTestFixture,
  checkAppGatewayResponse,
  sendEventAndCheckResponse,
  cleanMockTestFixture
} = require('./fixtures/commerce-mock')

describe("CommerceMock tests", function () {

  this.timeout(2 * 60 * 1000);
  this.slow(5000);
  const testNamespace = "test"

  it("CommerceMock test fixture should be ready", async function () {
    this.timeout(4 * 60 * 1000);
    await ensureCommerceMockTestFixture("mocks", testNamespace);
  });

  it("function should reach Commerce mock API through app gateway", async function () {
    await checkAppGatewayResponse()
  })

  it("order.created.v1 event should trigger the lastorder function", async function () {
    await sendEventAndCheckResponse();
  });

  it("Test namespaces should be deleted", async function () {
    await cleanMockTestFixture("mocks", testNamespace, false);
  });

})
