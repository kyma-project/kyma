const execa = require("execa");
const fs = require('fs');
const {
    getEnvOrThrow,
    debug,
    wait
} = require("../utils");
const { inspect } = require('util')

class KCPConfig {
    static fromEnv() {
        return new KCPConfig(
            getEnvOrThrow("KCP_KEB_API_URL"),
            getEnvOrThrow("KCP_OIDC_ISSUER_URL"),
            getEnvOrThrow("KCP_GARDENER_NAMESPACE"),
            getEnvOrThrow("KCP_TECH_USER_LOGIN"),
            getEnvOrThrow("KCP_TECH_USER_PASSWORD"),
            getEnvOrThrow("KCP_OIDC_CLIENT_ID"),
            getEnvOrThrow("KCP_OIDC_CLIENT_SECRET"),
            getEnvOrThrow("KCP_MOTHERSHIP_API_URL"),
            getEnvOrThrow("KCP_KUBECONFIG_API_URL"),
        );
    }
    constructor(host, issuerURL, gardenerNamespace, username, password, clientID, clientSecret, motherShipApiUrl, kubeConfigApiUrl) {
        this.host = host;
        this.issuerURL = issuerURL;
        this.gardenerNamespace = gardenerNamespace;
        this.username = username;
        this.password = password;
        this.clientID = clientID;
        this.clientSecret = clientSecret;
        this.motherShipApiUrl = motherShipApiUrl;
        this.kubeConfigApiUrl = kubeConfigApiUrl;
    }
}

class KCPWrapper {
    constructor(config) {
        this.kcpConfigPath = config.kcpConfigPath;
        this.gardenerNamespace = config.gardenerNamespace;
        this.clientID = config.clientID;
        this.clientSecret = config.clientSecret;
        this.issuerURL = config.issuerURL;
        this.motherShipApiUrl = config.motherShipApiUrl;
        this.kubeConfigApiUrl = config.kubeConfigApiUrl;

        this.username = config.username;
        this.password = config.password;
        this.host = config.host;

        this.kcpConfigPath = `config.yaml`;
        let stream = fs.createWriteStream(`${this.kcpConfigPath}`);
        stream.once("open", (_) => {
            stream.write(`gardener-namespace: ${this.gardenerNamespace}\n`);
            stream.write(`oidc-client-id: ${this.clientID}\n`);
            stream.write(`oidc-client-secret: ${this.clientSecret}\n`);
            stream.write(`keb-api-url: ${this.host}\n`);
            stream.write(`oidc-issuer-url: ${this.issuerURL}\n`);
            stream.write(`mothership-api-url: ${this.motherShipApiUrl}\n`);
            stream.write(`kubeconfig-api-url: ${this.kubeConfigApiUrl}\n`);
            stream.write(`username: ${this.username}\n`);
            stream.end();
        });
    }

    async runtimes(query) {
        let args = [`runtimes`, `--output`, `json`];
        if (query.account) {
            args = args.concat(`--account`, `${query.account}`);
        }
        if (query.subaccount) {
            args = args.concat(`--subaccount`, `${query.subaccount}`);
        }
        if (query.instanceID) {
            args = args.concat(`--instance-id`, `${query.instanceID}`);
        }
        if (query.runtimeID) {
            args = args.concat(`--runtime-id`, `${query.runtimeID}`);
        }
        if (query.region) {
            args = args.concat(`--region`, `${query.region}`);
        }
        if (query.shoot) {
            args = args.concat(`--shoot`, `${query.shoot}`);
        }
        if (query.state) {
            args = args.concat(`--state`, `${query.state}`);
        }
        if (query.ops) {
            args = args.concat(`--ops`);
        }
        let result = await this.exec(args);
        return JSON.parse(result)
    }

    async login() {
        const args = [`login`, `-u`, `${this.username}`, `-p`, `${this.password}`];
        return await this.exec(args);
    }

    async version() {
        const args = [`--version`];
        return await this.exec(args);
    }

    async upgradeKyma (instanceID, kymaUpgradeVersion) {
        const args = [`upgrade`, `kyma`, `--version=${kymaUpgradeVersion}`, `--target`, `instance-id=${instanceID}`];
        try {
            let res = await this.exec(args);
            
            // output if successful: "OrchestrationID: 22f19856-679b-4e68-b533-f1a0a46b1eed"
            // so we need to extract the uuid
            let orchestrationID = res.split(" ")[1]
            debug(`OrchestrationID: ${orchestrationID}`)

            try {
                let orchestrationStatus = await this.ensureOrchestrationSucceeded(orchestrationID)
                return orchestrationStatus
            } catch (error) {
                debug(error)
            }

            try {
                let runtime = await this.runtimes({subaccount: subaccount})
                debug(`Runtime Status: ${inspect(runtime, false, null, false)}`)
            } catch (error) {
                debug(error)
            }

            try {
                let orchestration = await this.getOrchestrationStatus(orchestrationID)
                debug(`Orchestration Status: ${inspect(orchestration, false, null, false)}`)
            } catch (error) {
                debug(error)
            }

            try {
                let operations = await this.getOrchestrationsOperations(orchestrationID)
                debug (`Operations: ${inspect(operations, false, null, false)}`)
            } catch (error) {
                debug(error)
            }

            throw (`Kyma Upgrade failed`);

        } catch (error) {
            debug(error)
            throw new Error(`failed during upgradeKyma`);
        }
    };

    async getRuntimeStatusOperations(instanceID) {
        await this.login();
        let runtimeStatus = await this.runtimes({instanceID: instanceID, ops: true})

        return JSON.stringify(runtimeStatus, null, `\t`)
    }

    async getOrchestrationsOperations(orchestrationID) {
        // debug(`Running getOrchestrationsOperations...`)
        const args = [`orchestration`,`${orchestrationID}`,`operations`, `-o`, `json`]
        try {
            let res = await this.exec(args);
            let operations = JSON.parse(res)
            // debug(`getOrchestrationsOperations output: ${operations}`)

            return operations
        } catch (error) {
            debug(error)
            throw new Error(`failed during getOrchestrationsOperations`);
        }
    }

    async getOrchestrationsOperationStatus(orchestrationID, operationID) {
        // debug(`Running getOrchestrationsOperationStatus...`)
        const args = [`orchestration`,`${orchestrationID}`,`--operation`, `${operationID}`, `-o`, `json`]
        try {
            let res = await this.exec(args);
            res = JSON.parse(res)

            return res
        } catch (error) {
            debug(error)
            throw new Error(`failed during getOrchestrationsOperationStatus`);
        }
    }

    async getOrchestrationStatus (orchestrationID) {
        // debug(`Running getOrchestrationStatus...`)
        const args = [`orchestrations`, `${orchestrationID}`, `-o`, `json`]
        try {
            let res = await this.exec(args);
            let o = JSON.parse(res)

            debug(`OrchestrationID: ${o.orchestrationID} (${o.type} to version ${o.parameters.kyma.version}), status: ${o.state}`)
            
            let operations = await this.getOrchestrationsOperations(o.orchestrationID)
            // debug(`Got ${operations.length} operations for OrchestrationID ${o.orchestrationID}`)

            let upgradeOperation = {}
            if (operations.count > 0) {
                upgradeOperation = await this.getOrchestrationsOperationStatus(orchestrationID, operations.data[0].operationID)
                debug(`OrchestrationID: ${orchestrationID}: OperationID: ${operations.data[0].operationID}: OperationStatus: ${upgradeOperation.state}`)
            } else {
                debug (`No operations in OrchestrationID ${o.orchestrationID}`)
            }

            return o
        } catch (error) {
            debug(error)
            throw new Error(`failed during getOrchestrationStatus`);
        }
    };

    async ensureOrchestrationSucceeded(orchenstrationID) {
        // Decides whether to go to the next step of while or not based on
        // the orchestration result (0 = succeeded, 1 = failed, 2 = cancelled, 3 = pending/other)
        debug(`Waiting for Kyma Upgrade with OrchestrationID ${orchenstrationID} to succeed...`)
        try {
            const res = await wait(
            () => this.getOrchestrationStatus(orchenstrationID),
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
        } catch (error) {
            debug(error)
            throw new Error(`failed during ensureOrchestrationSucceeded`);
        }
      }

    async exec(args) {
        try {
            const defaultArgs = [
                `--config`, `${this.kcpConfigPath}`,
            ];
            // debug([`>  kcp`, defaultArgs.concat(args).join(" ")].join(" "))
            let output = await execa(`kcp`, defaultArgs.concat(args));
            // debug(output);
            return output.stdout;
        } catch (err) {
            if (err.stderr === undefined) {
                throw new Error(`failed to process kcp binary output: ${err.toString()}`);
            }
            throw new Error(`kcp command failed: ${err.stderr.toString()}`);
        }
    }
}

module.exports = {
    KCPConfig,
    KCPWrapper,
}
