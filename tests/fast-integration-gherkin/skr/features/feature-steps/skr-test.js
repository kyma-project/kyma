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
} = require('../../../../fast-integration/utils');
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
    sendEvent,
    checkEventResponse,
    getRandomEventId,
    getVirtualServiceHost,
    sendInClusterEventWithRetry,
    getCommerceMockHost,
    getLegacyEventParams,
    getStructuredEventParams,
    getBinaryEventParams,
    ensureInClusterEventReceivedWithRetry
} = require('../../../../fast-integration/test/fixtures/commerce-mock');
const {
    AuditLogCreds,
    AuditLogClient,
    checkAuditLogs,
} = require('../../../../fast-integration/audit-log');
const { gatherOptions } = require('../../../../fast-integration/skr-test');

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
    const admins = userAdmin === 'old' ? [options.kebUserId]: options.administrators1;

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
    const shoot = await gardener.getShoot(this.context.shoot.name);

    this.context.updateSkrResponse = SKRSetup.updateSkrResponse;
    this.context.shoot = shoot;
});

Then(/^The update skr "([^"]*)" operation response should have a succeeded state$/, {timeout: 1000 * 60 * 20}, async(updateAdmins) => {
	let updateSkrResponse = this.context.updateSkrResponse;
    if (updateAdmins === 'admins'){
        updateSkrResponse = this.context.updateSkrAdminsResponse;
    }

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
    const kcp = SKRSetup.kcp;

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
        administrators: options.administrators1,
    };

	await SKRSetup.updateSKRAdmins(options.instanceID, customParams, false);
    const shoot = await gardener.getShoot(this.context.shoot.name);

    this.context.updateSkrAdminsResponse = SKRSetup.updateSkrAdminsResponse;
    this.context.shoot = shoot;
});

Then(/^The old admin no longer exists for the SKR service instance$/, async() => {
    const options = this.context.options;

    await ensureKymaAdminBindingDoesNotExistsForUser(options.kebUserId);
});

Given(/^Commerce Backend is set up$/, {timeout: 1000 * 60 * 60 * 3}, async() => {
	const options = this.context.options;

    await CommerceCompassMock.ensureCommerceWithCompassMockIsSetUp(options);

    const hostValues = await getCommerceMockHost();
    this.context.commerceHostValues = hostValues;
});

When(/^Function is called using a correct authorization token$/, {timeout: 1000 * 60 * 60 * 2}, async() => {
    const options = this.context.options;
    const hostValues = this.context.commerceHostValues;

	const successfulFunctionResponse = await callFunctionWithToken(options.testNS, hostValues.host);

    this.context.successfulFunctionResponse = successfulFunctionResponse;
});

Then(/^The function should be reachable$/, () => {
    const successfulFunctionResponse = this.context.successfulFunctionResponse;

	assertSuccessfulFunctionResponse(successfulFunctionResponse);
});

When(/^Function is called without an authorization token$/, async() => {
    const commerceHostValues = this.context.commerceHostValues;

	const unauthorizedFunctionResponse = await callFunctionWithNoToken(commerceHostValues.host);

    this.context.unauthorizedFunctionResponse = unauthorizedFunctionResponse;
});

Then(/^The function returns an error$/, () => {
    const unauthorizedFunctionResponse = this.context.unauthorizedFunctionResponse;

	assertUnauthorizedFunctionResponse(unauthorizedFunctionResponse);
});

When(/^A "([^"]*)" event is sent$/, {timeout: 60 * 60 * 1000}, async(eventEncoding) => {
    const commerceHost = this.context.commerceHostValues.host;
    const commerceMockHost = this.context.commerceHostValues.mockHost;

    let requestParams = null;
    if (eventEncoding === 'legacy'){
        requestParams = getLegacyEventParams();
    } else if (eventEncoding === 'structured'){
        requestParams = getStructuredEventParams();
    } else if (eventEncoding === 'binary'){
        requestParams = getBinaryEventParams();
    } else {
        console.error("Not supported eventEncoding type:", eventEncoding);
    }
	const eventResponse = await sendEvent(commerceMockHost, commerceHost, requestParams);

    this.context.eventResponse = eventResponse;
});

Then(/^The event should be received correctly$/, () => {
    const eventResponse = this.context.eventResponse;

	checkEventResponse(eventResponse);
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

Then(/^The event is received successfully$/, {timeout: 60 * 60 * 1000}, async() => {
    const mockHost = this.context.lastOrderMockHost;
    const eventId = this.context.eventId;

	await ensureInClusterEventReceivedWithRetry(mockHost, eventId);
});

Given(/^KEB plan is AWS$/, () => {
    const AWS_PLAN_ID = '361c511f-f939-4621-b228-d0fb79a1fe15';
    let auditLogs = null;

	if (process.env.KEB_PLAN_ID === AWS_PLAN_ID) {
        auditLogs = new AuditLogClient(AuditLogCreds.fromEnv());
    }

    this.context.auditLogs = auditLogs;
});

Then(/^Audit logs should be available$/, {timeout: 1000 * 60 * 60}, async() => {
	const auditLogs = this.context.auditLogs;

    if (auditLogs !== null){
        await checkAuditLogs(auditLogs, null);
    }
});

AfterAll({timeout: 1000 * 60 * 95}, async() => {
    const featureName = this.context.featureName;

    if (featureName === "skr-test"){
        // const options = this.context.options;

        // Delete commerce mock
        // await CommerceCompassMock.deleteCommerceMockResources(options.testNS);

        // Deprovision SKR
        await SKRSetup.deprovisionSKR();    
    }
});
