package reconciler

import (
	gotesting "testing"

	"github.com/cucumber/godog"
	"github.com/kyma-incubator/testdrape/godog/testing"
)

func initializeScenarios(sCtx *godog.ScenarioContext) {
	//Pre-condition steps
	testing.NewContext(sCtx).Register(`^KCP cluster created$`, createKCPCluster)
	testing.NewContext(sCtx).Register(`^SKR cluster created$`, createSKRCluster)
	testing.NewContext(sCtx).Register(`^Kyma reconciler installed in KCP cluster$`, installReconciler)
	testing.NewContext(sCtx).Register(`^Kyma CR with {int} centralized modules? and {int} decentralized modules?`+
		` created in {word} cluster$`, createKymaCR)
	testing.NewContext(sCtx).Register(`^Kyma CR updated by setting {int} centralized modules?`+
		` and {int} decentralized modules? in {word} cluster$`, updateKymaCR)
	testing.NewContext(sCtx).Register(`^Kyma CR updated with invalid change in {word} cluster$`, updateKymaCRInvalid)
	testing.NewContext(sCtx).Register(`^Kyma CR deleted in {word} cluster$`, deleteKymaCR)
	testing.NewContext(sCtx).Register(`^watcher webhook deleted$`, deleteWatcherWebhook)
	testing.NewContext(sCtx).Register(`^module operator deleted in {word} cluster$`, deleteModuleOperator)
	testing.NewContext(sCtx).Register(`^module CRDs? deleted in {word} cluster$`, deleteModuleCRD)
	testing.NewContext(sCtx).Register(`^{word} module template updated$`, updateModuleTemplate)

	//Assertions
	testing.NewContext(sCtx).Register(`^{int} d?e?centralized module CRs? created in {word} cluster$`, assertModuleCR)
	testing.NewContext(sCtx).Register(`^{int} manifest CRs? created in {word} cluster$`, assertManifestCR)
	testing.NewContext(sCtx).Register(`^{int} modules? deployed in {word} cluster$`, assertModuleDeployed)
	testing.NewContext(sCtx).Register(`^module CRs? deleted in {word} cluster$`, assertModuleCRNotExist)
	testing.NewContext(sCtx).Register(`^manifest CRs? deleted in {word} cluster$`, assertManifestCRNotExist)
	testing.NewContext(sCtx).Register(`^modules? undeployed in {word} cluster$`, assertModuleUndeployed)
	testing.NewContext(sCtx).Register(`^Kyma CR in state {word} within {int}sec$`, assertKymaCRState)
	testing.NewContext(sCtx).Register(`^Kyma CR conditions updated in {word} cluster$`, assertKymaCRConditionsUpdated)
	testing.NewContext(sCtx).Register(`^Kyma CR copied from {word} to {word} cluster$`, assertKymaCRCopied)
	testing.NewContext(sCtx).Register(`^Kyma CR contains event with {word}`, assertKymaCREvent)
	testing.NewContext(sCtx).Register(`^Validating webhook logs {word}`, assertValidatingWebhookLog)
	testing.NewContext(sCtx).Register(`^watcher webhook created within {int}sec$`, assertWatcherWebhookCreated)
	testing.NewContext(sCtx).Register(`^SKR cluster reconciled within {int}sec$`, assertSKRReconciled)
	testing.NewContext(sCtx).Register(`^module operator created within {int}sec in {word} cluster$`,
		assertModuleOperatorCreated)
	testing.NewContext(sCtx).Register(`^module CRD created within {int}sec in {word} cluster$`,
		assertModuleCRDCreated)
	testing.NewContext(sCtx).Register(`{word} module CRs? updated in {word} cluster`,
		assertModuleCRUpdated)
}

func TestFeatures(testT *gotesting.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: initializeScenarios,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: testT,
		},
	}

	if suite.Run() != 0 {
		testT.Error("non-zero status returned, failed to run feature tests")
	}
}
