const installer = require("../installer/helm");
const utils = require("../utils");
const k8s = require("@kubernetes/client-node");

async function installHelmCharts(creds) {
  const btpChart = "https://github.com/kyma-incubator/sap-btp-service-operator/releases/download/v0.1.9-custom/sap-btp-operator-custom-0.1.9.tar.gz";
  const btp = "sap-btp-operator";
  const btpValues = `manager.secret.clientid=${creds.clientId},manager.secret.clientsecret=${creds.clientSecret},manager.secret.url=${creds.smURL},manager.secret.tokenurl=${creds.url},cluster.id=${creds.clusterId}`
  namespace = {
    metadata: {
      name: btp,
    },
  };
  try {
    await utils.k8sCoreV1Api.createNamespace(namespace);
  } catch(err) {
    throw new Error(`failed to create namespace: ${btp} - received ${err.statusCode}`);
  }
  await installer.helmInstallUpgrade(btp, btpChart, btp, btpValues, null);
}

async function smInstanceBinding() {
  return {clientId:"cid",clientSecret:"ces",smURL:"smurl",url:"url",clusterId:"clusterId"}
}

module.exports = {
  smInstanceBinding,
  installHelmCharts,
};
