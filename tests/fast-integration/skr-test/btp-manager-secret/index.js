const os = require('os');
const {expect} = require('chai');

const {
    initializeK8sClient,
} = require('../../utils');

function btpManagerSecretTest() {
    describe('BTP Manager Secret Test', function () {
        before('REMOVE LATER: Initialize kubeconfig', async function(){
            await initializeK8sClient({kubeconfigPath: `${os.homedir()}/k3d_kubeconfig`});
        });
    });
}

module.exports = {
    btpManagerSecretTest,
};