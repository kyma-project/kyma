package main

import (
	arkplugin "github.com/heptio/ark/pkg/plugin"
	"github.com/kyma-project/kyma/tools/ark-plugins/internal/plugins"
	"github.com/kyma-project/kyma/tools/ark-plugins/internal/plugins/backup"
	"github.com/kyma-project/kyma/tools/ark-plugins/internal/plugins/restore"
	"github.com/sirupsen/logrus"
)

func main() {
	arkplugin.NewServer(arkplugin.NewLogger()).
		RegisterRestoreItemAction("si-restore-plugin", newRemoveServiceInstanceFields).
		RegisterRestoreItemAction("or-restore-plugin", newSetOwnerReference).
		RegisterRestoreItemAction("function-restore-plugin", newRestoreFunction).
		RegisterBackupItemAction("function-backup-plugin", newBackupFunction).
		Serve()
}

func newRemoveServiceInstanceFields(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.RemoveServiceInstanceFields{Log: logger}, nil
}

func newSetOwnerReference(logger logrus.FieldLogger) (interface{}, error) {
	return &plugins.SetOwnerReference{Log: logger}, nil
}

func newRestoreFunction(logger logrus.FieldLogger) (interface{}, error) {
	return &backup.FunctionPlugin{Log: logger}, nil
}

func newBackupFunction(logger logrus.FieldLogger) (interface{}, error) {
	return &restore.FunctionPluginRestore{Log: logger}, nil
}