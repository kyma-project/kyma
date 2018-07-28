package backup_test

import (
	"errors"
	"testing"

	etcdTypes "github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/platform/logger/spy"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sTesting "k8s.io/client-go/testing"
)

var etcdbackupsResource = schema.GroupVersionResource{Group: "etcd.database.coreos.com", Version: "v1beta2", Resource: "etcdbackups"}

func createEtcdBackupAction(eb *etcdTypes.EtcdBackup) k8sTesting.Action {
	return k8sTesting.NewCreateAction(etcdbackupsResource, eb.Namespace, eb)
}

func deleteEtcdBackupAction(eb *etcdTypes.EtcdBackup) k8sTesting.Action {
	return k8sTesting.NewDeleteAction(etcdbackupsResource, eb.Namespace, eb.Name)
}

func filterOutInformerActions(actions []k8sTesting.Action) []k8sTesting.Action {
	var ret []k8sTesting.Action
	for _, action := range actions {
		if action.GetVerb() == "list" || action.GetVerb() == "watch" {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func containsAction(expected k8sTesting.Action, gotActions []k8sTesting.Action) bool {
	for _, actual := range gotActions {
		switch a := actual.(type) {
		case k8sTesting.CreateAction:
			e, ok := expected.(k8sTesting.CreateAction)
			if !ok {
				continue
			}

			expObj := e.GetObject()
			actualObj := a.GetObject()

			equal := assert.ObjectsAreEqualValues(expObj, actualObj) &&
				e.Matches(a.GetVerb(), a.GetResource().Resource)
			if equal {
				return true
			}

		case k8sTesting.DeleteAction:
			e, ok := expected.(k8sTesting.DeleteAction)
			if !ok {
				continue
			}

			equal := assert.ObjectsAreEqual(a.GetName(), e.GetName()) &&
				e.Matches(a.GetVerb(), a.GetResource().Resource)
			if equal {
				return true
			}
		}
	}

	return false
}

func failingRector(verb, resource string) (string, string, k8sTesting.ReactionFunc) {
	failingFn := func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, errors.New("custom error")
	}
	return verb, resource, failingFn
}

func newLogSinkForErrors() *spy.LogSink {
	logSink := spy.NewLogSink()
	logSink.Logger.Logger.Level = logrus.ErrorLevel
	return logSink
}

func assertErrorContainsStatement(t *testing.T, err error, contains string) {
	assert.Contains(t, err.Error(), contains)
}
