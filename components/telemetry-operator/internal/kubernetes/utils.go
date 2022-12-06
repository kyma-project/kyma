package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetOrCreateConfigMap(ctx context.Context, c client.Client, name types.NamespacedName) (corev1.ConfigMap, error) {
	cm := corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: name.Name, Namespace: name.Namespace}}
	err := c.Get(ctx, client.ObjectKeyFromObject(&cm), &cm)
	if err == nil {
		return cm, nil
	}
	if errors.IsNotFound(err) {
		err = c.Create(ctx, &cm)
		if err == nil {
			return cm, nil
		}
	}
	return corev1.ConfigMap{}, err
}

func GetOrCreateSecret(ctx context.Context, c client.Client, name types.NamespacedName) (corev1.Secret, error) {
	secret := corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: name.Name, Namespace: name.Namespace}}
	err := c.Get(ctx, client.ObjectKeyFromObject(&secret), &secret)
	if err == nil {
		return secret, nil
	}
	if errors.IsNotFound(err) {
		err = c.Create(ctx, &secret)
		if err == nil {
			return secret, nil
		}
	}
	return corev1.Secret{}, err
}
