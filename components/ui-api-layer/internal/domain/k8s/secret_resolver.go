package k8s

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

func newSecretResolver(secretGetter v1.SecretsGetter) *secretResolver {
	return &secretResolver{
		converter:    secretConverter{},
		secretGetter: secretGetter,
	}
}

type secretResolver struct {
	secretGetter v1.SecretsGetter
	converter    secretConverter
}

func (r *secretResolver) SecretQuery(ctx context.Context, name, env string) (*gqlschema.Secret, error) {
	secret, err := r.secretGetter.Secrets(env).Get(name, metav1.GetOptions{})
	switch {
	case apierrors.IsNotFound(err):
		return nil, nil
	case err != nil:
		glog.Error(
			errors.Wrapf(err, "while getting %s [name: %s, environment: %s]", pretty.Secret, name, env))
		return nil, gqlerror.New(err, pretty.Secret, gqlerror.WithName(name), gqlerror.WithEnvironment(env))
	}

	return r.converter.ToGQL(secret), nil

}
