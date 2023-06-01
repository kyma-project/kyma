const {expect} = require('chai');
const {
  getSecret,
  getSecretData,
  k8sDelete,
  waitForSecret,
  k8sApply,
  waitForK8sObject,
} = require('../../utils');
const {BTPOperatorCreds} = require('../../smctl/helpers');

const secretName = 'sap-btp-manager';
const ns = 'kyma-system';
const expectedBtpOperatorCreds = BTPOperatorCreds.dummy();

const reconciliationTimeout = 1000 * 70;
let secretFromProvisioning;
let modifiedSecret;

function btpManagerSecretTest() {
  describe('BTP Manager Secret Test', function() {
    // Check if BTP Manager Secret with BTP Operator credentials is created properly
    it('should check if Secret exists and contains the expected data keys', async function() {
      secretFromProvisioning = await getSecret(secretName, ns);
      checkSecretDataKeys(secretFromProvisioning);
      modifiedSecret = JSON.parse(JSON.stringify(secretFromProvisioning));
    });
    // Check if the Secret contains expected values
    it('should check if Secret data values match expected values', async function() {
      const actualSecretData = await getSecretData(secretName, ns);
      checkSecretDataValues(actualSecretData);
    });
    // Check if the Secret is properly reconciled after deletion
    it('should check if Secret is reconciled after deletion', async function() {
      console.log(`Deleting the "sap-btp-manager" Secret`);
      await k8sDelete([secretFromProvisioning], ns);
      console.log(`Waiting for the reconciliation for ${reconciliationTimeout} ms`);
      await waitForSecret(secretName, ns, reconciliationTimeout);
      console.log(`Secret has been re-created. Checking Secret's data`);
      const actualSecret = await getSecret(secretName, ns);
      checkSecretDataKeys(actualSecret);
      const actualSecretData = await getSecretData(secretName, ns);
      checkSecretDataValues(actualSecretData);
      console.log(`Secret has been properly reconciled`);
    });
    // Check if the Secret is properly reconciled after being edited
    it('should check if Secret is reconciled after being edited', async function() {
      console.log(`Changing data in the "sap-btp-manager" Secret`);
      prepareSecretForApply(modifiedSecret);
      changeSecretData(modifiedSecret);
      console.log(`Applying edited Secret`);
      await k8sApply([modifiedSecret], ns);
      console.log(`Waiting ${reconciliationTimeout} ms until edited Secret is created`);
      await waitForSecret(secretName, ns, reconciliationTimeout);
      let actualSecret = await getSecret(secretName, ns);
      console.log(`Waiting for the reconciliation for ${reconciliationTimeout} ms`);
      await waitForK8sObject(
          `/api/v1/namespaces/${ns}/secrets`,
          {},
          (_type, _apiObj, watchObj) => {
            return (
              watchObj.object.metadata.name.includes(secretName) &&
                        watchObj.object.metadata.resourceVersion !== actualSecret.metadata.resourceVersion
            );
          },
          reconciliationTimeout,
          `Waiting for ${secretName} Secret reconciliation timeout (${reconciliationTimeout} ms)`,
      );
      console.log(`Secret has been reconciled. Checking Secret's data`);
      actualSecret = await getSecret(secretName, ns);
      checkSecretDataKeys(actualSecret);
      const actualSecretData = await getSecretData(secretName, ns);
      checkSecretDataValues(actualSecretData);
      console.log(`Secret is correct`);
    });
  });
}

function checkSecretDataKeys(secret) {
  console.log(`Checking the data keys of the "sap-btp-manager" Secret`);
  expect(secret).to.not.be.empty;
  expect(secret.metadata.labels['app.kubernetes.io/managed-by']).to.equal('kcp-kyma-environment-broker');
  expect(secret.data).to.have.property('clientid');
  expect(secret.data).to.have.property('clientsecret');
  expect(secret.data).to.have.property('sm_url');
  expect(secret.data).to.have.property('tokenurl');
  expect(secret.data).to.have.property('cluster_id');
}

function checkSecretDataValues(secret) {
  console.log(`Checking data values of the "sap-btp-manager" Secret`);
  expect(secret.clientid).to.equal(expectedBtpOperatorCreds.clientid);
  expect(secret.clientsecret).to.equal(expectedBtpOperatorCreds.clientsecret);
  expect(secret.sm_url).to.equal(expectedBtpOperatorCreds.smURL);
  expect(secret.tokenurl).to.equal(expectedBtpOperatorCreds.url);
}

function prepareSecretForApply(secret) {
  delete secret.metadata.uid;
  delete secret.metadata.resourceVersion;
  delete secret.metadata.creationTimestamp;
  delete secret.metadata.annotations;
  delete secret.metadata.managedFields;
}

function changeSecretData(secret) {
  secret.data.clientid = Buffer.from('edited-clientid').toString('base64');
  secret.data.clientsecret = '';
  secret.data.sm_url = '';
  secret.data.tokenurl = Buffer.from('edited-tokenurl').toString('base64');
}

module.exports = {
  btpManagerSecretTest,
};
