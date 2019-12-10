package subscription

import (
	"context"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"knative.dev/pkg/metrics"
)

const (
	// LabelSubscriptionName is the label for immutable name of the name
	LabelSubscriptionName = "name"
	// LabelNamespaceName is the label for immutable name of the namespace
	LabelNamespaceName = "subscription_ns"
)

var (
	namespaceKey = tag.MustNewKey(LabelNamespaceName)
	nameKey      = tag.MustNewKey(LabelSubscriptionName)
)

// reporter holds cached metric objects to report Kyma Subscription Controller metrics.
type reporter struct {
	ctx context.Context
}

var (
	kymaSubscriptionReadyM = stats.Int64(
		"kyma_subscription_status",
		"kyma subscription status",
		stats.UnitDimensionless,
	)
	knativeSubscriptionReadyM = stats.Int64(
		"knative_subscription_status",
		"knative subscription status",
		stats.UnitDimensionless,
	)
	knativeChannelReadyM = stats.Int64(
		"knative_channel_status",
		"knative channel status",
		stats.UnitDimensionless,
	)

	_ StatsReporter = (*reporter)(nil)
)

//KymaSubscriptionReportArgs Kyma Subscription Report Args
type KymaSubscriptionReportArgs struct {
	Namespace string
	Name      string
	Ready     bool
}

//KnativeSubscriptionReportArgs Knative Subscription Report Args
type KnativeSubscriptionReportArgs struct {
	Namespace string
	Name      string
	Ready     bool
}

//KnativeChannelReportArgs Knative Channel Report Args
type KnativeChannelReportArgs struct {
	Name  string
	Ready bool
}

// StatsReporter defines the interface for sending Kyma Subscription Controller metrics.
type StatsReporter interface {
	// ReportKymaSubscriptionGauge captures the kyma subscription count.
	ReportKymaSubscriptionGauge(args *KymaSubscriptionReportArgs) error

	// ReportKnativeSubscriptionGauge captures the knative subscription count.
	ReportKnativeSubscriptionGauge(args *KnativeSubscriptionReportArgs) error

	// ReportKnativeChannelGauge captures the knative channel count.
	ReportKnativeChannelGauge(args *KnativeChannelReportArgs) error
}

func (r *reporter) generateKymaSubscriptionTag(args *KymaSubscriptionReportArgs) (context.Context, error) {
	return tag.New(
		r.ctx,
		tag.Insert(namespaceKey, args.Namespace),
		tag.Insert(nameKey, args.Name))
}

func (r *reporter) generateKnativeSubscriptionTag(args *KnativeSubscriptionReportArgs) (context.Context, error) {
	return tag.New(
		r.ctx,
		tag.Insert(namespaceKey, args.Namespace),
		tag.Insert(nameKey, args.Name))
}

func (r *reporter) generateKnativeChannelTag(args *KnativeChannelReportArgs) (context.Context, error) {
	return tag.New(
		r.ctx,
		tag.Insert(nameKey, args.Name))
}

func (r *reporter) ReportKymaSubscriptionGauge(args *KymaSubscriptionReportArgs) error {
	ctx, err := r.generateKymaSubscriptionTag(args)
	if err != nil {
		return err
	}
	switch args.Ready {
	case true:
		metrics.Record(ctx, kymaSubscriptionReadyM.M(1))
	case false:
		metrics.Record(ctx, kymaSubscriptionReadyM.M(0))
	}
	return nil
}

func (r *reporter) ReportKnativeSubscriptionGauge(args *KnativeSubscriptionReportArgs) error {
	ctx, err := r.generateKnativeSubscriptionTag(args)
	if err != nil {
		return err
	}
	switch args.Ready {
	case true:
		metrics.Record(ctx, knativeSubscriptionReadyM.M(1))
	case false:
		metrics.Record(ctx, knativeSubscriptionReadyM.M(0))
	}
	return nil
}

func (r *reporter) ReportKnativeChannelGauge(args *KnativeChannelReportArgs) error {
	ctx, err := r.generateKnativeChannelTag(args)
	if err != nil {
		return err
	}
	switch args.Ready {
	case true:
		metrics.Record(ctx, knativeChannelReadyM.M(1))
	case false:
		metrics.Record(ctx, knativeChannelReadyM.M(0))
	}
	return nil
}

// NewStatsReporter creates a reporter that collects and reports source metrics.
func NewStatsReporter() (StatsReporter, error) {
	ctx, err := tag.New(
		context.Background(),
	)
	if err != nil {
		return nil, err
	}
	return &reporter{ctx: ctx}, nil
}

func registerMetrics() {
	kymaSubscriptionTagKeys := []tag.Key{
		namespaceKey,
		nameKey,
	}
	knativeSubscriptionTagKeys := []tag.Key{
		namespaceKey,
		nameKey,
	}
	knativeChannelTagKeys := []tag.Key{
		nameKey,
	}

	// Create view to see our measurements.
	if err := view.Register(
		&view.View{
			Description: kymaSubscriptionReadyM.Description(),
			Measure:     kymaSubscriptionReadyM,
			Aggregation: view.LastValue(),
			TagKeys:     kymaSubscriptionTagKeys,
		},
		&view.View{
			Description: knativeSubscriptionReadyM.Description(),
			Measure:     knativeSubscriptionReadyM,
			Aggregation: view.LastValue(),
			TagKeys:     knativeSubscriptionTagKeys,
		},
		&view.View{
			Description: knativeChannelReadyM.Description(),
			Measure:     knativeChannelReadyM,
			Aggregation: view.LastValue(),
			TagKeys:     knativeChannelTagKeys,
		},
	); err != nil {
		panic(err)
	}
}

// NewKymaSubscriptionReportArgs constructs a kymaSubscriptionReportArgs
func NewKymaSubscriptionReportArgs(namespace string, name string, ready bool) *KymaSubscriptionReportArgs {
	return &KymaSubscriptionReportArgs{
		Namespace: namespace,
		Name:      name,
		Ready:     ready,
	}
}

// NewKnativeSubscriptionReportArgs constructs a NewKnativeSubscriptionReportArgs
func NewKnativeSubscriptionReportArgs(namespace string, name string, ready bool) *KnativeSubscriptionReportArgs {
	return &KnativeSubscriptionReportArgs{
		Namespace: namespace,
		Name:      name,
		Ready:     ready,
	}
}

// NewKnativeChannelReportArgs constructs a NewKnativeChannelReportArgs
func NewKnativeChannelReportArgs(name string, ready bool) *KnativeChannelReportArgs {
	return &KnativeChannelReportArgs{
		Name:  name,
		Ready: ready,
	}
}
