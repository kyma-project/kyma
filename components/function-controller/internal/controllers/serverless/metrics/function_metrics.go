package metrics

import (
	"fmt"
	"time"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	NameLabelKey      = "name"
	NamespaceLabelKey = "namespace"
	TypeLabelKey      = "type"
	RuntimeLabelKey   = "runtime"
)

type PrometheusStatsCollector struct {
	conditionGaugeSet                map[string]bool
	conditionGauges                  map[serverlessv1alpha1.ConditionType]*prometheus.GaugeVec
	FunctionConfiguredStatusGaugeVec *prometheus.GaugeVec
	FunctionBuiltStatusGaugeVec      *prometheus.GaugeVec
	FunctionRunningStatusGaugeVec    *prometheus.GaugeVec
}

func NewPrometheusStatsCollector() *PrometheusStatsCollector {
	functionConfiguredStatusGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "function_configured_status_duration_millisecond",
		Help: "time passed per function from creation until Configured state",
	}, []string{
		NameLabelKey,
		NamespaceLabelKey,
		TypeLabelKey,
		RuntimeLabelKey,
	})
	functionBuiltStatusGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "function_built_status_duration_millisecond",
		Help: "time passed per function from creation until Built state",
	}, []string{
		NameLabelKey,
		NamespaceLabelKey,
		TypeLabelKey,
		RuntimeLabelKey,
	})

	functionRunningStatusGaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "function_running_status_duration_millisecond",
		Help: "time passed per function from creation until Running state",
	}, []string{
		NameLabelKey,
		NamespaceLabelKey,
		TypeLabelKey,
		RuntimeLabelKey,
	})

	instance := &PrometheusStatsCollector{
		conditionGaugeSet: map[string]bool{},
		conditionGauges: map[serverlessv1alpha1.ConditionType]*prometheus.GaugeVec{
			serverlessv1alpha1.ConditionRunning:            functionRunningStatusGaugeVec,
			serverlessv1alpha1.ConditionBuildReady:         functionBuiltStatusGaugeVec,
			serverlessv1alpha1.ConditionConfigurationReady: functionConfiguredStatusGaugeVec,
		},
		FunctionConfiguredStatusGaugeVec: functionConfiguredStatusGaugeVec,
		FunctionBuiltStatusGaugeVec:      functionBuiltStatusGaugeVec,
		FunctionRunningStatusGaugeVec:    functionRunningStatusGaugeVec,
	}

	return instance
}

func (p *PrometheusStatsCollector) Register() {
	metrics.Registry.MustRegister(
		p.FunctionConfiguredStatusGaugeVec,
		p.FunctionBuiltStatusGaugeVec,
		p.FunctionRunningStatusGaugeVec)
}

func (p *PrometheusStatsCollector) UpdateReconcileStats(f *serverlessv1alpha1.Function, cond serverlessv1alpha1.Condition) {
	if _, ok := p.conditionGauges[cond.Type]; !ok { // we don't have a gauge for this condition type
		return
	}
	// a trick to avoid overriding the initial gauge values, since those are
	// what we are interested in. If the function is updated, it initial
	// metrics are _probably_ already collected.
	if f.Generation > 1 {
		return
	}
	gaugeID := p.createGaugeID(f.UID, cond.Type)
	if p.conditionGaugeSet[gaugeID] { // the gauge for this function condition type is already set
		return
	}

	// If the condition status is not true, yet, we will try later.
	// Except for the ConfigReady condition, we always push the metric for it.
	if cond.Status != corev1.ConditionTrue && cond.Type != serverlessv1alpha1.ConditionConfigurationReady {
		return
	}

	labels := p.createLabels(f)
	p.conditionGauges[cond.Type].With(labels).Set(float64(time.Since(f.CreationTimestamp.Time).Milliseconds()))
	p.conditionGaugeSet[gaugeID] = true
}

func (p *PrometheusStatsCollector) createGaugeID(uid types.UID, conditionType serverlessv1alpha1.ConditionType) string {
	return fmt.Sprintf("%s-%s", uid, conditionType)
}

func (p *PrometheusStatsCollector) createLabels(f *serverlessv1alpha1.Function) prometheus.Labels {
	return prometheus.Labels{
		NameLabelKey:      f.Name,
		NamespaceLabelKey: f.Namespace,
		RuntimeLabelKey:   string(f.Spec.Runtime),
		TypeLabelKey:      string(f.Spec.Type),
	}
}
