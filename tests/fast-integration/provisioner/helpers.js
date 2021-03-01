const { expect } = require("chai");
const { fs } = require("fs");

class ProvisionerConfig {
  constructor() {
    this.provisionerHost = process.env["PROVISIONER_HOST"] || "";
    this.clientID = process.env["PROVISIONER_CLIENT_ID"] || "";
    this.clientSecret = process.env["PROVISIONER_CLIENT_SECRET"] || "";
    this.tenantID = process.env["PROVISIONER_TENANT"] || "";
  }
}

async function getKubeconfig(provisioner, runtimeID) {
  const response = await provisioner.runtimeStatus(runtimeID);
  expect(response).to.have.nested.property(
    "data.runtimeStatus.runtimeConfiguration.kubeconfig"
  );
  kubeconfig = response.data.runtimeStatus.runtimeConfiguration.kubeconfig;
  fs.writeFile("kubeconfig.yaml", kubeconfig);
}

module.exports = {
  ProvisionerConfig,
  getKubeconfig,
};
