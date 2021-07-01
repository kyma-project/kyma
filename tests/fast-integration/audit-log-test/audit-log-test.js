const { 
    AuditLogCreds,
    AuditLogClient,
    // waitForK8sResources,
    createNamespace,
    deployK8sResources,
    deleteK8sResources,
    waitForAuditLogs,
    checkAuditLogs
  } = require("../audit-log");

const {
  kubectlApplyDir,
  kubectlDeleteDir,
  sleep,
} = require("../utils");

const cred = new AuditLogClient(AuditLogCreds.fromEnv())

// Test Logs
describe("Check for audit logs", function(){
      
  this.timeout(10 * 60 * 1000);
  this.slow(5000);

  // if (process.env.KEB_REGION == "cf-eu10" && process.env.KEB_PLAN_ID =="b1a5764e-2ea1-4f95-94c0-2b4538b37b55") {
    it (`creates namespace audit-test`, async function(){
      const groups = [
        { "resName": "commerce-binding", "groupName": "servicecatalog.k8s.io", "action": "create" },
        { "resName": "commerce-binding", "groupName": "servicecatalog.k8s.io", "action": "delete" },
        { "resName": "lastorder", "groupName": "serverless.kyma-project.io", "action": "create" },
        { "resName": "lastorder", "groupName": "serverless.kyma-project.io", "action": "delete" },
        {"resName":"commerce-mock", "groupName": "deployments", "action": "create"},
        {"resName":"commerce-mock", "groupName": "deployments", "action": "delete"}
      ]
    await checkAuditLogs(cred, groups)
  
    })
  // }

  

  // // Deprovision SKR
  // it("Deprovision SKR", async function() {
  //   await deprovisionSKR(keb, runtimeID);
  // });

  // it("Unregister SKR resources from Compass", async function() {
  //   await unregisterKymaFromCompass(director, scenarioName);
  // });
  // it (`creates k8s resources`, async function(){
  //   await kubectlApplyDir("./audit-log/fixtures", "audit-test")
  // })

  // it (`waits for k8s resources to be created`, async function() {
  //   await waitForK8sResources()
  // })

  // it (`deletes function`, async function(){
  //   await kubectlDeleteDir("./audit-log/fixtures", "audit-test")
  // })  
  // it (`sleeps for 2 minutes so that logs are sent to audit log server`, async function(){
  //   await sleep(120000);

  // })
  // it (`fetch logs`, async function() {
  //   await cred.fetchLogs()
  // })
  // it(`checks networking istio group is logged for action create`, async function() {
  //   await cred.parseLogs("monitoring.coreos.com", "create")
  // })
  // it(`checks networking istio group is logged for action delete`, async function() {
  //   await cred.parseLogs("monitoring.coreos.com", "delete")
  // })
  // it(`checks authorization group is logged for action create`, async function() {
  //   await cred.parseLogs("rbac.authorization.k8s.io", "create")
  // })
  // it(`checks authorization group is logged for action delete`, async function() {
  //   await cred.parseLogs("rbac.authorization.k8s.io", "delete")
  // })
  // it(`checks serverless group is logged for action create`, async function() {
  //   await cred.parseLogs("serverless.kyma-project.io", "create")
  // })

  // it(`checks serverless group is logged for action delete`, async function() {
  //   await cred.parseLogs("serverless.kyma-project.io", "delete")
  // })

  // it(`checks core group is logged for action create`, async function() {
  //   await cred.parseLogs("foo-config", "create")
  // })
  // it(`checks core group is logged for action delete`, async function() {
  //   await cred.parseLogs("foo-config", "delete")
  // })
})