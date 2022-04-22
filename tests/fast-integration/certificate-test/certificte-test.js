const {getSecretData} = require('../utils');
const {assert} = require('chai');
const forge = require('node-forge');

describe('Certificate test', async function() {
  this.timeout(3000);
  this.slow(3 * 1000);
  it('Checking if installed ingress gateway certificate is valid', async () => {
    await checkDefaultCertificateIsValid();
  });
});

async function checkDefaultCertificateIsValid() {
  const kymaGateWaySecret = await getSecretData('kyma-gateway-certs', 'istio-system');
  assert.isNotEmpty(kymaGateWaySecret, 'Ingress certificate should not be empty');
  const tlsCert = kymaGateWaySecret['tls.crt'];
  const cert = forge.pki.certificateFromPem(tlsCert);
  const date = new Date();
  date.setDate(date.getDate() + 90);
  assert.isAtLeast(cert.validity.notAfter, date,
      'Certificate will expire in less than 3 months, please create a new one');
  assert.isAtMost(cert.validity.notBefore, new Date(), 'Certificate is not yet valid');
}
