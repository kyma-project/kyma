package metrics

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/require"

	//"github.com/prometheus/client_golang/testutil"
	"github.com/prometheus/client_golang/prometheus/testutil"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_UpdateFunctionStatusGauge(t *testing.T) {
	t.Run("Check if gauge is not updated second generation", func(t *testing.T) {
		//GIVEN
		stas := NewPrometheusStatsCollector()
		startTimestamp := v1.NewTime(time.Date(2000, 01, 01, 00, 00, 0, 0, time.Local))
		f := &serverlessv1alpha1.Function{
			ObjectMeta: v1.ObjectMeta{
				Name:              "test-fn",
				Namespace:         "test-namespace",
				Generation:        1,
				CreationTimestamp: startTimestamp,
			},
			Spec: serverlessv1alpha1.FunctionSpec{
				Runtime: serverlessv1alpha1.Python39,
				Type:    serverlessv1alpha1.SourceTypeGit,
			},
		}
		cond := serverlessv1alpha1.Condition{
			Type:   serverlessv1alpha1.ConditionRunning,
			Status: corev1.ConditionTrue,
		}

		//WHEN
		stas.UpdateReconcileStats(f, cond)

		//THEN
		value := testutil.ToFloat64(stas.FunctionRunningStatusGaugeVec.With(stas.createLabels(f)))
		require.InDelta(t, time.Since(startTimestamp.Time).Milliseconds(), value, 0.1, "The gauge value wasn't set")
		require.True(t, stas.conditionGaugeSet[stas.createGaugeID(f.UID, cond.Type)])

		// The next update of stats within the same condition shouldn't update gauge value,
		// that's why creationTimestamp is increased to proof, that gauge value didn't change.
		//GIVEN
		f.ObjectMeta.Generation = f.ObjectMeta.Generation + 1
		f.ObjectMeta.CreationTimestamp = v1.NewTime(startTimestamp.Add(10 * time.Second))

		//WHEN
		stas.UpdateReconcileStats(f, cond)

		//THEN
		value = testutil.ToFloat64(stas.FunctionRunningStatusGaugeVec.With(stas.createLabels(f)))
		require.InDelta(t, time.Since(startTimestamp.Time).Milliseconds(), value, 0.1, "The gauge value was updated but shouldn't")
	})

	t.Run("Always collect metrics for Configuration ready condition", func(t *testing.T) {
		//GIVEN
		stas := NewPrometheusStatsCollector()
		startTimestamp := v1.NewTime(time.Date(2000, 01, 01, 00, 00, 0, 0, time.Local))
		f := &serverlessv1alpha1.Function{
			ObjectMeta: v1.ObjectMeta{
				Name:              "test-fn",
				Namespace:         "test-namespace",
				Generation:        1,
				CreationTimestamp: startTimestamp,
			},
			Spec: serverlessv1alpha1.FunctionSpec{
				Runtime: serverlessv1alpha1.Python39,
				Type:    serverlessv1alpha1.SourceTypeGit,
			},
		}
		cond := serverlessv1alpha1.Condition{
			Type:   serverlessv1alpha1.ConditionConfigurationReady,
			Status: corev1.ConditionFalse,
		}

		//WHEN
		stas.UpdateReconcileStats(f, cond)

		//THEN
		value := testutil.ToFloat64(stas.FunctionConfiguredStatusGaugeVec.With(stas.createLabels(f)))
		require.InDelta(t, time.Since(startTimestamp.Time).Milliseconds(), value, 0.1, "The gauge value wasn't set")
		require.True(t, stas.conditionGaugeSet[stas.createGaugeID(f.UID, cond.Type)])
	})
	t.Run("Don't collect metrics for condition status false", func(t *testing.T) {
		//GIVEN
		stas := NewPrometheusStatsCollector()
		startTimestamp := v1.NewTime(time.Date(2000, 01, 01, 00, 00, 0, 0, time.Local))
		f := &serverlessv1alpha1.Function{
			ObjectMeta: v1.ObjectMeta{
				Name:              "test-fn",
				Namespace:         "test-namespace",
				Generation:        1,
				CreationTimestamp: startTimestamp,
			},
			Spec: serverlessv1alpha1.FunctionSpec{
				Runtime: serverlessv1alpha1.Python39,
				Type:    serverlessv1alpha1.SourceTypeGit,
			},
		}
		cond := serverlessv1alpha1.Condition{
			Type:   serverlessv1alpha1.ConditionRunning,
			Status: corev1.ConditionFalse,
		}

		//WHEN
		stas.UpdateReconcileStats(f, cond)

		//THEN
		value := testutil.ToFloat64(stas.FunctionRunningStatusGaugeVec.With(stas.createLabels(f)))
		require.InDelta(t, 0, value, 0.1, "The gauge value was updated but shouldn't")
		require.False(t, stas.conditionGaugeSet[stas.createGaugeID(f.UID, cond.Type)])
	})
}
