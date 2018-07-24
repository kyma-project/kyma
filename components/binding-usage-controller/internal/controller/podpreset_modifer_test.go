package controller_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8sSettings "k8s.io/api/settings/v1alpha1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"
)

func TestPodPresetModifierUpsertPodPreset(t *testing.T) {
	tests := map[string]struct {
		objAlreadyInK8s []runtime.Object
		ppToCreate      *k8sSettings.PodPreset
		expActions      []k8sTesting.Action
	}{
		"create existing PodPreset": {
			objAlreadyInK8s: []runtime.Object{
				fixPodPreset(),
			},
			ppToCreate: fixPodPreset(),

			expActions: []k8sTesting.Action{
				createPodPresetAction(fixPodPreset()),
				deletePodPresetAction(fixPodPreset()),
				createPodPresetAction(fixPodPreset()),
			},
		},
		"create not existing PodPreset": {
			objAlreadyInK8s: []runtime.Object{},
			ppToCreate:      fixPodPreset(),

			expActions: []k8sTesting.Action{
				createPodPresetAction(fixPodPreset()),
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			fakeCli := fake.NewSimpleClientset(tc.objAlreadyInK8s...)
			ppModifier := controller.NewPodPresetModifier(fakeCli.SettingsV1alpha1())

			// when
			err := ppModifier.UpsertPodPreset(tc.ppToCreate)

			// then
			assert.NoError(t, err)
			performedActions := fakeCli.Actions()
			require.Equal(t, len(performedActions), len(tc.expActions))
			for idx, expAction := range tc.expActions {
				checkAction(t, performedActions[idx], expAction)
			}

		})
	}
}

func TestPodPresetModifierEnsurePodPresetDeleted(t *testing.T) {
	tests := map[string]struct {
		objAlreadyInK8s []runtime.Object
		ppToDelete      *k8sSettings.PodPreset
	}{
		"delete existing PodPreset": {
			objAlreadyInK8s: []runtime.Object{
				fixPodPreset(),
			},
			ppToDelete: fixPodPreset(),
		},
		"delete not existing PodPreset": {
			objAlreadyInK8s: []runtime.Object{},
			ppToDelete:      fixPodPreset(),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			fakeCli := fake.NewSimpleClientset(tc.objAlreadyInK8s...)
			ppModifier := controller.NewPodPresetModifier(fakeCli.SettingsV1alpha1())

			// when
			err := ppModifier.EnsurePodPresetDeleted(tc.ppToDelete.Namespace, tc.ppToDelete.Name)

			// then
			assert.NoError(t, err)
			performedActions := fakeCli.Actions()
			require.Len(t, performedActions, 1)
			checkAction(t, deletePodPresetAction(tc.ppToDelete), performedActions[0])
		})
	}
}

func fixPodPreset() *k8sSettings.PodPreset {
	return &k8sSettings.PodPreset{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "pp-test",
			Namespace: "ns-test",
		},
		Spec: k8sSettings.PodPresetSpec{
			Selector: metaV1.LabelSelector{
				MatchLabels: map[string]string{
					"test-key": "test-value",
				},
			},
		},
	}
}
