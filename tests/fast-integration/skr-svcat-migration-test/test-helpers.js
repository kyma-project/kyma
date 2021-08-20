const installer = require("../installer/helm");
const execa = require("execa");
const fs = require('fs');
const os = require('os');
const {
    getEnvOrThrow, getConfigMap
} = require("../utils");

class SMCreds {
    static fromEnv() {
        return new SMCreds(
            // TODO: rename to BTP_SM_ADMIN_CLIENTID
            getEnvOrThrow("BTP_OPERATOR_CLIENTID"),
            // TODO: rename to BTP_SM_ADMIN_CLIENTID
            getEnvOrThrow("BTP_OPERATOR_CLIENTSECRET"),
            // TODO: rename to BTP_SM_URL
            getEnvOrThrow("BTP_OPERATOR_URL")
        );
    }

    constructor(clientid, clientsecret, url) {
        this.clientid = clientid;
        this.clientsecret = clientsecret;
        this.url = url;
    }
}

async function saveKubeconfig(kubeconfig) {
    fs.mkdirSync(`${os.homedir()}/.kube`, true);
    fs.writeFileSync(`${os.homedir()}/.kube/config`, kubeconfig);
}

async function readClusterID() {
    let cm = await getConfigMap("cluster-info", "kyma-system")
    return cm.data.id
}

async function installBTPOperatorHelmChart(creds, clusterId) {
    const btpChart = "https://github.com/kyma-incubator/sap-btp-service-operator/releases/download/v0.1.10/sap-btp-operator-v0.1.10.tgz";
    const btp = "sap-btp-operator";
    const btpValues = `manager.secret.clientid=${creds.clientId},manager.secret.clientsecret=${creds.clientSecret},manager.secret.url=${creds.smURL},manager.secret.tokenurl=${creds.url},cluster.id=${clusterId}`
    try {
        await installer.helmInstallUpgrade(btp, btpChart, btp, btpValues, null, ["--create-namespace"]);
    } catch (error) {
        if (error.stderr === undefined) {
            throw new Error(`failed to install ${btp}: ${error}`);
        }
        throw new Error(`failed to install ${btp}: ${error.stderr}`);
    }
}

async function installBTPServiceOperatorMigrationHelmChart() {
    const chart = "https://github.com/kyma-incubator/sc-removal/releases/download/0.3.0/sap-btp-service-operator-migration-0.3.0.tar.gz";
    const btp = "sap-btp-service-operator-migration";
    const image = {
        repository: "eu.gcr.io/sap-se-cx-gopher/sap-btp-service-operator-migration",
        tag: "v0.3.0"
    }
    const values = `image.repository=${image.repository},image.tag=${image.tag}`

    try {
        await installer.helmInstallUpgrade(btp, chart, "sap-btp-operator", values, null, ["--create-namespace"]);
    } catch (error) {
        if (error.stderr === undefined) {
            throw new Error(`failed to install ${btp}: ${error}`);
        }
        throw new Error(`failed to install ${btp}: ${error.stderr}`);
    }
}

async function provisionPlatform(creds, svcatPlatform) {
    let args = [];
    try {
        args = [`login`, `-a`, creds.url, `--param`, `subdomain=e2etestingscmigration`, `--auth-flow`, `client-credentials`]
        await execa(`smctl`, args.concat([`--client-id`, creds.clientid, `--client-secret`, creds.clientsecret]));

        // $ smctl register-platform <name> kubernetes -o json
        // Output:
        // {
        //   "id": "<platform-id/cluster-id>",
        //   "name": "<name>",
        //   "type": "kubernetes",
        //   "created_at": "...",
        //   "updated_at": "...",
        //   "credentials": {
        //     "basic": {
        //       "username": "...",
        //       "password": "..."
        //     }
        //   },
        //   "labels": {
        //     "subaccount_id": [
        //       "..."
        //     ]
        //   },
        //   "ready": true
        // }
        args = [`register-platform`, svcatPlatform, `kubernetes`, `-o`, `json`]
        let registerPlatformOut = await execa(`smctl`, args);
        let platform = JSON.parse(registerPlatformOut.stdout)

        return {
            clusterId: platform.id,
            name: platform.name,
            credentials: platform.credentials.basic,
        }

    } catch (error) {
        if (error.stderr === undefined) {
            throw new Error(`failed to process output of "smctl ${args.join(' ')}": ${error}`);
        }
        throw new Error(`failed "smctl ${args.join(' ')}": ${error.stderr}`);
    }
}

async function smInstanceBinding(btpOperatorInstance, btpOperatorBinding) {
    let args = [];
    try {

        args = [`provision`, btpOperatorInstance, `service-manager`, `service-operator-access`, `--mode=sync`]
        await execa(`smctl`, args);

        // Move to Operator Install
        args = [`bind`, btpOperatorInstance, btpOperatorBinding, `--mode=sync`];
        await execa(`smctl`, args);
        args = [`get-binding`, btpOperatorBinding, `-o`, `json`];
        let out = await execa(`smctl`, args);
        let b = JSON.parse(out.stdout)
        let c = b.items[0].credentials

        return {
            clientId: c.clientid,
            clientSecret: c.clientsecret,
            smURL: c.sm_url,
            url: c.url,
            instanceId: b.items[0].service_instance_id,
        }

    } catch (error) {
        if (error.stderr === undefined) {
            throw new Error(`failed to process output of "smctl ${args.join(' ')}": ${error}`);
        }
        throw new Error(`failed "smctl ${args.join(' ')}": ${error.stderr}`);
    }
}

async function markForMigration(creds, svcatPlatform, btpOperatorInstanceId) {
    let errors = [];
    let args = [];
    try {
        args = [`login`, `-a`, creds.url, `--param`, `subdomain=e2etestingscmigration`, `--auth-flow`, `client-credentials`]
        await execa(`smctl`, args.concat([`--client-id`, creds.clientid, `--client-secret`, creds.clientsecret]));
    } catch (error) {
        errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
    }

    try {
        // usage: smctl curl -X PUT -d '{"sourcePlatformID": ":platformID"}' /v1/migrate/service_operator/:instanceID
        let data = {sourcePlatformID: svcatPlatform}
        args = ['curl', '-X', 'PUT', '-d', JSON.stringify(data), '/v1/migrate/service_operator/' + btpOperatorInstanceId]
        await execa('smctl', args)
    } catch (error) {
        errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
    }
    if (errors.length > 0) {
        throw new Error(errors.join(", "));
    }
}

async function cleanupInstanceBinding(creds, svcatPlatform, btpOperatorInstance, btpOperatorBinding) {
    let errors = [];
    let args = [];
    try {
        args = [`login`, `-a`, creds.url, `--param`, `subdomain=e2etestingscmigration`, `--auth-flow`, `client-credentials`]
        await execa(`smctl`, args.concat([`--client-id`, creds.clientid, `--client-secret`, creds.clientsecret]));
    } catch (error) {
        errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
    }

    try {
        args = [`unbind`, btpOperatorInstance, btpOperatorBinding, `-f`, `--mode=sync`];
        await execa(`smctl`, args);
        // let {stdout} = await execa(`smctl`, args);
        // if (stdout !== "Service Binding successfully deleted.") {
        //     errors = errors.concat([`failed "smctl ${args.join(' ')}": ${stdout}`])
        // }
    } catch (error) {
        errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
    }

    try {
        // hint: probably should fail cause that instance created other instannces (after the migration is done)
        args = [`deprovision`, btpOperatorInstance, `-f`, `--mode=sync`];
        let {stdout} = await execa(`smctl`, args);
        if (stdout !== "Service Instance successfully deleted.") {
            errors = errors.concat([`failed "smctl ${args.join(' ')}": ${stdout}`])
        }
    } catch (error) {
        errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
    }

    try {
        args = [`delete-platform`, svcatPlatform, `-f`, "--cascade"];
        let {stdout} = await execa(`smctl`, args);
        if (stdout !== "Platform(s) successfully deleted.") {
            errors = errors.concat([`failed "smctl ${args.join(' ')}": ${stdout}`])
        }
    } catch (error) {
        errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
    }

    if (errors.length > 0) {
        throw new Error(errors.join(", "));
    }
}

module.exports = {
    provisionPlatform,
    smInstanceBinding,
    cleanupInstanceBinding,
    installBTPOperatorHelmChart,
    installBTPServiceOperatorMigrationHelmChart,
    saveKubeconfig,
    markForMigration,
    readClusterID,
    SMCreds,
};
