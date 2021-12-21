package serverless

import (
	"fmt"
	"time"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
)

const (
	NameLabelKey      = "name"
	NamespaceLabelKey = "namespace"
	TypeLabelKey      = "type"
	RuntimeLabelKey   = "runtime"
)

var (
	ReconcileCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "function_controller_counter",
		Help: "number of reconciles",
	})
	FunctionConfiguredStatusGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "function_configured_status_duration_millisecond",
		Help: "time passed per function from creation until Configured state",
	}, []string{
		NameLabelKey,
		NamespaceLabelKey,
		TypeLabelKey,
		RuntimeLabelKey,
	})
	FunctionBuiltStatusGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "function_built_status_duration_millisecond",
		Help: "time passed per function from creation until Built state",
	}, []string{
		NameLabelKey,
		NamespaceLabelKey,
		TypeLabelKey,
		RuntimeLabelKey,
	})

	FunctionRunningStatusGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "function_running_status_duration_millisecond",
		Help: "time passed per function from creation until Running state",
	}, []string{
		NameLabelKey,
		NamespaceLabelKey,
		TypeLabelKey,
		RuntimeLabelKey,
	})

	conditionGauges = map[serverlessv1alpha1.ConditionType]*prometheus.GaugeVec{
		serverlessv1alpha1.ConditionRunning:            FunctionRunningStatusGaugeVec,
		serverlessv1alpha1.ConditionBuildReady:         FunctionBuiltStatusGaugeVec,
		serverlessv1alpha1.ConditionConfigurationReady: FunctionConfiguredStatusGaugeVec,
	}

	ConditionGaugeSet = map[string]bool{}
)

func updateFunctionStatusGauge(f *serverlessv1alpha1.Function, cond serverlessv1alpha1.Condition) {
	if _, ok := conditionGauges[cond.Type]; !ok { // we don't have a gauge for this condition type
		return
	}

	gaugeID := fmt.Sprintf("%s-%s", f.UID, cond.Type)
	if ConditionGaugeSet[gaugeID] { // the gauge for this function condition type is already set
		return
	}
	// if the condition status is not true, yet, we will try later. Except for the ConfigReady condition since we set that directly to true
	if cond.Status != corev1.ConditionTrue && cond.Type != serverlessv1alpha1.ConditionConfigurationReady {
		return
	}

	labels := prometheus.Labels{
		NameLabelKey:      f.Name,
		NamespaceLabelKey: f.Namespace,
		RuntimeLabelKey:   string(f.Spec.Runtime),
		TypeLabelKey:      string(f.Spec.Type),
	}
	conditionGauges[cond.Type].With(labels).Set(float64(time.Since(f.CreationTimestamp.Time).Milliseconds()))
	ConditionGaugeSet[gaugeID] = true
}
