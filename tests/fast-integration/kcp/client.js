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
        // this.configFile = config.file();
        this.gardenerNamespace = config.gardenerNamespace;
        this.clientID = config.clientID;
        this.clientSecret = config.clientSecret;
        this.issuerURL = config.issuerURL;

        this.username = config.username;
        this.password = config.password;
        this.host = config.host;
    }

    async runtimes(query) {
        let args = [`runtimes`, `--output`, `json`];
        if (query.account) {
            args.concat(`--account`, `${query.account}`);
        }
        if (query.subaccount) {
            args.concat(`--subaccount`, `${query.subaccount}`);
        }
        if (query.instanceID) {
            args.concat(`--instance-id`, `${query.instanceID}`);
        }
        if (query.runtimeID) {
            args.concat(`--runtime-id`, `${query.runtimeID}`);
        }
        if (query.region) {
            args.concat(`--region`, `${query.region}`);
        }
        if (query.shoot) {
            args.concat(`--shoot`, `${query.shoot}`);
        }
        if (query.state) {
            args.concat(`--state`, `${query.state}`);
        }
        console.log(args)
        return await this.exec(args);
    }

    async login() {
        const args = [`login`, `-u`, `${this.username}`, `-p`, `${this.password}`];
        return await this.exec(args);
    }

    async exec(args) {
        try {
            const defaultArgs = [
                `--gardener-namespace`, `${this.gardenerNamespace}`,
                `--keb-api-url`, `${this.host}`,
                `--oidc-issuer-url`, `${this.issuerURL}`,
                `--oidc-client-id`, `${this.clientID}`,
                `--oidc-client-secret`, `${this.clientSecret}`,
            ];
            let output = await execa(`kcp`, defaultArgs.concat(args));
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
