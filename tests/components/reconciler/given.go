package reconciler

import "github.com/kyma-incubator/testdrape/godog/testing"

func createKCPCluster(t *testing.T) {
	t.Logf("Creating KCP cluster")
}

func createSKRCluster(t *testing.T) {
	t.Logf("Creating customer cluster")
}

func installReconciler(t *testing.T) {
	t.Logf("Install reconciler")
}

func createKymaCR(t *testing.T, centralMods, decentralMods int64, cluster string) {
	t.Logf("Create Kyma CR with %d centralised and %d dencentralized modules in cluster '%s'",
		centralMods, decentralMods, cluster)
}

func updateKymaCR(t *testing.T, centralMods, decentralMods int64, cluster string) {
	t.Logf("Updating Kyma CR by adding %d centralised and %d decentralized modules in cluster '%s'",
		centralMods, decentralMods, cluster)
}

func updateKymaCRInvalid(t *testing.T, cluster string) {
	t.Logf("Updating Kyma CR with invalid change in cluster '%s'", cluster)
}

func deleteKymaCR(t *testing.T, cluster string) {
	t.Logf("Delete Kyma CR in cluster '%s'", cluster)
}

func deleteWatcherWebhook(t *testing.T) {
	t.Logf("Delete watcher webhook")
}

func deleteModuleOperator(t *testing.T, cluster string) {
	t.Logf("Delete module operator in cluster '%s'", cluster)
}

func deleteModuleCRD(t *testing.T, cluster string) {
	t.Logf("Delete module CRD in cluster '%s'", cluster)
}

func updateModuleTemplate(t *testing.T, moduleType string) {
	t.Logf("Update %s module template", moduleType)
}
