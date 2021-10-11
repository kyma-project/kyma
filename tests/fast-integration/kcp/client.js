const execa = require("execa");
const fs = require('fs');
const {getEnvOrThrow, debug} = require("../utils");

class KCPConfig {
    kcpConfigPath = `config.yaml`;
    static fromEnv() {
        return new KCPConfig(
            getEnvOrThrow("KCP_KEB_API_URL"),
            getEnvOrThrow("KCP_OIDC_ISSUER_URL"),
            getEnvOrThrow("KCP_GARDENER_NAMESPACE"),
            getEnvOrThrow("KCP_TECH_USER_LOGIN"),
            getEnvOrThrow("KCP_TECH_USER_PASSWORD"),
            getEnvOrThrow("KCP_OIDC_CLIENT_ID"),
            getEnvOrThrow("KCP_OIDC_CLIENT_SECRET"),
        );
    }
    constructor(host, issuerURL, gardenerNamespace, username, password, clientID, clientSecret) {
        this.host = host;
        this.issuerURL = issuerURL;
        this.gardenerNamespace = gardenerNamespace;
        this.login = username;
        this.password = password;
        this.clientID = clientID;
        this.clientSecret = clientSecret;
    }

    file() {
        let stream = fs.createWriteStream(`${this.kcpConfigPath}`);
        stream.once("open", (_) => {
            stream.write(`gardener-namespace: ${this.gardenerNamespace}\n`);
            stream.write(`oidc-client-id: ${this.clientID}\n`);
            stream.write(`oidc-client-secret: ${this.clientSecret}\n`);
            stream.write(`keb-api-url: ${this.host}\n`);
            stream.write(`oidc-issuer-url: ${this.issuerURL}\n`);
            stream.end();
        })
        return this.kcpConfigPath;
    }
}

class KCPWrapper {
    constructor(config) {
        this.configFile = config.file();
        this.username = config.username;
        this.password = config.password;
        this.host = config.host;
    }

    async runtimes(...customOptions) {
        let args = [`runtimes`, `--output`, `json`];
        customOptions.forEach((option) => {
            if (option.account) {
                args += [`--account`, `${option.account}`];
            }
            if (option.subaccount) {
                args += [`--subaccount`, `${option.subaccount}`];
            }
            if (option.instanceID) {
                args += [`--instance-id`, `${option.instanceID}`];
            }
            if (option.runtimeID) {
                args += [`--runtime-id`, `${option.runtimeID}`];
            }
            if (option.region) {
                args += [`--region`, `${option.region}`];
            }
            if (option.shoot) {
                args += [`--shoot`, `${option.shoot}`];
            }
            if (option.state) {
                args += [`--state`, `${option.state}`];
            }
        });
        return await this.execCmd(args);
    }

    async login() {
        const args = [`login`, `--username`, `${this.username}`, `--password`, `${this.password}`];
        return await this.execCmd(args);
    }

    async execCmd(args) {
        debug(args);
        try {
            let output = await execa(`kcp`, args + [`--config`, `${this.configFile}`]);
            debug(output);
            return output;
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
