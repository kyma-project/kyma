package main

import (
	arkplugin "github.com/heptio/ark/pkg/plugin"
	"github.com/kyma-project/kyma/tools/ark-plugins/internal/plugins"
	"github.com/sirupsen/logrus"
)

func main() {
	arkplugin.NewServer(arkplugin.NewLogger()).
		RegisterRestoreItemAction("si-restore-plugin", newRemoveServiceInstanceFields).
		RegisterRestoreItemAction("or-restore-plugin", newSetOwnerReference).
		Serve()
}

func newRemoveServiceInstanceFields(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.RemoveServiceInstanceFields{Log: logger}, nil
}

func newSetOwnerReference(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.SetOwnerReference{Log: logger}, nil
}
