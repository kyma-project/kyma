const {
    debug,
    wait,
    switchDebug
} = require("../utils");

const execa = require("execa");
const { expect } = require("chai");
switchDebug(on = true)

async function kcpVersion () {
    let args = []
    try {
        args = [`--version`]
        debug(`Executing: "kcp ${args.join(' ')}"`)
        let output = await execa(`kcp`, args);
        let version = output.stdout.split(" ")[2]

        return version
    } catch (error) {
        console.log(error)
        if (error.stderr === undefined) {
            throw new Error(`failed to process output of "kcp ${args.join(' ')}": ${error}`);
        }
        throw new Error(`failed "kcp ${args.join(' ')}": ${error.stderr}`);
    }
};

async function kcpLogin (kcpconfigPath, kcpUser, kcpPassword) {
    let version = await kcpVersion()
    debug(`Using KCP-CLI Version: ${version}`)

    debug(`Running kcpLogin...`)
    let args = []
    // Note: dummycmd is just for output
    let dummycmd=`kcp login --config ${kcpconfigPath} -u $KCP_TECH_USER_LOGIN -p $KCP_TECH_USER_PASSWORD`
    try {
        args = [`login`, `--config`, `${kcpconfigPath}`, `-u`, `${kcpUser}`, `-p`, `${kcpPassword}`]
        let output = await execa(`kcp`, args);
        debug(`"${dummycmd}" exited with code ${output.exitCode}`)
        return output
    } catch (error) {
        if (error.stderr === undefined) {
            throw new Error(`failed to process output of "${dummycmd}": ${error}`);
        }
        throw new Error(`failed "${dummycmd}": ${error.stderr}`);
    }
};

async function kcpUpgrade (kcpconfigPath, subaccount, kymaUpgradeVersion) {
    debug(`Running kymaUpgrade...`)
    let args = []
    try {
        args = [`--verbose=6`, `upgrade`, `kyma`, `--config`, `${kcpconfigPath}`, `--version`, `"${kymaUpgradeVersion}"`, `--target`, `subaccount=${subaccount}`]
        debug(`Executing: "kcp ${args.join(' ')}"`)
        let output = await execa(`kcp`, args);
        debug(output)
        // output if successful: "OrchestrationID: 22f19856-679b-4e68-b533-f1a0a46b1eed"
        // so we need to extract the uuid
        let orchestrationID = output.stdout.split(" ")[1]
        console.log(`OrchestrationID: ${orchestrationID}`)
        let orchestrationStatus = await ensureOrchestrationSucceeded(kcpconfigPath, orchestrationID)

        return orchestrationStatus
    } catch (error) {
        console.log(error)
        if (error.stderr === undefined) {
            throw new Error(`failed to process output of "kcp ${args.join(' ')}": ${error}`);
        }
        throw new Error(`failed "kcp ${args.join(' ')}": ${error.stderr}`);
    }
};

kcpUpgrade("/Users/cvoigt/tmp/config-dev-auto.yaml", "cee91d4e-b54e-4ee0-8258-ee3e15d57ad2", "2.0.0-rc4")

async function getOrchestrationStatus (kcpconfigPath, orchestrationID) {
    let args = []
    try {
        args = [`orchestrations`,  `--config`, `${kcpconfigPath}`, `${orchestrationID}`, `-o`, `json`]
        let orchestrations = await execa(`kcp`, args);
        let o = JSON.parse(orchestrations.stdout)
        debug(`OrchestrationStatus: orchestrationID: ${o.orchestrationID} (${o.type}), status: ${o.state}`)

        let operations = await getOrchestrationsOperations(kcpconfigPath, o.orchestrationID)
        debug(`Got ${operations.length} operations for OrchestrationID ${o.orchestrationID}`)
        let upgradeOperation = {}
        if (operations.length > 0) {
            upgradeOperation = await getOrchestrationsOperationStatus(kcpconfigPath, orchestrationID, operations[0].operationID)
            debug(`OrchestrationID: ${orchestrationID}: OperationID: ${operations[0].operationID}: OperationStatus: ${upgradeOperation.state}`)
            debug(upgradeOperation)
        } else {
            debug (`No operations in OrchestrationID ${o.orchestrationID}`)
        }


        return o
    } catch (error) {
        console.log(error)
        if (error.stderr === undefined) {
            throw new Error(`failed to process output of "kcp ${args.join(' ')}": ${error}`);
        }
        throw new Error(`failed "kcp ${args.join(' ')}": ${error.stderr}`);
    }
};

async function getOrchestrationsOperations(kcpconfigPath, orchestrationID) {
    let args = []
    try {
        args = [`--config`, `${kcpconfigPath}`, `orchestration`,`${orchestrationID}`,`operations`, `-o`, `json`]
        let res = await execa(`kcp`, args);
        let operations = JSON.parse(res.stdout)

        debug(`getOrchestrationsOperations output: ${operations}`)

        if (operations.data === undefined) {
            return []
        }

        return operations.data
    } catch (error) {
        console.log(error)
        if (error.stderr === undefined) {
            throw new Error(`failed to process output of "kcp ${args.join(' ')}": ${error}`);
        }
        throw new Error(`failed "kcp ${args.join(' ')}": ${error.stderr}`);
    }
}

async function getOrchestrationsOperationStatus(kcpconfigPath, orchestrationID, operationID) {
    debug(`Running getOrchestrationsOperationStatus...`)
    let args = []
    try {
        args = [`--config`, `${kcpconfigPath}`, `orchestration`,`${orchestrationID}`,`--operation`, `${operationID}`, `-o`, `json`]
        let operation = await execa(`kcp`, args);

        debug(`getOrchestrationsOperationStatus output: ${operation.stdout}`)

        return JSON.parse(operation.stdout)
    } catch (error) {
        console.log(error)
        if (error.stderr === undefined) {
            throw new Error(`failed to process output of "kcp ${args.join(' ')}": ${error}`);
        }
        throw new Error(`failed "kcp ${args.join(' ')}": ${error.stderr}`);
    }
}

async function ensureOrchestrationSucceeded(kcpconfigPath, orchenstrationID) {
    // Decides whether to go to the next step of while or not based on
    // the orchestration result (0 = succeeded, 1 = failed, 2 = cancelled, 3 = pending/other)
    debug(`Running ensureOrchestrationSucceeded...`)
    const res = await wait(
      () => getOrchestrationStatus(kcpconfigPath, orchenstrationID),
      (res) => res && res.state && (res.state === "succeeded" || res.state === "failed"),
      1000*60*15, // 15 min
      1000 * 30 // 30 seconds
    );

    if(res.state !== "succeeded") {
        debug("KEB Orchestration Status:", res);
        throw(`orchestration didn't succeed in 15min: ${JSON.stringify(res)}`);
    }

    const descSplit = res.description.split(" ");
    if (descSplit[1] !== "1") {
        throw(`orchestration didn't succeed (number of scheduled operations should be equal to 1): ${JSON.stringify(res)}`);
    }
  
    return res;
  }

module.exports = {
    kcpLogin,
    kcpUpgrade,
    kcpVersion
}