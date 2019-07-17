package main

import (
	veleroplugin "github.com/heptio/velero/pkg/plugin/framework"
	"github.com/kyma-project/kyma/tools/velero-plugins/internal/plugins"
	"github.com/sirupsen/logrus"
)

func main() {
	veleroplugin.NewServer().
		RegisterRestoreItemAction("kyma-project.io/si-restore-plugin", newRemoveServiceInstanceFields).
		RegisterRestoreItemAction("kyma-project.io/or-restore-plugin", newSetOwnerReference).
		Serve()
}

func newRemoveServiceInstanceFields(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.RemoveServiceInstanceFields{Log: logger}, nil
}

func newSetOwnerReference(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.SetOwnerReference{Log: logger}, nil
}
