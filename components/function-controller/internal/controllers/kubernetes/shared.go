package kubernetes

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ConfigLabel              = "serverless.kyma-project.io/config"
	RuntimeLabel             = "serverless.kyma-project.io/runtime"
	RbacLabel                = "serverless.kyma-project.io/rbac"
	RoleLabelValue           = "role"
	RoleBindingLabelValue    = "rolebinding"
	CredentialsLabelValue    = "credentials"
	ServiceAccountLabelValue = "service-account"
	RuntimeLabelValue        = "runtime"
)

type Config struct {
	BaseNamespace                 string        `envconfig:"default=kyma-system"`
	ExcludedNamespaces            []string      `envconfig:"default=istio-system;kube-node-lease;kube-public;kube-system;kyma-installer;kyma-integration;kyma-system;natss;compass-system"`
	RoleRequeueDuration           time.Duration `envconfig:"default=1m"`
	RoleBindingRequeueDuration    time.Duration `envconfig:"default=1m"`
	ConfigMapRequeueDuration      time.Duration `envconfig:"default=1m"`
	SecretRequeueDuration         time.Duration `envconfig:"default=1m"`
	ServiceAccountRequeueDuration time.Duration `envconfig:"default=1m"`
}

func getNamespaces(ctx context.Context, client client.Client, base string, excluded []string) ([]string, error) {
	var namespaces corev1.NamespaceList
	if err := client.List(ctx, &namespaces); err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, namespace := range namespaces.Items {
		if !isExcludedNamespace(namespace.GetName(), base, excluded) && namespace.Status.Phase != corev1.NamespaceTerminating {
			names = append(names, namespace.GetName())
		}
	}

	return names, nil
}

func isExcludedNamespace(name, base string, excluded []string) bool {
	if name == base {
		return true
	}

	for _, namespace := range excluded {
		if name == namespace {
			return true
		}
	}

	return false
}
