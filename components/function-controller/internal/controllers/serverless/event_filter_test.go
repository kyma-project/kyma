package serverless

import (
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestIsFunctionStatusUpdate(t *testing.T) {
	//GIVEN

	nopLogger := zap.NewNop().Sugar()

	testCases := map[string]struct {
		event  event.UpdateEvent
		result bool
	}{
		"Function status is not updated": {
			result: true,
			event: event.UpdateEvent{
				ObjectOld: &serverlessv1alpha2.Function{},
				ObjectNew: &serverlessv1alpha2.Function{},
			},
		},
		"Function status is updated": {
			result: false,
			event: event.UpdateEvent{
				ObjectOld: &serverlessv1alpha2.Function{},
				ObjectNew: &serverlessv1alpha2.Function{
					Status: serverlessv1alpha2.FunctionStatus{
						Conditions: []serverlessv1alpha2.Condition{
							{
								Type:               serverlessv1alpha2.ConditionBuildReady,
								Status:             corev1.ConditionUnknown,
								LastTransitionTime: metav1.Time{},
								Reason:             "test reason",
								Message:            "test message",
							},
						},
					},
				},
			},
		},
		"Not function update event": {
			result: true,
			event: event.UpdateEvent{
				ObjectOld: &corev1.Pod{},
				ObjectNew: &corev1.Pod{},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			fn := IsNotFunctionStatusUpdate(nopLogger)
			//WHEN
			actual := fn(testCase.event)
			//THEN
			assert.Equal(t, testCase.result, actual)
		})
	}
}
