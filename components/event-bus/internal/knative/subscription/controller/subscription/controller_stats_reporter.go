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
	LabelNamespaceName = "exported_namespace"
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
	kymaSubscriptionCountM = stats.Int64(
		"total_kyma_subscriptions",
		"Number of kyma subscriptions",
		stats.UnitDimensionless,
	)
	knativeSubscriptionCountM = stats.Int64(
		"total_knative_subscriptions",
		"Number of knative subscriptions",
		stats.UnitDimensionless,
	)
	knativeChannelsCountM = stats.Int64(
		"total_knative_channels",
		"Number of knative channels",
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

	// ReportKnativeChannelsGauge captures the knative channel count.
	ReportKnativeChannelsGauge(args *KnativeChannelReportArgs) error
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

func (r *reporter) generateKnativeChannelsTag(args *KnativeChannelReportArgs) (context.Context, error) {
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
		metrics.Record(ctx, kymaSubscriptionCountM.M(1))
	case false:
		metrics.Record(ctx, kymaSubscriptionCountM.M(0))
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
		metrics.Record(ctx, knativeSubscriptionCountM.M(1))
	case false:
		metrics.Record(ctx, knativeSubscriptionCountM.M(0))
	}
	return nil
}

func (r *reporter) ReportKnativeChannelsGauge(args *KnativeChannelReportArgs) error {
	ctx, err := r.generateKnativeChannelsTag(args)
	if err != nil {
		return err
	}
	switch args.Ready {
	case true:
		metrics.Record(ctx, knativeChannelsCountM.M(1))
	case false:
		metrics.Record(ctx, knativeChannelsCountM.M(0))
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
	knativeChannelsTagKeys := []tag.Key{
		nameKey,
	}

	// Create view to see our measurements.
	if err := view.Register(
		&view.View{
			Description: kymaSubscriptionCountM.Description(),
			Measure:     kymaSubscriptionCountM,
			Aggregation: view.LastValue(),
			TagKeys:     kymaSubscriptionTagKeys,
		},
		&view.View{
			Description: knativeSubscriptionCountM.Description(),
			Measure:     knativeSubscriptionCountM,
			Aggregation: view.LastValue(),
			TagKeys:     knativeSubscriptionTagKeys,
		},
		&view.View{
			Description: knativeChannelsCountM.Description(),
			Measure:     knativeChannelsCountM,
			Aggregation: view.LastValue(),
			TagKeys:     knativeChannelsTagKeys,
		},
	); err != nil {
		panic(err)
	}
}

// NewKymaSubscriptionsReportArgs constructs a kymaSubscriptionReportArgs
func NewKymaSubscriptionsReportArgs(namespace string, name string, ready bool) *KymaSubscriptionReportArgs {
	return &KymaSubscriptionReportArgs{
		Namespace: namespace,
		Name:      name,
		Ready:     ready,
	}
}

// NewKnativeSubscriptionsReportArgs constructs a NewKnativeSubscriptionsReportArgs
func NewKnativeSubscriptionsReportArgs(namespace string, name string, ready bool) *KnativeSubscriptionReportArgs {
	return &KnativeSubscriptionReportArgs{
		Namespace: namespace,
		Name:      name,
		Ready:     ready,
	}
}

// NewKnativeChannelsReportArgs constructs a NewKnativeChannelsReportArgs
func NewKnativeChannelsReportArgs(name string, ready bool) *KnativeChannelReportArgs {
	return &KnativeChannelReportArgs{
		Name:  name,
		Ready: ready,
	}
}
