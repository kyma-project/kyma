package v1alpha1_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/onsi/gomega/types"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

const (
	generation = 123
	imageTag   = "123"
)

func Test(t *testing.T) {
	errTest := errors.New("test error")

	testCases := []struct {
		desc    string
		provide func() v1alpha1.FunctionStatus
		match   GomegaMatcher
	}{
		{
			desc: "FunctionStatusDeploySucceeded",
			provide: func() v1alpha1.FunctionStatus {
				return *fn().FunctionStatusDeploySucceeded()
			},
			match: MatchFields(IgnoreExtras, Fields{
				"Phase":              Equal(v1alpha1.FunctionPhaseRunning),
				"ImageTag":           Equal(imageTag),
				"ObservedGeneration": BeNumerically("==", generation),
				"Conditions": ContainElement(MatchFields(IgnoreExtras, Fields{
					"Type":               Equal(v1alpha1.ConditionTypeDeployed),
					"Reason":             Equal(v1alpha1.ConditionReasonDeploySucceeded),
					"Message":            BeZero(),
					"LastTransitionTime": Not(BeZero()),
				})),
			}),
		},
		{
			desc: "FunctionStatusInitializing",
			provide: func() v1alpha1.FunctionStatus {
				return *fn().FunctionStatusInitializing()
			},
			match: MatchFields(IgnoreExtras, Fields{
				"Phase":              Equal(v1alpha1.FunctionPhaseInitializing),
				"ImageTag":           Equal(imageTag),
				"ObservedGeneration": BeNumerically("==", generation),
				"Conditions":         BeEmpty(),
			}),
		},
		{
			desc: "FunctionStatusConfigFailed",
			provide: func() v1alpha1.FunctionStatus {
				return *fn().FunctionStatusUpdateConfigFailed(errTest)
			},
			match: MatchFields(IgnoreExtras, Fields{
				"Phase":              Equal(v1alpha1.FunctionPhaseFailed),
				"ImageTag":           Equal(imageTag),
				"ObservedGeneration": BeNumerically("==", generation),
				"Conditions": ContainElement(MatchFields(IgnoreExtras, Fields{
					"Type":               Equal(v1alpha1.ConditionTypeError),
					"Reason":             Equal(v1alpha1.ConditionReasonUpdateConfigFailed),
					"Message":            Equal(errTest.Error()),
					"LastTransitionTime": Not(BeZero()),
				})),
			}),
		},
		{
			desc: "FunctionStatusUpdateConfigSucceeded",
			provide: func() v1alpha1.FunctionStatus {
				return *fn().FunctionStatusUpdateConfigSucceeded("456")
			},
			match: MatchFields(IgnoreExtras, Fields{
				"Phase":              Equal(v1alpha1.FunctionPhaseBuilding),
				"ImageTag":           Equal("456"),
				"ObservedGeneration": BeNumerically("==", generation),
				"Conditions": ContainElement(MatchFields(IgnoreExtras, Fields{
					"Type":               Equal(v1alpha1.ConditionTypeInitialized),
					"Reason":             Equal(v1alpha1.ConditionReasonUpdateConfigSucceeded),
					"Message":            BeZero(),
					"LastTransitionTime": Not(BeZero()),
				})),
			}),
		},
		{
			desc: "FunctionStatusUpdateRuntimeConfig",
			provide: func() v1alpha1.FunctionStatus {
				return *fn().FunctionStatusUpdateRuntimeConfig()
			},
			match: MatchFields(IgnoreExtras, Fields{
				"Phase":              Equal(v1alpha1.FunctionPhaseBuilding),
				"ImageTag":           Equal(imageTag),
				"ObservedGeneration": BeNumerically("==", generation),
				"Conditions": ContainElement(MatchFields(IgnoreExtras, Fields{
					"Type":               Equal(v1alpha1.ConditionTypeInitialized),
					"Reason":             Equal(v1alpha1.ConditionReasonUpdateRuntimeConfig),
					"Message":            BeZero(),
					"LastTransitionTime": Not(BeZero()),
				})),
			}),
		},
		{
			desc: "FunctionStatusBuildRunning",
			provide: func() v1alpha1.FunctionStatus {
				return *fn().FunctionStatusBuildRunning()
			},
			match: MatchFields(IgnoreExtras, Fields{
				"Phase":              Equal(v1alpha1.FunctionPhaseBuilding),
				"ImageTag":           Equal(imageTag),
				"ObservedGeneration": BeNumerically("==", generation),
				"Conditions":         BeEquivalentTo(fn().Status.Conditions),
			}),
		},
		{
			desc: "FunctionStatusBuildSucceed",
			provide: func() v1alpha1.FunctionStatus {
				return *fn().FunctionStatusBuildSucceed()
			},
			match: MatchFields(IgnoreExtras, Fields{
				"Phase":              Equal(v1alpha1.FunctionPhaseDeploying),
				"ImageTag":           Equal(imageTag),
				"ObservedGeneration": BeNumerically("==", generation),
				"Conditions": ContainElement(MatchFields(IgnoreExtras, Fields{
					"Type":               Equal(v1alpha1.ConditionTypeImageCreated),
					"Reason":             Equal(v1alpha1.ConditionReasonBuildSucceeded),
					"Message":            BeZero(),
					"LastTransitionTime": Not(BeZero()),
				})),
			}),
		},
		{
			desc: "FunctionStatusGetConfigFailed",
			provide: func() v1alpha1.FunctionStatus {
				return *fn().FunctionStatusGetConfigFailed(errTest)
			},
			match: MatchFields(IgnoreExtras, Fields{
				"Phase":              Equal(v1alpha1.FunctionPhaseFailed),
				"ImageTag":           Not(BeEmpty()),
				"ObservedGeneration": BeNumerically("==", generation),
				"Conditions": ContainElement(MatchFields(IgnoreExtras, Fields{
					"Type":               Equal(v1alpha1.ConditionTypeError),
					"Reason":             Equal(v1alpha1.ConditionReasonGetConfigFailed),
					"Message":            Equal(errTest.Error()),
					"LastTransitionTime": Not(BeZero()),
				})),
			}),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			g := NewWithT(t)
			actual := tC.provide()
			g.Expect(actual).Should(tC.match)
		})
	}
}

func fn() *v1alpha1.Function {
	return &v1alpha1.Function{
		ObjectMeta: v1.ObjectMeta{
			Generation: generation,
		},
		Status: v1alpha1.FunctionStatus{
			ImageTag: imageTag,
		},
	}
}
