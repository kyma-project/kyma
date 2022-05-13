const {Given, When, Then, AfterAll} = require('cucumber');  
const {expect} = require('chai');
const {SKRSetup} = require('../../../helpers/skr-setup');
const {CommerceCompassMock} = require('../../../helpers/commerce-mock-with-compass-setup');
const {
    debug,
    ensureKymaAdminBindingDoesNotExistsForUser,
    ensureKymaAdminBindingExistsForUser,
    isDebugEnabled,
    printStatusOfInClusterEventingInfrastructure
} = require('../../../helpers/utils');
const {
    ensureValidShootOIDCConfig,
    ensureValidOIDCConfigInCustomerFacingKubeconfig,
    ensureOperationSucceeded
} = require('../../../../fast-integration/kyma-environment-broker');
const {keb, gardener} = require('../../../../fast-integration/skr-test/provision/provision-skr');
const {
    callFunctionWithToken,
    assertSuccessfulFunctionResponse,
    assertUnauthorizedFunctionResponse,
    callFunctionWithNoToken,
    sendLegacyEvent,
    checkLegacyEventResponse,
    getRandomEventId,
    getVirtualServiceHost,
    sendInClusterEventWithRetry
} = require('../../../../fast-integration/test/fixtures/commerce-mock');

this.context = new Object();

Given(/^SKR is provisioned$/, {timeout: 60 * 60 * 1000 * 3}, async () => {
    this.context.featureName = "skr-test";
	await SKRSetup.provisionSKR();

    const options = SKRSetup.options;
    const shoot = SKRSetup.shoot;

    this.context.options = options;
    this.context.shoot = shoot;
});

Then(/^"([^"]*)" OIDC config is applied on the shoot cluster$/, (oidcConfig) => {
    const shoot = this.context.shoot;
    const options = this.context.options;
    let oidc = options.oidc0;
    if (oidcConfig !== 'Initial'){
        oidc = options.oidc1;
    }

    ensureValidShootOIDCConfig(shoot, oidc);
});

Then(/^"([^"]*)" OIDC config is part of the kubeconfig$/, async (oidcConfig) => {
    const options = this.context.options;
    let oidc = options.oidc0;
    if (oidcConfig !== 'Initial'){
        oidc = options.oidc1;
    }

	await ensureValidOIDCConfigInCustomerFacingKubeconfig(keb, options.instanceID, oidc);
});

Then(/^Admin binding exists for "([^"]*)" user$/, async(userAdmin) => {
	const options = this.context.options;
    const admins = userAdmin === 'old' ? [options.administrator0]: options.administrators1;

    admins.forEach(async (admin) => {
        await ensureKymaAdminBindingExistsForUser(admin)
    });
});

When(/^SKR service is updated$/, async() => {
    const options = this.context.options;
    const customParams = {
        oidc: options.oidc1,
    };

	await SKRSetup.updateSKR(options.instanceID, customParams, false);

    this.context.updateSkrResponse = SKRSetup.updateSkrResponse;
});

Then(/^The operation response should have a succeeded state$/, {timeout: 1000 * 60 * 20}, async() => {
	const updateSkrResponse = this.context.updateSkrResponse;
    const kcp = SKRSetup.kcp;
    const instanceID = this.context.options.instanceID;
    const shootName = this.context.shoot.name;
    const updateTimeout = 1000 * 60 * 20; // 20m

    expect(updateSkrResponse).to.have.property('operation');

    const operationID = updateSkrResponse.operation;
    debug(`Operation ID ${operationID}`);

    await ensureOperationSucceeded(keb, kcp, instanceID, operationID, updateTimeout);

    const shoot = await gardener.getShoot(shootName);

    this.context.operationID = operationID;
    this.context.shoot = shoot;
});

Then(/^Runtime status should be fetched successfully$/, async() => {
    const options = this.context.options;

	try {
      const runtimeStatus = await kcp.getRuntimeStatusOperations(options.instanceID);
      console.log(`\nRuntime status: ${runtimeStatus}`);
      await kcp.reconcileInformationLog(runtimeStatus);
    } catch (e) {
      console.log(`before hook failed: ${e.toString()}`);
    }
});

When(/^The admins for the SKR service are updated$/, async() => {
	const options = this.context.options;
    const customParams = {
        oidc: options.administrators1,
    };

    console.log("customParams", customParams);

	await SKRSetup.updateSKRAdmins(options.instanceID, customParams, false);

    this.context.updateSkrAdminsResponse = SKRSetup.updateSkrAdminsResponse;
});

Then(/^The old admin no longer exists for the SKR service instance$/, async() => {
    const options = this.context.options;

    await ensureKymaAdminBindingDoesNotExistsForUser(options.administrator0);
});

Given(/^Commerce Backend is set up$/, async() => {
	const options = this.context.options;

    await CommerceCompassMock.ensureCommerceWithCompassMockIsSetUp(options);
});

When(/^Function is called using a correct authorization token$/, async() => {
    const options = this.context.options;

	const successfulFunctionResponse = await callFunctionWithToken(options.testNS);

    this.context.successfulFunctionResponse = successfulFunctionResponse;
});

Then(/^The function should be reachable$/, () => {
    const successfulFunctionResponse = this.context.successfulFunctionResponse;

	assertSuccessfulFunctionResponse(successfulFunctionResponse);
});

When(/^Function is called without an authorization token$/, async() => {
	const error = await callFunctionWithNoToken();

    this.context.unauthorizedFunctionResponse = error;
});

Then(/^The function returns an error$/, () => {
    const unauthorizedFunctionResponse = this.context.unauthorizedFunctionResponse;

	assertUnauthorizedFunctionResponse(unauthorizedFunctionResponse);
});

When(/^A legacy event is sent$/, async() => {
	const legacyEventResponse = await sendLegacyEvent();

    this.context.legacyEventResponse = legacyEventResponse;
});


Then(/^The event should be received correctly$/, () => {
    const legacyEventResponse = this.context.legacyEventResponse;

	checkLegacyEventResponse(legacyEventResponse);
});

When(/^An in-cluster "([^"]*)" event is sent$/, {timeout: 60 * 60 * 1000}, async(eventEncoding) => {
    const targetNamespace = this.context.options.testNS;

	const eventId = getRandomEventId(eventEncoding);
    const mockHost = await getVirtualServiceHost(targetNamespace, 'lastorder');

    if (isDebugEnabled()) {
        await printStatusOfInClusterEventingInfrastructure(targetNamespace, eventEncoding, 'lastorder');
    }

    await sendInClusterEventWithRetry(mockHost, eventId, eventEncoding);

    this.context.lastOrderMockHost = mockHost;
    this.context.eventId = eventId;
});

Then(/^The event is received successfully$/, () => {
    const mockHost = this.context.mockHost;
    const eventId = this.context.eventId;

	ensureInClusterEventReceivedWithRetry(mockHost, eventId);
});

AfterAll({timeout: 1000 * 60 * 95}, async() => {
    const featureName = this.context.featureName;

    // if (featureName === "skr-test"){
    //     const options = this.context.options;

    //     // Delete commerce mock
    //     await CommerceCompassMock.deleteCommerceMockResources(options.testNS);

    //     // Deprovision SKR
    //     await SKRSetup.deprovisionSKR();    
    // }
});
