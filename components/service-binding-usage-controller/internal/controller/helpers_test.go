package controller_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/platform/logger/spy"
	sbuTypes "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	svcatSettings "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/settings/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sTesting "k8s.io/client-go/testing"
)

var deploymentsResource = schema.GroupVersionResource{Group: "apps", Version: "v1beta2", Resource: "deployments"}

var podpresetsResource = schema.GroupVersionResource{Group: "settings.k8s.io", Version: "v1alpha1", Resource: "podpresets"}

func createPodPresetAction(pp *svcatSettings.PodPreset) k8sTesting.Action {
	return k8sTesting.NewCreateAction(podpresetsResource, pp.Namespace, pp)
}

func deletePodPresetAction(pp *svcatSettings.PodPreset) k8sTesting.Action {
	return k8sTesting.NewDeleteAction(podpresetsResource, pp.Namespace, pp.Name)
}

var servicebindingusagesResource = schema.GroupVersionResource{Group: "servicecatalog.ysf.io", Version: "v1alpha1", Resource: "servicebindingusages"}

func updateUsageAction(sbu *sbuTypes.ServiceBindingUsage) k8sTesting.UpdateAction {
	return k8sTesting.NewUpdateAction(servicebindingusagesResource, sbu.Namespace, sbu)
}

var configmapsResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}

func updateConfigMapAction(sbu *coreV1.ConfigMap) k8sTesting.UpdateAction {
	return k8sTesting.NewUpdateAction(configmapsResource, sbu.Namespace, sbu)
}

func getConfigMapAction(sbu *coreV1.ConfigMap) k8sTesting.GetAction {
	return k8sTesting.NewGetAction(configmapsResource, sbu.Namespace, sbu.Name)
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

func checkAction(t *testing.T, expected, actual k8sTesting.Action) {
	assert.Truef(t, expected.Matches(actual.GetVerb(), actual.GetResource().Resource),
		"actions not matched: expected [%#v] got [%#v]", expected, actual)

	switch a := actual.(type) {
	case k8sTesting.CreateAction:
		e, ok := expected.(k8sTesting.CreateAction)
		assert.True(t, ok)

		expObject := e.GetObject()
		object := a.GetObject()

		assert.Equal(t, expObject, object)
	case k8sTesting.UpdateAction:
		e, ok := expected.(k8sTesting.UpdateAction)
		assert.True(t, ok)

		expObject := e.GetObject()
		object := a.GetObject()

		assert.Equal(t, expObject, object)
	case k8sTesting.PatchAction:
		e, ok := expected.(k8sTesting.PatchAction)
		assert.True(t, ok)

		expPatch := e.GetPatch()
		patch := a.GetPatch()

		assert.Equal(t, expPatch, patch)
	}
}

func failingReactor(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
	return true, nil, errors.New("custom error")
}

func newLogSinkForErrors() *spy.LogSink {
	logSink := spy.NewLogSink()
	logSink.Logger.Logger.Level = logrus.ErrorLevel
	return logSink
}

func awaitForChanAtMost(t *testing.T, ch <-chan struct{}, timeout time.Duration) {
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatalf("timeout occured when waiting for channel")
	}
}

func fixConfigMap(data map[string]string) *coreV1.ConfigMap {
	return &coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: "system",
			Name:      "system-cm",
		},
		Data: data,
	}
}

func assertErrorContainsStatement(t *testing.T, err error, contains string) {
	assert.Contains(t, err.Error(), contains)
}

func mustMarshal(v interface{}) string {
	marshaled, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("while marshaling, got err: %v", err))
	}

	return string(marshaled)
}
