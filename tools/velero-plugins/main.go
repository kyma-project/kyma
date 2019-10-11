package main

import (
	veleroplugin "github.com/heptio/velero/pkg/plugin/framework"
	"github.com/sirupsen/logrus"
	"github.com/kyma-project/kyma/tools/velero-plugins/internal/plugins"
)

func main() {
	veleroplugin.NewServer().
		RegisterRestoreItemAction("kyma-project.io/instances-restore-plugin", newRemoveServiceInstanceFields).
		RegisterRestoreItemAction("kyma-project.io/secrets-restore-plugin", newSetOwnerReference).
		Serve()
}

func newRemoveServiceInstanceFields(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.RemoveServiceInstanceFields{Log: logger}, nil
}

func newSetOwnerReference(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.SetOwnerReference{Log: logger}, nil
}
