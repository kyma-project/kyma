package reconciler

import (
	gotesting "testing"

	"go.uber.org/zap"

	"github.com/cucumber/godog"
	"github.com/kyma-incubator/testdrape/godog/testing"
)

func log(msg string, args ...interface{}) {
	zap.S().Infof(msg, args...)
}

func createKCPCluster(t *testing.T) {
	log("Creating KCP cluster")
}

func createSKRCluster(t *testing.T) {
	log("Creating customer cluster")
}

func installReconciler(t *testing.T) {
	log("Install reconciler")
}

func createKymaCR(t *testing.T, centralMods, decentralMods int64, cluster string) {
	log("Create Kyma CR with %d centralised and %d dencentralized modules in cluster '%s'",
		centralMods, decentralMods, cluster)
}

func updateKymaCR(t *testing.T, centralMods, decentralMods int64, cluster string) {
	log("Updating Kyma CR by adding %d centralised and %d decentralized modules in cluster '%s'",
		centralMods, decentralMods, cluster)
}

func deleteKymaCR(t *testing.T, cluster string) {
	log("Delete Kyma CR in cluster '%s'", cluster)
}

func updateKymaCRInvalid(t *testing.T, cluster string) {
	log("Updating Kyma CR with invalid change in cluster '%s'", cluster)
}

func assertModuleCR(t *testing.T, expected int64, cluster string) {
	log("Assert %d module CRs in cluster '%s'", expected, cluster)
}

func assertModuleCRNotExist(t *testing.T, cluster string) {
	assertModuleCR(t, 0, cluster)
}

func assertManifestCR(t *testing.T, expected int64, cluster string) {
	log("Assert %d manifest CRs in cluster '%s'", expected, cluster)
}

func assertManifestCRNotExist(t *testing.T, cluster string) {
	assertManifestCR(t, 0, cluster)
}

func assertModuleDeployed(t *testing.T, expected int64, cluster string) {
	log("Assert %d manifest CRs in cluster '%s'", expected, cluster)
}

func assertModuleUndeployed(t *testing.T, cluster string) {
	assertModuleDeployed(t, 0, cluster)
}

func assertKymaCRState(t *testing.T, state string, timeout int64) {
	log("Assert Kyma CR is in '%s' state within '%d' sec", state, timeout)
}

func assertKymaCRConditionsUpdated(t *testing.T, cluster string) {
	log("Assert Kyma CR conditions were updated in cluster '%s'", cluster)
}

func assertKymaCRCopied(t *testing.T, fromCluster, toCluster string) {
	log("Assert Kyma CR copied from '%s' to '%s' cluster", fromCluster, toCluster)
}

func assertKymaCREvent(t *testing.T, severity string) {
	log("Assert Kyma CR contains event with severity '%s'", severity)
}

func assertValidatingWebhookLog(t *testing.T, severity string) {
	log("Assert validating Webhook logs '%s'", severity)
}

func initializeScenarios(sCtx *godog.ScenarioContext) {
	// Pre-condition steps
	testing.NewContext(sCtx).Register(`^KCP cluster created$`, createKCPCluster)
	testing.NewContext(sCtx).Register(`^SKR cluster created$`, createSKRCluster)
	testing.NewContext(sCtx).Register(`^Kyma reconciler installed in KCP cluster$`, installReconciler)
	testing.NewContext(sCtx).Register(`^Kyma CR with (\d+) centralized modules? and (\d+) decentralized modules?`+
		` created in (\w+) cluster$`, createKymaCR)
	testing.NewContext(sCtx).Register(`^Kyma CR updated by setting (\d+) centralized modules?`+
		` and (\d+) decentralized modules? in (\w+) cluster$`, updateKymaCR)
	testing.NewContext(sCtx).Register(`^Kyma CR updated with invalid change in (\w+) cluster$`, updateKymaCRInvalid)
	testing.NewContext(sCtx).Register(`^Kyma CR deleted in (\w+) cluster$`, deleteKymaCR)

	testing.NewContext(sCtx).Register(`^(\d+) d?e?centralized module CRs? created in (\w+) cluster$`, assertModuleCR)
	testing.NewContext(sCtx).Register(`^(\d+) manifest CRs? created in (\w+) cluster$`, assertManifestCR)
	testing.NewContext(sCtx).Register(`^(\d+) modules? deployed in (\w+) cluster$`, assertModuleDeployed)
	testing.NewContext(sCtx).Register(`^module CRs? deleted in (\w+) cluster$`, assertModuleCRNotExist)
	testing.NewContext(sCtx).Register(`^manifest CRs? deleted in (\w+) cluster$`, assertManifestCRNotExist)
	testing.NewContext(sCtx).Register(`^modules? undeployed in (\w+) cluster$`, assertModuleUndeployed)
	testing.NewContext(sCtx).Register(`^Kyma CR in state (\w+) within (\d+)sec$`, assertKymaCRState)
	testing.NewContext(sCtx).Register(`^Kyma CR conditions updated in (\w+) cluster$`, assertKymaCRConditionsUpdated)
	testing.NewContext(sCtx).Register(`^Kyma CR copied from (\w+) to (\w+) cluster$`, assertKymaCRCopied)
	testing.NewContext(sCtx).Register(`^Kyma CR contains event with (\w+)`, assertKymaCREvent)
	testing.NewContext(sCtx).Register(`^Validating webhook logs (\w+)`, assertValidatingWebhookLog)
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
