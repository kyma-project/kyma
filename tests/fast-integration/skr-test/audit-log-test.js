const { 
    Creds,
    AuditLogClient,
    waitForK8sResources,
    createNamespace,
  } = require("../audit-log");

const {
  kubectlApplyDir,
  kubectlDeleteDir,
  sleep,
} = require("../utils");

const cred = new AuditLogClient(Creds.fromEnv())

describe("Check for audit logs", function(){
      
  this.timeout(10 * 60 * 1000);
  this.slow(5000);

  it (`creates namespace audit-test`, async function(){
    await createNamespace("audit-test")
  })

  it (`creates k8s resources`, async function(){
    await kubectlApplyDir("./audit-log/fixtures", "audit-test")
  })

  it (`waits for k8s resources to be created`, async function() {
    await waitForK8sResources()
  })

  it (`deletes function`, async function(){
    await kubectlDeleteDir("./audit-log/fixtures", "audit-test")
  })  
  it (`sleeps for 2 minutes so that logs are sent to audit log server`, async function(){
    await sleep(120000);

  })
  it (`fetch logs`, async function() {
    await cred.fetchLogs()
  })
  it(`checks networking istio group is logged for action create`, async function() {
    await cred.parseLogs("networking.istio.io", "create")
  })
  it(`checks networking istio group is logged for action delete`, async function() {
    await cred.parseLogs("networking.istio.io", "delete")
  })
  it(`checks authorization group is logged for action create`, async function() {
    await cred.parseLogs("rbac.authorization.k8s.io", "create")
  })
  it(`checks authorization group is logged for action delete`, async function() {
    await cred.parseLogs("rbac.authorization.k8s.io", "delete")
  })
  it(`checks serverless group is logged for action create`, async function() {
    await cred.parseLogs("serverless.kyma-project.io", "create")
  })

  it(`checks serverless group is logged for action delete`, async function() {
    await cred.parseLogs("serverless.kyma-project.io", "delete")
  })

  it(`checks core group is logged for action create`, async function() {
    await cred.parseLogs("foo-config", "create")
  })
  it(`checks core group is logged for action delete`, async function() {
    await cred.parseLogs("foo-config", "delete")
  })
})