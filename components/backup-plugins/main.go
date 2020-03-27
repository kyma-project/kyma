package main

import (
	"github.com/kyma-project/kyma/components/backup-plugins/internal/plugins"
	"github.com/sirupsen/logrus"
	veleroplugin "github.com/vmware-tanzu/velero/pkg/plugin/framework"
)

func main() {
	veleroplugin.NewServer().
		RegisterRestoreItemAction("kyma-project.io/instances-restore-plugin", newRemoveServiceInstanceFields).
		RegisterRestoreItemAction("kyma-project.io/secrets-restore-plugin", newSetOwnerReference).
		RegisterRestoreItemAction("kyma-project.io/nats-channels-restore-plugin", newIgnoreNatssChannelService).
		RegisterRestoreItemAction("kyma-project.io/knative-kyma-integration-restore-plugin", newKnativeKymaIntegration).
		RegisterRestoreItemAction("kyma-project.io/by-label-restore-plugin", newIgnoreByLabel).
		Serve()
}

func newRemoveServiceInstanceFields(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.RemoveServiceInstanceFields{Log: logger}, nil
}

func newSetOwnerReference(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.SetOwnerReference{Log: logger}, nil
}

func newIgnoreNatssChannelService(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.IgnoreNatssChannelService{Log: logger}, nil
}

func newIgnoreByLabel(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.IgnoreByLabel{Log: logger}, nil
}
func newKnativeKymaIntegration(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.IgnoreKnative{Log: logger}, nil
}
