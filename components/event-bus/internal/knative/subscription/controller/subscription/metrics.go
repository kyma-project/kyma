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
	subscriptionCountM = stats.Int64(
		"total_kyma_subscriptions",
		"Number of kyma subscriptions",
		stats.UnitDimensionless,
	)
	namespaceKey = tag.MustNewKey(LabelNamespaceName)
	nameKey      = tag.MustNewKey(LabelSubscriptionName)
	readyKey     = tag.MustNewKey(LabelReadyName)

	_ SubscriptionsStatsReporter = (*reporter)(nil)
)

// reporter holds cached metric objects to report Kyma Subscription Controller metrics.
type reporter struct {
	ctx context.Context
}

type ReportArgs struct {
	Namespace string
	Name      string
	Ready     string
}

// SubscriptionsStatsReporter defines the interface for sending Kyma Subscription Controller metrics.
type SubscriptionsStatsReporter interface {
	// ReportSubscriptionCount captures the subscription count. It records one per call.
	ReportSubscriptionCount(args *ReportArgs) error
}

func (r *reporter) generateTag(args *ReportArgs) (context.Context, error) {
	return tag.New(
		r.ctx,
		tag.Insert(namespaceKey, args.Namespace),
		tag.Insert(nameKey, args.Name),
		tag.Insert(readyKey, args.Ready))
}

func (r *reporter) ReportSubscriptionCount(args *ReportArgs) error {
	ctx, err := r.generateTag(args)
	if err != nil {
		return err
	}
	metrics.Record(ctx, subscriptionCountM.M(1))
	return nil
}

// NewStatsReporter creates a reporter that collects and reports source metrics.
func NewStatsReporter() (SubscriptionsStatsReporter, error) {
	ctx, err := tag.New(
		context.Background(),
	)
	if err != nil {
		return nil, err
	}
	return &reporter{ctx: ctx}, nil
}

func registerMetrics() {
	tagKeys := []tag.Key{
		namespaceKey,
		nameKey,
		readyKey,
	}

	// Create view to see our measurements.
	if err := view.Register(
		&view.View{
			Description: subscriptionCountM.Description(),
			Measure:     subscriptionCountM,
			Aggregation: view.Count(),
			TagKeys:     tagKeys,
		},
	); err != nil {
		panic(err)
	}
}
