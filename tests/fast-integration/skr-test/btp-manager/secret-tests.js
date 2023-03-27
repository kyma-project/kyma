const { expect } = require("chai");

function btpManagerSecretTests() {
    describe('Test BTP Manager Secret', function() {
        // Check if the secret with BTP credentials is created properly
        it('should check if secret exists and contains the expected data', async function() {
            console.log(`Checking the contents of the "sap-btp-manager" secret`);
            const secretToCheck = await getSecret('sap-btp-manager', 'kyma-system');
            expect(secretToCheck).to.not.be.empty;
            expect(secretToCheck.metadata.labels['app.kubernetes.io/managed-by']).to.equal('kcp-kyma-environment-broker');
            expect(secretToCheck.data).to.have.property('clientid');
            expect(secretToCheck.data).to.have.property('clientsecret');
            expect(secretToCheck.data).to.have.property('sm_url');
            expect(secretToCheck.data).to.have.property('tokenurl');
            expect(secretToCheck.data).to.have.property('cluster_id');
          });

        it('simulate secret deletion', async function(){
            try {
                await k8sDeleteSecret("sap-btp-manager", "secret", "opaque")
            }
            catch(e) {
                expect(e).to.be.not.empty;
            }

        })
    });
}

module.exports = {
    btpManagerSecretTests,
};