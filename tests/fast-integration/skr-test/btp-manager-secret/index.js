const os = require('os');
const {expect} = require('chai');
const {
    initializeK8sClient,
    getSecret,
    getSecretData,
    k8sDelete,
    waitForSecret,
    k8sApply,
    waitForK8sObject,
} = require('../../utils');
const {BTPOperatorCreds} = require('../../smctl/helpers');

let suiteTimeout = 1000 * 60 * 5;
const secretName = 'sap-btp-manager';
const ns = 'kyma-system';
const expectedBtpOperatorCreds = BTPOperatorCreds.dummy();

let expectedSecret, modifiedSecret;

function btpManagerSecretTest() {
    describe('BTP Manager Secret Test', function () {
        this.timeout(suiteTimeout);
        before('REMOVE LATER: Initialize kubeconfig', async function(){
            await initializeK8sClient({kubeconfigPath: `${os.homedir()}/k3d_kubeconfig`});
        });
        // Check if BTP Manager Secret with BTP Operator credentials is created properly
        it('should check if Secret exists and contains the expected data keys', async function() {
            expectedSecret = await getSecret(secretName, ns);
            checkSecretDataKeys(expectedSecret)
            console.log(expectedSecret)
            modifiedSecret = JSON.parse(JSON.stringify(expectedSecret))
        });
        // Check if the Secret contains expected values
        it('should check if Secret data values match expected values', async function() {
            const actualSecretData = await getSecretData(secretName, ns);
            console.log(actualSecretData)
            expect(actualSecretData.clientid).to.equal('test_clientid');
            expect(actualSecretData.clientsecret).to.equal('test_clientsecret');
            expect(actualSecretData.sm_url).to.equal('test_sm_url');
            expect(actualSecretData.tokenurl).to.equal('test_tokenurl');
        });
        // Check if the Secret is properly reconciled after deletion
        it('should check if Secret is reconciled after deletion', async function() {
            console.log(`Deleting the "sap-btp-manager" Secret`);
            await k8sDelete([expectedSecret], ns);
            console.log(`Waiting for the reconciliation for 90s`);
            await waitForSecret(secretName, ns, 1000* 90);
            console.log(`Secret has been re-created. Checking Secret's data`);
            const actualSecret = await getSecret(secretName, ns);
            checkSecretDataKeys(actualSecret);
            const actualSecretData = await getSecretData(secretName, ns);
            expect(actualSecretData.clientid).to.equal('test_clientid');
            expect(actualSecretData.clientsecret).to.equal('test_clientsecret');
            expect(actualSecretData.sm_url).to.equal('test_sm_url');
            expect(actualSecretData.tokenurl).to.equal('test_tokenurl');
            console.log(`Secret has been properly reconciled`)
        });
        // Check if the Secret is properly reconciled after being edited
        it('should check if Secret is reconciled after being edited', async function() {
            prepareSecretForApply(modifiedSecret)
            changeSecretData(modifiedSecret)
            console.log(`Changing data in the "sap-btp-manager" Secret`);
            await k8sApply([modifiedSecret], ns)
            let actualSecret = await getSecret(secretName, ns);
            console.log(`Waiting for the reconciliation for 90s`);
            await waitForK8sObject(
                `/api/v1/namespaces/${ns}/secrets`,
                {},
                (_type, _apiObj, watchObj) => {
                    return (
                        watchObj.object.metadata.name.includes(secretName) &&
                        watchObj.object.metadata.resourceVersion !== actualSecret.metadata.resourceVersion
                    );
                },
                1000 * 90,
                `Waiting for ${secretName} Secret reconciliation timeout (90 s)`,
            );
            console.log(`Secret has been reconciled`);
            actualSecret = await getSecret(secretName, ns);
            checkSecretDataKeys(actualSecret);
            const actualSecretData = await getSecretData(secretName, ns);
            checkSecretDataValues(actualSecret);
            console.log(`Secret is correct`)
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