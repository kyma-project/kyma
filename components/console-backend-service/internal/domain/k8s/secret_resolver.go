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
	Update(name, namespace string, update v1.Secret) (*v1.Secret, error)
	Delete(name, namespace string) error
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
	ToGQL(in *v1.Secret) (*gqlschema.Secret, error)
	ToGQLs(in []*v1.Secret) ([]gqlschema.Secret, error)
	GQLJSONToSecret(in gqlschema.JSON) (v1.Secret, error)
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
	return r.converter.ToGQL(secret)
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

	return r.converter.ToGQLs(secrets)
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

func (r *secretResolver) UpdateSecretMutation(ctx context.Context, name string, namespace string, update gqlschema.JSON) (*gqlschema.Secret, error) {
	secret, err := r.converter.GQLJSONToSecret(update)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s` from namespace `%s`", pretty.Secret, name, namespace))
		return nil, gqlerror.New(err, pretty.Secret, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	updated, err := r.secretSvc.Update(name, namespace, secret)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s` from namespace %s", pretty.Secret, name, namespace))
		return nil, gqlerror.New(err, pretty.Secret, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return r.converter.ToGQL(updated)
}

func (r *secretResolver) DeleteSecretMutation(ctx context.Context, name string, namespace string) (*gqlschema.Secret, error) {
	secret, err := r.secretSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding %s `%s` in namespace `%s`", pretty.Secret, name, namespace))
		return nil, gqlerror.New(err, pretty.Secret, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	secretCopy := secret.DeepCopy()
	deletedSecret, err := r.converter.ToGQL(secretCopy)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting %s `%s` from namespace `%s`", pretty.Secret, name, namespace))
		return nil, gqlerror.New(err, pretty.Secret, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	err = r.secretSvc.Delete(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s` from namespace `%s`", pretty.Secret, name, namespace))
		return nil, gqlerror.New(err, pretty.Secret, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return deletedSecret, nil
}
