const uuid = require("uuid");
const {
  KEBConfig,
  KEBClient,
  provisionSKR,
  deprovisionSKR,
  updateSKR,
  ensureValidShootOIDCConfig,
  ensureValidOIDCConfigInCustomerFacingKubeconfig,
} = require("../kyma-environment-broker");
const {
  DirectorConfig,
  DirectorClient,
  addScenarioInCompass,
  assignRuntimeToScenario,
  unregisterKymaFromCompass,
} = require("../compass");
const { GardenerConfig, GardenerClient } = require("../gardener");
const {
  ensureCommerceMockWithCompassTestFixture,
  checkFunctionResponse,
  sendEventAndCheckResponse,
  deleteMockTestFixture,
} = require("../test/fixtures/commerce-mock");
const {
  debug,
  genRandom,
  initializeK8sClient,
  ensureKymaAdminBindingExistsForUser,
  ensureKymaAdminBindingDoesNotExistsForUser,
  getEnvOrThrow,
} = require("../utils");

const {
  AuditLogCreds,
  AuditLogClient,
  checkAuditLogs,
  checkAuditEventsThreshold,
} = require("../audit-log");

describe("SKR test", function () {
  const keb = new KEBClient(KEBConfig.fromEnv());
  const gardener = new GardenerClient(GardenerConfig.fromEnv());
  const director = new DirectorClient(DirectorConfig.fromEnv());

  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const scenarioName = `test-${suffix}`;
  const runtimeID = uuid.v4();
  const oidc0 = {
    clientID: "abc-xyz",
    groupsClaim: "groups",
    issuerURL: "https://custom.ias.com",
    signingAlgs: ["RS256"],
    usernameClaim: "sub",
    usernamePrefix: "-",
  };

  const administrator0 = getEnvOrThrow("KEB_USER_ID");

  const oidc1 = {
    clientID: "foo-bar",
    groupsClaim: "groups1",
    issuerURL: "https://new.custom.ias.com",
    signingAlgs: ["RS256"],
    usernameClaim: "email",
    usernamePrefix: "acme-",
  };
  const administrators1 = ["admin1@acme.com", "admin2@acme.com"];

  debug(
    `RuntimeID ${runtimeID}`,
    `Scenario ${scenarioName}`,
    `Runtime ${runtimeName}`,
    `Application ${appName}`
  );

  const testNS = "skr-test";
  const AWS_PLAN_ID = "361c511f-f939-4621-b228-d0fb79a1fe15";

  this.timeout(60 * 60 * 1000 * 3); // 3h
  this.slow(5000);

  let skr;

  it(`Provision SKR with ID ${runtimeID}`, async function () {
    const customParams = {
      oidc: oidc0,
    };

    skr = await provisionSKR(keb, gardener, runtimeID, runtimeName, null, null, customParams);
    initializeK8sClient({ kubeconfig: skr.shoot.kubeconfig });
  });

  it(`Assure initial OIDC config is applied on shoot cluster`, async function () {
    ensureValidShootOIDCConfig(skr.shoot, oidc0);
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

    skr = await updateSKR(keb, gardener, runtimeID, skr.shoot.name, customParams);
  });

  it(`Assure updated OIDC config is applied on shoot cluster`, async function () {
    ensureValidShootOIDCConfig(skr.shoot, oidc1);
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

    skr = await updateSKR(keb, gardener, runtimeID, skr.shoot.name, customParams);
  });

  it(`Assure only new cluster admins are configured`, async function () {
    await ensureKymaAdminBindingExistsForUser(administrators1[0]);
    await ensureKymaAdminBindingExistsForUser(administrators1[1]);
    await ensureKymaAdminBindingDoesNotExistsForUser(administrator0);
  });

  it("Assign SKR to scenario", async function () {
    await addScenarioInCompass(director, scenarioName);
    await assignRuntimeToScenario(director, skr.shoot.compassID, scenarioName);
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
  if (process.env.KEB_PLAN_ID == AWS_PLAN_ID) {
    const auditlogs = new AuditLogClient(AuditLogCreds.fromEnv());

    it("Check audit logs", async function () {
      await checkAuditLogs(auditlogs);
    });

    it("Amount of audit events must not exceed a certain threshold", async function () {
      await checkAuditEventsThreshold(2.5);
    });
  }

  it("Deprovision SKR", async function () {
    await deprovisionSKR(keb, runtimeID);
  });

  it("Unregister SKR resources from Compass", async function () {
    await unregisterKymaFromCompass(director, scenarioName);
  });
});
