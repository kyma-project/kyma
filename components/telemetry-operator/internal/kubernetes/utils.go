package kubernetes

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Utils struct {
	client client.Client
}

func NewUtils(client client.Client) *Utils {
	return &Utils{
		client: client,
	}
}

func (u *Utils) GetOrCreateConfigMap(ctx context.Context, name types.NamespacedName) (corev1.ConfigMap, error) {
	cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name.Name, Namespace: name.Namespace}}
	err := u.getOrCreate(ctx, &cm)
	if err != nil {
		return corev1.ConfigMap{}, err
	}
	return cm, nil
}

func (u *Utils) GetOrCreateSecret(ctx context.Context, name types.NamespacedName) (corev1.Secret, error) {
	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name.Name, Namespace: name.Namespace}}
	err := u.getOrCreate(ctx, &secret)
	if err != nil {
		return corev1.Secret{}, err
	}
	return secret, nil
}

// Gets or creates the given obj in the Kubernetes cluster. obj must be a struct pointer so that obj can be updated with the content returned by the Server.
func (u *Utils) getOrCreate(ctx context.Context, obj client.Object) error {
	err := u.client.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	if err != nil && errors.IsNotFound(err) {
		return u.client.Create(ctx, obj)
	}
	return err
}
