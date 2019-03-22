package k8s

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

//go:generate mockery -name=secretSvc -output=automock -outpkg=automock -case=underscore
type secretSvc interface {
	Find(name, namespace string) (*v1.Secret, error)
	List(namespace string, params pager.PagingParams) ([]*v1.Secret, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

func newSecretResolver(svc secretSvc) *secretResolver {
	return &secretResolver{
		converter: &secretConverter{},
		secretSvc: svc,
	}
}

//go:generate mockery -name=gqlSecretConverter -output=automock -outpkg=automock -case=underscore
type gqlSecretConverter interface {
	ToGQL(in *v1.Secret) *gqlschema.Secret
	ToGQLs(in []*v1.Secret) []gqlschema.Secret
}

type secretResolver struct {
	converter gqlSecretConverter
	secretSvc secretSvc
}

func (r *secretResolver) SecretQuery(ctx context.Context, name, ns string) (*gqlschema.Secret, error) {
	secret, err := r.secretSvc.Find(name, ns)
	switch {
	case k8serrors.IsNotFound(err):
		return nil, nil
	case err != nil:
		glog.Error(
			errors.Wrapf(err, "while getting %s [name: %s, namespace: %s]", pretty.Secret, name, ns))
		return nil, gqlerror.New(err, pretty.Secret, gqlerror.WithName(name), gqlerror.WithNamespace(ns))
	}
	return r.converter.ToGQL(secret), nil
}

func (r *secretResolver) SecretsQuery(ctx context.Context, ns string, first *int, offset *int) ([]gqlschema.Secret, error) {
	secrets, err := r.secretSvc.List(ns, pager.PagingParams{
		First:  first,
		Offset: offset,
	})

	switch {
	case k8serrors.IsNotFound(err):
		return nil, nil
	case err != nil:
		glog.Error(
			errors.Wrapf(err, "while getting secrets [namespace: %s]", ns))
		return nil, gqlerror.New(err, pretty.Secret, gqlerror.WithNamespace(ns))
	}

	return r.converter.ToGQLs(secrets), nil
}

func (r *secretResolver) SecretEventSubscription(ctx context.Context, namespace string) (<-chan gqlschema.SecretEvent, error) {
	channel := make(chan gqlschema.SecretEvent, 1)
	filter := func(secret *v1.Secret) bool {
		return secret != nil && secret.Namespace == namespace
	}

	secretListener := listener.NewSecret(channel, filter, r.converter)
	r.secretSvc.Subscribe(secretListener)
	go func() {
		defer close(channel)
		defer r.secretSvc.Unsubscribe(secretListener)
		<-ctx.Done()
	}()

	return channel, nil
}
