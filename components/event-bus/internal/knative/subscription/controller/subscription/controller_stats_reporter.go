package subscription

import (
	"context"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"

	"knative.dev/pkg/metrics"
)

const (
	// LabelSubscriptionName is the label for immutable name of the namespace that the service is deployed
	LabelSubscriptionName = "name"
	LabelNamespaceName    = "exported_namespace"
	LabelReadyName        = "ready"
)

var (
	namespaceKey = tag.MustNewKey(LabelNamespaceName)
	nameKey      = tag.MustNewKey(LabelSubscriptionName)
	readyKey     = tag.MustNewKey(LabelReadyName)
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
		"total_knative_channel",
		"Number of knative channels",
		stats.UnitDimensionless,
	)

	_ StatsReporter = (*reporter)(nil)
)

type kymaSubscriptionReportArgs struct {
	Namespace string
	Name      string
	Ready     string
}

type knativeSubscriptionReportArgs struct {
	Namespace string
	Name      string
	Ready     string
}

type knativeChannelsReportArgs struct {
	Ready     string
}

// StatsReporter defines the interface for sending Kyma Subscription Controller metrics.
type StatsReporter interface {
	// ReportSubscriptionGauge captures the kyma subscription count.
	ReportKymaSubscriptionGauge(args *kymaSubscriptionReportArgs) error
	// DeleteSubscriptionGauge deletes the kyma subscription count.
	DeleteKymaSubscriptionGauge(args *kymaSubscriptionReportArgs) error
	// ReportSubscriptionGauge captures the knative subscription count.
	ReportKnativeSubscriptionGauge(args *knativeSubscriptionReportArgs) error
	// DeleteSubscriptionGauge deletes the knative subscription count.
	DeleteKnativeSubscriptionGauge(args *knativeSubscriptionReportArgs) error
	// ReportSubscriptionGauge captures the knative channel count.
	ReportKnativeChannelsGauge(args *knativeChannelsReportArgs) error
	// DeleteSubscriptionGauge deletes the knative channel count.
	DeleteKnativeChannelsGauge(args *knativeChannelsReportArgs) error
}

func (r *reporter) generateKymaSubscriptionTag(args *kymaSubscriptionReportArgs) (context.Context, error) {
	return tag.New(
		r.ctx,
		tag.Insert(namespaceKey, args.Namespace),
		tag.Insert(nameKey, args.Name),
		tag.Insert(readyKey, args.Ready))
}

func (r *reporter) generateKnativeSubscriptionTag(args *knativeSubscriptionReportArgs) (context.Context, error) {
	return tag.New(
		r.ctx,
		tag.Insert(namespaceKey, args.Namespace),
		tag.Insert(nameKey, args.Name),
		tag.Insert(readyKey, args.Ready))
}

func (r *reporter) generateKnativeChannelsTag(args *knativeChannelsReportArgs) (context.Context, error) {
	return tag.New(
		r.ctx,
		tag.Insert(readyKey, args.Ready))
}


func (r *reporter) ReportKymaSubscriptionGauge(args *kymaSubscriptionReportArgs) error {
	ctx, err := r.generateKymaSubscriptionTag(args)
	if err != nil {
		return err
	}
	metrics.Record(ctx, kymaSubscriptionCountM.M(1))
	return nil
}

func (r *reporter) DeleteKymaSubscriptionGauge(args *kymaSubscriptionReportArgs) error {
	ctx, err := r.generateKymaSubscriptionTag(args)
	if err != nil {
		return err
	}
	metrics.Record(ctx, kymaSubscriptionCountM.M(0))
	//TODO Delete
	return nil
}

func (r *reporter) ReportKnativeSubscriptionGauge(args *knativeSubscriptionReportArgs) error {
	ctx, err := r.generateKnativeSubscriptionTag(args)
	if err != nil {
		return err
	}
	metrics.Record(ctx, knativeSubscriptionCountM.M(1))
	return nil
}

func (r *reporter) DeleteKnativeSubscriptionGauge(args *knativeSubscriptionReportArgs) error {
	ctx, err := r.generateKnativeSubscriptionTag(args)
	if err != nil {
		return err
	}
	metrics.Record(ctx, knativeSubscriptionCountM.M(0))
	//TODO Delete
	return nil
}

func (r *reporter) ReportKnativeChannelsGauge(args *knativeChannelsReportArgs) error {
	ctx, err := r.generateKnativeChannelsTag(args)
	if err != nil {
		return err
	}
	metrics.Record(ctx, knativeChannelsCountM.M(1))
	return nil
}

func (r *reporter) DeleteKnativeChannelsGauge(args *knativeChannelsReportArgs) error {
	ctx, err := r.generateKnativeChannelsTag(args)
	if err != nil {
		return err
	}
	metrics.Record(ctx, knativeChannelsCountM.M(0))
	//TODO Delete
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
		readyKey,
	}
	knativeSubscriptionTagKeys := []tag.Key{
		namespaceKey,
		nameKey,
		readyKey,
	}
	knativeChannelsTagKeys := []tag.Key{
		readyKey,
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

func NewKymaSubscriptionsReportArgs(namespace string, name string, ready string) *kymaSubscriptionReportArgs {
	return &kymaSubscriptionReportArgs{
		Namespace: namespace,
		Name:      name,
		Ready:     ready,
	}
}

func NewKnativeSubscriptionsReportArgs(namespace string, name string, ready string) *kymaSubscriptionReportArgs {
	return &kymaSubscriptionReportArgs{
		Namespace: namespace,
		Name:      name,
		Ready:     ready,
	}
}

func NewKnativeChannelsReportArgs(namespace string, name string, ready string) *kymaSubscriptionReportArgs {
	return &kymaSubscriptionReportArgs{
		Namespace: namespace,
		Name:      name,
		Ready:     ready,
	}
}
