const { 
    waitForFunction,
    k8sApply,
    kubectlApplyDir,
    kubectlDeleteDir,
    sleep
} = require("../utils");

async function waitForK8sResources(){
    await waitForFunction("audit-test-fn", "audit-test", timeout = 120000);
}

async function createNamespace(namespace) {
    await k8sApply([namespaceObj(namespace)])
}

async function deployK8sResources() {
    // try {
    //     const resp = await axios.get(url, {
    //             headers: {
    //                 "Authorization": `Bearer ${token}`
    //             }
    //         })
    //     this._logs = resp.data
    // }
    // catch(err) {
    //     const msg = "Error when fetching logs from audit log service"
    //     if (err.response) {
    //         throw new Error(
    //         `${msg}: ${err.response.status} ${err.response.statusText}`
    //         );
    //     } else {
    //         throw new Error(`${msg}: ${err.toString()}`);
    //     }
    // }
    await kubectlApplyDir("./audit-log/fixtures", "audit-test")
    await waitForK8sResources()
}

async function deleteK8sResources() {
    await kubectlDeleteDir("./audit-log/fixtures", "audit-test")
}

async function waitForAuditLogs() {
    await sleep(120000);
}

async function checkAuditLogs(cred) {
    let logs = await cred.fetchLogs();
    console.log(logs)
    let groups = [
        "monitoring.coreos.com",
        "rbac.authorization.k8s.io",
        "serverless.kyma-project.io",
        "foo-config"
    ]
    let actions = [
        "create",
        "delete"
    ]
    logs.forEach (element => {
        groups.forEach(group => {
            actions.forEach(action => {
                
            })
        });
    })
    
}




function namespaceObj(name) {
    return {
      apiVersion: "v1",
      kind: "Namespace",
      metadata: { name },
    };
  }

module.exports = {
    waitForK8sResources,
    createNamespace,
    deployK8sResources,
    deleteK8sResources,
    waitForAuditLogs,
    checkAuditLogs
}