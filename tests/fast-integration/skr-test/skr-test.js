const uuid = require("uuid");
const {
  KEBConfig,
  KEBClient,
  provisionSKR,
  deprovisionSKR,
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
const { debug, genRandom, initializeK8sClient } = require("../utils");

const {
  AuditLogCreds,
  AuditLogClient,
  checkAuditLogs,
  checkAuditEventsThreshold,
} = require("../audit-log");

describe("SKR test", function () {
  const keb = new KEBClient(KEBConfig.fromEnv());
  const gardener = new GardenerClient(GardenerConfig.fromEnv());
  // const director = new DirectorClient(DirectorConfig.fromEnv());

  const suffix = genRandom(4);
  const appName = `app-${suffix}`;
  const runtimeName = `kyma-${suffix}`;
  const scenarioName = `test-${suffix}`;
  const runtimeID = uuid.v4();

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
      oidc: {
        clientID: "clientID",
        groupsClaim: "groups",
        issuerURL: "https://dupa.com",
        signingAlgs: ["RS256"],
        usernameClaim: "sub",
        usernamePrefix: "-",
      },
      administrators: ["ziomeczek@admin.com"],
    };
    //-- KK set oidc config {} i admins []
    skr = await provisionSKR(
      keb,
      gardener,
      runtimeID,
      runtimeName,
      null,
      null,
      customParams
    );
    initializeK8sClient({ kubeconfig: skr.shoot.kubeconfig });

    let oidcConfig = skr.shoot.oidcConfig;
    console.log("Oidc", oidcConfig);

    //-- KK assert oidc config {} via gardener client
    // assert on skr object // KK ??
    //-- KK assert admins using k8s core API client

    //owrapowac zeby nie smiecic
    //i.e  waitForK8sObject( 'ClusterRoleBinding ', 'kyma-admin ' )
  });

  // it(`assure oidc`)

  // it(`assure cluster admins`)

  // it(`Update SKR service instance`, async function () {
  //   //-- KK dodac parametry oidc i admins []
  //   skr = await updateSKR(keb, gardener, runtimeID, runtimeName);

  //   //-- KK assert oidc gardener client
  //   gardener.getShoot(); // KK ??
  //   //-- KK assert admins using k8s core API client
  // });

  // it("Assign SKR to scenario", async function () {
  //   await addScenarioInCompass(director, scenarioName);
  //   await assignRuntimeToScenario(director, skr.shoot.compassID, scenarioName);
  // });

  // it("CommerceMock test fixture should be ready", async function () {
  //   await ensureCommerceMockWithCompassTestFixture(
  //     director,
  //     appName,
  //     scenarioName,
  //     "mocks",
  //     testNS
  //   );
  // });

  // it("function should be reachable through secured API Rule", async function () {
  //   await checkFunctionResponse(testNS);
  // });

  // it("order.created.v1 event should trigger the lastorder function", async function () {
  //   await sendEventAndCheckResponse();
  // });

  // it("Deletes the resources that have been created", async function () {
  //   await deleteMockTestFixture("mocks", testNS);
  // });

  // // Check audit log for AWS
  // if (process.env.KEB_PLAN_ID == AWS_PLAN_ID) {
  //   const auditlogs = new AuditLogClient(AuditLogCreds.fromEnv());

  //   it("Check audit logs", async function () {
  //     await checkAuditLogs(auditlogs);
  //   });

  //   it("Amount of audit events must not exceed a certain threshold", async function () {
  //     await checkAuditEventsThreshold(2.5);
  //   });
  // }

  it("Deprovision SKR", async function () {
    await deprovisionSKR(keb, runtimeID);
  });

  it("Unregister SKR resources from Compass", async function () {
    await unregisterKymaFromCompass(director, scenarioName);
  });
});
