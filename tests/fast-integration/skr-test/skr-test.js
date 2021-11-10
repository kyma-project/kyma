const {
  updateSKR,
  ensureValidShootOIDCConfig,
  ensureValidOIDCConfigInCustomerFacingKubeconfig,
} = require("../kyma-environment-broker");
const {
  ensureCommerceMockWithCompassTestFixture,
  checkFunctionResponse,
  sendEventAndCheckResponse,
  deleteMockTestFixture,
} = require("../test/fixtures/commerce-mock");
const {
  ensureKymaAdminBindingExistsForUser,
  ensureKymaAdminBindingDoesNotExistsForUser,
} = require("../utils");
const {
  AuditLogCreds,
  AuditLogClient,
  checkAuditLogs,
  checkAuditEventsThreshold,
} = require("../audit-log");
const {keb, gardener, director} = require('./helpers');
const {prometheusPortForward} = require("../monitoring/client");

function OIDCE2ETest(shoot, runtimeID, options) {
  describe('OIDC E2E Test', function () {
    const oidc0 = options.oidc0;
    const oidc1 = options.oidc1;

    const administrator0 = options.administrator0;
    const administrators1 = options.administrators1;

    it(`Assure initial OIDC config is applied on shoot cluster`, async function () {
      ensureValidShootOIDCConfig(shoot, oidc0);
    });

    it(`Assure initial OIDC config is part of kubeconfig`, async function () {
      await ensureValidOIDCConfigInCustomerFacingKubeconfig(keb, runtimeID, oidc0);
    });

    it(`Assure initial cluster admin`, async function () {
      await ensureKymaAdminBindingExistsForUser(administrator0);
    });

    it(`Update SKR service instance with OIDC config`, async function () {
      const customParams = {
        oidc: oidc1,
      };

      skr = await updateSKR(keb, gardener, runtimeID, shoot.name, customParams);
    });

    it(`Assure updated OIDC config is applied on shoot cluster`, async function () {
      ensureValidShootOIDCConfig(shoot, oidc1);
    });

    it(`Assure updated OIDC config is part of kubeconfig`, async function () {
      await ensureValidOIDCConfigInCustomerFacingKubeconfig(keb, runtimeID, oidc1);
    });

    it(`Assure cluster admin is preserved`, async function () {
      await ensureKymaAdminBindingExistsForUser(administrator0);
    });

    it(`Update SKR service instance with new admins`, async function () {
      const customParams = {
        administrators: administrators1,
      };

      await updateSKR(keb, gardener, runtimeID, shoot.name, customParams);
    });

    it(`Assure only new cluster admins are configured`, async function () {
      await ensureKymaAdminBindingExistsForUser(administrators1[0]);
      await ensureKymaAdminBindingExistsForUser(administrators1[1]);
      await ensureKymaAdminBindingDoesNotExistsForUser(administrator0);
    });
  });
}

function CommerceMockTest(options) {
  describe("SKR test", function () {
    const testNS = options.testNS;
    const appName = options.appName;
    const scenarioName = options.scenarioName;
    const AWS_PLAN_ID = "361c511f-f939-4621-b228-d0fb79a1fe15";

    let cancelPortForward = null;
    before(() => {
      cancelPortForward = prometheusPortForward();
    });

    after(() => {
      cancelPortForward();
    });

    it("CommerceMock test fixture should be ready", async function () {
      await ensureCommerceMockWithCompassTestFixture(
          director,
          appName,
          scenarioName,
          "mocks",
          testNS
      );
    });

    it("function should be reachable through secured API Rule", async function () {
      await checkFunctionResponse(testNS);
    });

    it("order.created.v1 event should trigger the lastorder function", async function () {
      await sendEventAndCheckResponse();
    });

    it("Deletes the resources that have been created", async function () {
      await deleteMockTestFixture("mocks", testNS);
    });

    //Check audit log for AWS
    if (process.env.KEB_PLAN_ID === AWS_PLAN_ID) {
      const auditlogs = new AuditLogClient(AuditLogCreds.fromEnv());

      it("Check audit logs", async function () {
        await checkAuditLogs(auditlogs);
      });

      it("Amount of audit events must not exceed a certain threshold", async function () {
        await checkAuditEventsThreshold(4);
      });
    }
  });
}

module.exports = {
  CommerceMockTest,
  OIDCE2ETest
}
