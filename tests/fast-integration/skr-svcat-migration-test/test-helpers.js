const installer = require("../installer/helm");
const k8s = require("@kubernetes/client-node");
const execa = require("execa");

async function installHelmCharts(creds) {
  const btpChart = "https://github.com/kyma-incubator/sap-btp-service-operator/releases/download/v0.1.9-custom/sap-btp-operator-custom-0.1.9.tar.gz";
  const btp = "sap-btp-operator";
  const btpValues = `manager.secret.clientid=${creds.clientId},manager.secret.clientsecret=${creds.clientSecret},manager.secret.url=${creds.smURL},manager.secret.tokenurl=${creds.url},cluster.id=${creds.clusterId}`
  try {
    await installer.helmInstallUpgrade(btp, btpChart, btp, btpValues, null, ["--create-namespace"]);
  } catch(error) {
    if (error.stderr === undefined) {
      throw new Error(`failed to install btp-operator: ${error}`);
    }
    throw new Error(`failed to install btp-operator: ${error.stderr}`);
  }
}

async function smInstanceBinding(url, clientid, clientsecret, svcatPlatform, btpOperatorInstance, btpOperatorBinding) {
  try {
    let args = [`login`, `-a`, url, `--param`, `subdomain=e2etestingscmigration`, `--auth-flow`, `client-credentials`]
    await execa(`smctl`, args.concat([`--client-id`, clientid, `--client-secret`, clientsecret]));

    args = [`register-platform`, svcatPlatform, `kubernetes`]
    await execa(`smctl`, args);

    args = [`provision`, btpOperatorInstance, `service-manager`, `service-operator-access`, `--mode=sync`]
    await execa(`smctl`, args);

    args = [`bind`, btpOperatorInstance, btpOperatorBinding, `--mode=sync`];
    await execa(`smctl`, args);
    args = [`get-binding`, btpOperatorBinding, `-o`, `json`];
    let out = await execa(`smctl`, args);
    let b = JSON.parse(out.stdout)
    let c = b.items[0].credentials
    //TODO figure out how to find clusterid
    return {clientId:c.clientid,clientSecret:c.clientsecret,smURL:c.sm_url,url:c.url,clusterId:"TODO"};
  } catch(error) {
    if (error.stderr === undefined) {
      throw new Error(`failed to process output of "smctl ${args.join(' ')}": ${error}`);
    }
    throw new Error(`failed "smctl ${args.join(' ')}": ${error.stderr}`);
  }
}

async function cleanupInstanceBinding(svcatPlatform, btpOperatorInstance, btpOperatorBinding) {
  let errors = [];
  let args = [];
  try {
    args = [`unbind`, btpOperatorInstance, btpOperatorBinding, `-f`];
    let x = await execa(`smctl`, args);
  } catch(error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }

  try {
    args = [`delete-platform`, svcatPlatform, `-f`];
    let x = await execa(`smctl`, args);
  } catch(error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }

  try {
    args = [`deprovision`, btpOperatorInstance, `-f`];
    let x = await execa(`smctl`, args);
  } catch(error) {
    errors = errors.concat([`failed "smctl ${args.join(' ')}": ${error.stderr}\n${error}`]);
  }
  if (errors.length > 0) {
    throw new Error(errors.join(", "));
  }
}

module.exports = {
  smInstanceBinding,
  cleanupInstanceBinding,
  installHelmCharts,
};
