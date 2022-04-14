const {Given, When, Then} = require('cucumber');  
const {getSecretData} = require('../../../fast-integration/utils');
const {assert} = require('chai');
const forge = require('node-forge');

this.context = new Object();

Given(/^The "([^"]*)" secret is retrieved from "([^"]*)" namespace$/, async (args1,args2) => {
	if (!this.context.kymaGateWaySecret){
		this.context.kymaGateWaySecret = await getSecretData(args1, args2);
	}
});

Then(/^Ingress certificate data should not be empty$/, () => {
	const gatewaySecret = this.context.kymaGateWaySecret

	assert.isNotEmpty(gatewaySecret, 'Ingress certificate should not be empty');
});

Given(/^The certificate is extracted from the secret data$/, () => {
	const gatewaySecret = this.context.kymaGateWaySecret;
	const tlsCert = gatewaySecret['tls.crt'];

	const certificate = forge.pki.certificateFromPem(tlsCert);
	this.context.certificate = certificate;
});

When(/^The date of today is set$/, () => {
	this.context.date = new Date();
	const todayDate = this.context.date;

	todayDate.setDate(todayDate.getDate() + 90);
	this.context.date = todayDate;
});

Then(/^The validity date of the certificate should be after the date of today$/, () => {
	const todayDate = this.context.date;
	const certificate = this.context.certificate;

	assert.isTrue(certificate.validity.notAfter >= todayDate, 'Certificate is going to outdate, please create new one');
});

Then(/^The validity date of the certificate should not be earlier than the date of today$/, () => {
	const certificate = this.context.certificate;

	assert.isTrue(certificate.validity.notBefore <= new Date(), 'Certificate is not yet valid');
});
