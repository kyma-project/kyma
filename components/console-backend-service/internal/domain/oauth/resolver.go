package oauth

import (
	"context"
	"errors"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/ory/hydra-maester/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClientList []*v1alpha1.OAuth2Client

func (l *ClientList) Append() interface{} {
	e := &v1alpha1.OAuth2Client{}
	*l = append(*l, e)
	return e
}

func (r *Resolver) OAuth2ClientsQuery(ctx context.Context, namespace string) ([]*v1alpha1.OAuth2Client, error) {
	items := ClientList{}
	var err error

	err = r.Service().ListInNamespace(namespace, &items)
	return items, err
}

func (r *Resolver) OAuth2ClientQuery(ctx context.Context, name, namespace string) (*v1alpha1.OAuth2Client, error) {
	var result *v1alpha1.OAuth2Client
	err := r.Service().GetInNamespace(name, namespace, &result)
	return result, err
}

func (r *Resolver) CreateOAuth2Client(ctx context.Context, name string, namespace string, params v1alpha1.OAuth2ClientSpec) (*v1alpha1.OAuth2Client, error) {
	client := &v1alpha1.OAuth2Client{
		TypeMeta: metav1.TypeMeta{
			APIVersion: oAuth2ClientGroupVersionResource.GroupVersion().String(),
			Kind:       oAuth2ClientKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: params,
	}
	result := &v1alpha1.OAuth2Client{}
	err := r.Service().Create(client, result)
	return result, err
}

func (r *Resolver) UpdateOAuth2Client(ctx context.Context, name string, namespace string, generation int64, params v1alpha1.OAuth2ClientSpec) (*v1alpha1.OAuth2Client, error) {
	result := &v1alpha1.OAuth2Client{}
	err := r.Service().UpdateInNamespace(name, namespace, result, func() error {
		if result.Generation > generation {
			return errors.New("resource already modified")
		}

		result.Spec = params
		return nil
	})
	return result, err
}

func (r *Resolver) DeleteOAuth2Client(ctx context.Context, name string, namespace string) (*v1alpha1.OAuth2Client, error) {
	result := &v1alpha1.OAuth2Client{}
	err := r.Service().DeleteInNamespace(namespace, name, result)
	return result, err
}

func (r *Resolver) OAuth2ClientSubscription(ctx context.Context, namespace string) (<-chan *gqlschema.OAuth2ClientEvent, error) {
	channel := make(chan *gqlschema.OAuth2ClientEvent, 1)
	filter := func(client v1alpha1.OAuth2Client) bool {
		return client.Namespace == namespace
	}

	unsubscribe, err := r.Service().Subscribe(NewEventHandler(channel, filter))
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(channel)
		defer unsubscribe()
		<-ctx.Done()
	}()

	return channel, nil
}

func (r *Resolver) ErrorField(ctx context.Context, obj *v1alpha1.OAuth2Client) (*v1alpha1.ReconciliationError, error) {
	if obj.Status.ReconciliationError.Code == "" {
		return nil, nil
	}
	return &obj.Status.ReconciliationError, nil
}
