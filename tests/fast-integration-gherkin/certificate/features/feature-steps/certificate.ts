import { Given, When, Then } from '@cucumber/cucumber';
const { getSecretData } = require('../../../../fast-integration/utils');
import { assert } from 'chai';
import forge from 'node-forge';

interface IContext {
	featureName: string;
	kymaGateWaySecret: any;
	certificate: forge.pki.Certificate;
	date: Date;
}

const context: IContext = {} as IContext;

Given(/^The "([^"]*)" secret is retrieved from "([^"]*)" namespace$/, async (args1,args2) => {
	context.featureName = "certificate-test";

	if (!context.kymaGateWaySecret){
		context.kymaGateWaySecret = await getSecretData(args1, args2);
	}
});

Then(/^Ingress certificate data should not be empty$/, () => {
	const gatewaySecret = context.kymaGateWaySecret

	assert.isNotEmpty(gatewaySecret, 'Ingress certificate should not be empty');
});

Given(/^The certificate is extracted from the secret data$/, () => {
	const gatewaySecret = context.kymaGateWaySecret;
	const tlsCert = gatewaySecret['tls.crt'];

	const certificate = forge.pki.certificateFromPem(tlsCert);
	context.certificate = certificate;
});

When(/^The date of today is set$/, () => {
	context.date = new Date();
	const todayDate = context.date;

	todayDate.setDate(todayDate.getDate() + 90);
	context.date = todayDate;
});

Then(/^The validity date of the certificate should be after the date of today$/, () => {
	const todayDate = context.date;
	const certificate = context.certificate;

	assert.isTrue(certificate.validity.notAfter >= todayDate, 'Certificate is going to outdate, please create new one');
});

Then(/^The validity date of the certificate should not be earlier than the date of today$/, () => {
	const certificate = context.certificate;

	assert.isTrue(certificate.validity.notBefore <= new Date(), 'Certificate is not yet valid');
});