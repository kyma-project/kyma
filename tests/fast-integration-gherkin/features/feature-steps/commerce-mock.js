const {CommerceMock} = require('../../helpers/commerce-mock-setup');
const {kubectlPortForward} = require('fast-integration-tests');
const {
    sendInClusterEvent,
    verifyInClusterEventIsReceivedCorrectly
} = require('../../../fast-integration/test/fixtures/commerce-mock')
const {Given, When, Then} = require('cucumber');

this.context = new Object();

Given(/^All pods in the cluster are listed$/, async() => {
	const initialPodsRestarts = await CommerceMock.listAllPodsInCluster();

    this.context.initialPodsRestarts = initialPodsRestarts;
});

Given(/^The Commerce backend is set up$/, async() => {
    this.context.testNamespace = 'test';
    this.context.withCentralAppConnectivity = (process.env.WITH_CENTRAL_APP_CONNECTIVITY === 'true');

    const testNamespace = this.context.testNamespace;
    const withCentralAppConnectivity = this.context.withCentralAppConnectivity;

	await CommerceMock.ensureCommerceMockIsSetUp('mocks', testNamespace, withCentralAppConnectivity);
});

Given(/^Loki port is forwarded$/, () => {
    this.context.lokiPort = 3100;

    const lokiPort = this.context.lokiPort;
	kubectlPortForward('kyma-system', 'logging-loki-0', lokiPort);
});

When(/^A {string} event is sent$/, async (eventEncoding) => {
	const testNamespace = this.context.testNamespace;

    const inClusterEventDetails = await sendInClusterEvent(testNamespace, eventEncoding);

    this.context.lastOrderMockHost = inClusterEventDetails.mockHost;
    if(eventEncoding == 'structured'){
        this.context.structuredEventId = inClusterEventDetails.eventId;
    }else{
        this.context.binaryEventId = inClusterEventDetails.eventId;
    }
});

Then(/^The {string} event should be received correctly$/, async (eventEncoding) => {
    let lastOrderMockHost = this.context.lastOrderMockHost;
    let eventId;
	if(eventEncoding == 'structured'){
        eventId = this.context.structuredEventId;
    }else{
        eventId = this.context.binaryEventId;
    }

    await verifyInClusterEventIsReceivedCorrectly(lastOrderMockHost, eventId);
});
