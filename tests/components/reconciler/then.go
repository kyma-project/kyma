package reconciler

import "github.com/kyma-incubator/testdrape/godog/testing"

func assertModuleCR(t *testing.T, expected int64, cluster string) {
	t.Logf("Assert %d module CRs in cluster '%s'", expected, cluster)
}

func assertModuleCRNotExist(t *testing.T, cluster string) {
	assertModuleCR(t, 0, cluster)
}

func assertManifestCR(t *testing.T, expected int64, cluster string) {
	t.Logf("Assert %d manifest CRs in cluster '%s'", expected, cluster)
}

func assertManifestCRNotExist(t *testing.T, cluster string) {
	assertManifestCR(t, 0, cluster)
}

func assertModuleDeployed(t *testing.T, expected int64, cluster string) {
	t.Logf("Assert %d manifest CRs in cluster '%s'", expected, cluster)
}

func assertModuleUndeployed(t *testing.T, cluster string) {
	assertModuleDeployed(t, 0, cluster)
}

func assertKymaCRState(t *testing.T, state string, timeout int64) {
	t.Logf("Assert Kyma CR is in '%s' state within '%d' sec", state, timeout)
}

func assertKymaCRConditionsUpdated(t *testing.T, cluster string) {
	t.Logf("Assert Kyma CR conditions were updated in cluster '%s'", cluster)
}

func assertKymaCRCopied(t *testing.T, fromCluster, toCluster string) {
	t.Logf("Assert Kyma CR copied from '%s' to '%s' cluster", fromCluster, toCluster)
}

func assertKymaCREvent(t *testing.T, severity string) {
	t.Logf("Assert Kyma CR contains event with severity '%s'", severity)
}

func assertValidatingWebhookLog(t *testing.T, severity string) {
	t.Logf("Assert validating Webhook logs '%s'", severity)
}

func assertWatcherWebhookCreated(t *testing.T, cluster string) {
	t.Logf("Assert watcher Webhook created in '%s' cluster", cluster)
}

func assertSKRReconciled(t *testing.T) {
	t.Logf("Assert SKR cluster reconciled")
}

func assertModuleOperatorCreated(t *testing.T, timeout int, cluster string) {
	t.Logf("Assert model operator created within %d sec in '%s' cluster ", timeout, cluster)
}

func assertModuleCRDCreated(t *testing.T, timeout int, cluster string) {
	t.Logf("Assert module CRD created within %d sec in '%s' cluster ", timeout, cluster)
}

func assertModuleCRUpdated(t *testing.T, moduleType, cluster string) {
	t.Logf("Assert %s module CR created in '%s' cluster ", moduleType, cluster)
}
