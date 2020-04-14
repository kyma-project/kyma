package configwatcher

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	baseNamespace      = "base-namespace"
	excludedNamespace1 = "excluded1"
	excludedNamespace2 = "excluded2"
)

var (
	excludedNamespaces = []string{excludedNamespace1, excludedNamespace2}
)

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
				ConfigLabel:  RuntimeLabelValue,
				RuntimeLabel: runtimeLabel,
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
				ConfigLabel:      CredentialsLabelValue,
				CredentialsLabel: credentialLabel,
			},
		},
		Data: map[string][]byte{
			"foo": []byte(`bar`),
			"bar": []byte(`foo`),
		},
		Type: corev1.SecretTypeBasicAuth,
	}
}

func fixServiceAccount(name, namespace, registryCredentials, imagePull string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				ConfigLabel: ServiceAccountLabelValue,
			},
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: registryCredentials,
			},
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			{
				Name: imagePull,
			},
		},
	}
}
