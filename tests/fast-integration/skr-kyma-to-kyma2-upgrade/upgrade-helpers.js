const {
    debug,
    wait
} = require("../utils");
const execa = require("execa");
const { expect } = require("chai");

async function kcpLogin (kcpconfigPath, kcpUser, kcpPassword) {
    debug(`Running kcpLogin...`)
    let args = []
    try {
        args = [`login`, `--config`, `${kcpconfigPath}`, `-u`, `${kcpUser}`, `-p`, `${kcpPassword}`]
        let output = await execa(`kcp`, args);
        debug(`"kcp login --config ${kcpconfigPath} -u $KCP_TECH_USER_LOGIN -p $KCP_TECH_USER_PASSWORD" exited with code ${output.exitCode}`)
        return output
    } catch (error) {
        console.log(error)
        debug.log(error)
        if (error.stderr === undefined) {
            throw new Error(`failed to process output of "kcp ${args.join(' ')}": ${error}`);
        }
        throw new Error(`failed "kcp ${args.join(' ')}": ${error.stderr}`);
    }
};

async function kcpUpgrade (kcpconfigPath, subaccount, runtimeID, kymaUpgradeVersion) {
    debug(`Running kcpUpgrade...`)
    let args = []
    try {
        args = [`upgrade`, `kyma`, `--config`, `${kcpconfigPath}`, `--version`, `"${kymaUpgradeVersion}"`, `--target`, `subaccount=${subaccount},runtime-id=${runtimeID}`]
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

        debug(`getOrchestrationStatus output: ${orchestrations.stdout}`)

        return JSON.parse(orchestrations.stdout)
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
      1000 * 5 // 5 seconds
    );
  
    debug("KEB Orchestration Status:", res);

    if(res.state !== "succeeded") {
        throw(`orchestration didn't succeed in 15min: ${JSON.stringify(res)}`);
    }
  
    return res;
  }

// getOrchestrationStatus ("dev.yaml", "210779e4-bd9f-4fb7-aa10-888520038da5")

module.exports = {
    kcpLogin,
    kcpUpgrade
}