package bebEventing
import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

func (r *Resolver) EventSubscriptionQuery(ctx context.Context, namespace string, name string) (*gqlschema.EventSubscription, error) {
 var result *EventSubscription
 err := r.Service().GetInNamespace(name, namespace, &result)

 return &gqlschema.EventSubscription{Name:result.Name, Namespace:result.Namespace}, err
}
