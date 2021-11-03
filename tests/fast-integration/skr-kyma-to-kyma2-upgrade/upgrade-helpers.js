const {
    debug,
    wait
} = require("../utils");
const execa = require("execa");
const { expect } = require("chai");

async function kcpLogin (kcpconfigPath, kcpUser, kcpPassword) {
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
        args = [`upgrade`, `kyma`, `--config`, `${kcpconfigPath}`, `--version`, `"${kymaUpgradeVersion}"`, `--target`, `subaccount=${subaccount}`]
        debug(`Executing: "kcp ${args.join(' ')}"`)
        let output = await execa(`kcp`, args);
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

async function getOrchestrationStatus (kcpconfigPath, orchestrationID) {
    debug(`Running getOrchestrationStatus...`)
    let args = []
    try {
        args = [`orchestrations`,  `--config`, `${kcpconfigPath}`, `${orchestrationID}`, `-o`, `json`]
        let orchestrations = await execa(`kcp`, args);
        let o = JSON.parse(orchestrations.stdout)
        debug(`OrchestrationStatus: orchestrationID: ${o.orchestrationID} (${o.type}), status: ${o.state}`)

        return o
    } catch (error) {
        console.log(error)
        if (error.stderr === undefined) {
            throw new Error(`failed to process output of "kcp ${args.join(' ')}": ${error}`);
        }
        throw new Error(`failed "kcp ${args.join(' ')}": ${error.stderr}`);
    }
};

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
    kcpUpgrade
}