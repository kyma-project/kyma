package config

import (
	"time"

	resource_watcher "github.com/kyma-project/kyma/components/function-controller/pkg/resource-watcher"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	baseNamespace      = "base-namespace"
	excludedNamespace1 = "excluded1"
	excludedNamespace2 = "excluded2"
	namespaceName      = "namespace"
	credentialName     = "credential"
	runtimeName        = "runtime"
	serviceAccountName = "sa"
)

var (
	excludedNamespaces = []string{excludedNamespace1, excludedNamespace2}
)

func fixServicesForController(config resource_watcher.Config) *resource_watcher.Services {
	fixNamespace := fixNamespace(namespaceName, false)
	fixCredential := fixCredential(credentialName, baseNamespace, resource_watcher.RegistryCredentialsLabelValue)
	fixRuntime := fixRuntime(runtimeName, baseNamespace, "foo")
	fixSA := fixServiceAccount(serviceAccountName, baseNamespace, "credential")

	clientset := fixFakeClientset(fixNamespace, fixCredential, fixRuntime, fixSA)
	return resource_watcher.NewResourceWatcherServices(clientset.CoreV1(), config)
}

func fixConfigForController() resource_watcher.Config {
	return resource_watcher.Config{
		EnableControllers:       true,
		BaseNamespace:           baseNamespace,
		ExcludedNamespaces:      excludedNamespaces,
		NamespaceRelistInterval: 60 * time.Hour,
	}
}

func fixFakeClientset(objects ...runtime.Object) *fake.Clientset {
	return fake.NewSimpleClientset(objects...)
}

func fixNamespace(name string, hasTerminatingStatus bool) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
	}

	if hasTerminatingStatus {
		ns.Status.Phase = corev1.NamespaceTerminating
	}

	return ns
}

func fixRuntime(name, namespace, runtimeLabel string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				resource_watcher.ConfigLabel:  resource_watcher.RuntimeLabelValue,
				resource_watcher.RuntimeLabel: runtimeLabel,
			},
		},
		Data: map[string]string{
			"foo": "bar",
			"bar": "foo",
		},
	}
}

func fixCredential(name, namespace, credentialLabel string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				resource_watcher.ConfigLabel:      resource_watcher.CredentialsLabelValue,
				resource_watcher.CredentialsLabel: credentialLabel,
			},
		},
		Data: map[string][]byte{
			"foo": []byte(`bar`),
			"bar": []byte(`foo`),
		},
		Type: corev1.SecretTypeBasicAuth,
	}
}

func fixServiceAccount(name, namespace, secretName string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				resource_watcher.ConfigLabel: resource_watcher.ServiceAccountLabelValue,
			},
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: secretName,
			},
		},
	}
}
