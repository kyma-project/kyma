const os = require('os');
const {expect} = require('chai');
const {
    initializeK8sClient,
    getSecret,
} = require('../../utils');

const secretName = 'sap-btp-manager';
const ns = 'kyma-system';

let expectedSecret, modifiedSecret;

function btpManagerSecretTest() {
    describe('BTP Manager Secret Test', function () {
        before('REMOVE LATER: Initialize kubeconfig', async function(){
            await initializeK8sClient({kubeconfigPath: `${os.homedir()}/k3d_kubeconfig`});
        });
        // Check if BTP Manager Secret with BTP Operator credentials is created properly
        it('should check if Secret exists and contains the expected data keys', async function() {
            console.log(`Checking the data keys of the "sap-btp-manager" Secret`);
            expectedSecret = await getSecret(secretName, ns);
            expect(expectedSecret).to.not.be.empty;
            expect(expectedSecret.metadata.labels['app.kubernetes.io/managed-by']).to.equal('kcp-kyma-environment-broker');
            expect(expectedSecret.data).to.have.property('clientid');
            expect(expectedSecret.data).to.have.property('clientsecret');
            expect(expectedSecret.data).to.have.property('sm_url');
            expect(expectedSecret.data).to.have.property('tokenurl');
            expect(expectedSecret.data).to.have.property('cluster_id');
            modifiedSecret = JSON.parse(JSON.stringify(expectedSecret))
        });
    });
}

module.exports = {
    btpManagerSecretTest,
};