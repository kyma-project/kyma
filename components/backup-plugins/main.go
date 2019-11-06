package main

import (
	veleroplugin "github.com/heptio/velero/pkg/plugin/framework"
	"github.com/kyma-project/kyma/components/backup-plugins/internal/plugins"
	"github.com/sirupsen/logrus"
)

func main() {
	veleroplugin.NewServer().
		RegisterRestoreItemAction("kyma-project.io/instances-restore-plugin", newRemoveServiceInstanceFields).
		RegisterRestoreItemAction("kyma-project.io/secrets-restore-plugin", newSetOwnerReference).
		RegisterRestoreItemAction("kyma-project.io/nats-channels-restore-plugin", newIgnoreNatssChannelService).
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
