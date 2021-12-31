package metrics

import (
	"testing"
	"time"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_UpdateFunctionStatusGauge(t *testing.T) {
	//GIVEN
	stas := NewPrometheusStatsCollector()
	timestamp := v1.NewTime(time.Unix(10, 0))
	f := &serverlessv1alpha1.Function{
		ObjectMeta: v1.ObjectMeta{
			Name:              "test-fn",
			Namespace:         "test-namespace",
			Generation:        0,
			CreationTimestamp: timestamp,
		},
		Spec: serverlessv1alpha1.FunctionSpec{
			Runtime: serverlessv1alpha1.Python39,
			Type:    serverlessv1alpha1.SourceTypeGit,
		},
	}
	cond := serverlessv1alpha1.Condition{
		Type: serverlessv1alpha1.ConditionConfigurationReady,
	}

	//WHEN
	stas.UpdateFunctionStatusGauge(f, cond)

	//THEN
	require.True(t, stas.conditionGaugeSet[stas.createGaugeID(f.UID, cond.Type)])

	//TODO: how to check if gauge was set with correct value
	//
	f.ObjectMeta.Generation = 1

}
