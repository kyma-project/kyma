package config

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/kyma/components/function-controller/pkg/configwatcher"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	baseNamespace          = "base-namespace"
	includedNamespace      = "eipsteindidntkillhimself"
	excludedNamespace      = "excluded"
	registryCredentialName = "registry-credential"
	imagePullSecretName    = "image-pull-secret"
	runtimeName            = "runtime"
	serviceAccountName     = "sa"
	runtimeLabel           = "nodejs8"
)

var (
	excludedNamespaces = []string{excludedNamespace, "kube-system", "kube-node-lease", "kube-public", "default"}
	relistInterval     = 1 * time.Second
)

var clientset *fake.Clientset

func fixServicesForController(config configwatcher.Config) *configwatcher.Services {
	fixCredential1 := fixCredential(registryCredentialName, baseNamespace, registryCredentialName, nil)
	fixCredential2 := fixCredential(imagePullSecretName, baseNamespace, imagePullSecretName, nil)
	fixRuntime := fixRuntime(runtimeName, baseNamespace, runtimeLabel, nil)
	fixSA := fixServiceAccount(serviceAccountName, baseNamespace, map[string]string{
		"foo": "bar",
	})

	clientset = fixFakeClientset(fixCredential1, fixCredential2, fixRuntime, fixSA)
	return configwatcher.NewConfigWatcherServices(clientset.CoreV1(), config)
}

func fixConfigForController() configwatcher.Config {
	return configwatcher.Config{
		EnableControllers:       true,
		BaseNamespace:           baseNamespace,
		ExcludedNamespaces:      excludedNamespaces,
		NamespaceRelistInterval: relistInterval,
	}
}

func fixFakeClientset(objects ...runtime.Object) *fake.Clientset {
	return fake.NewSimpleClientset(objects...)
}

func fixNamespace(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
	}
}

func fixRuntime(name, namespace, runtimeLabel string, additionalLabels map[string]string) *corev1.ConfigMap {
	cm := corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				configwatcher.ConfigLabel:  configwatcher.RuntimeLabelValue,
				configwatcher.RuntimeLabel: runtimeLabel,
			},
		},
		Data: map[string]string{
			"foo": "bar",
			"bar": "foo",
		},
	}

	if additionalLabels != nil {
		for key, value := range additionalLabels {
			cm.Labels[key] = value
		}
	}

	return &cm
}

func fixCredential(name, namespace, credentialLabel string, additionalLabels map[string]string) *corev1.Secret {
	secret := corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				configwatcher.ConfigLabel:      configwatcher.CredentialsLabelValue,
				configwatcher.CredentialsLabel: credentialLabel,
			},
		},
		Data: map[string][]byte{
			"username": []byte(`bar`),
			"password": []byte(`foo`),
		},
		Type: corev1.SecretTypeBasicAuth,
	}

	if additionalLabels != nil {
		for key, value := range additionalLabels {
			secret.Labels[key] = value
		}
	}

	return &secret
}

func fixServiceAccount(name, namespace string, additionalLabels map[string]string) *corev1.ServiceAccount {
	serviceAccount := corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				configwatcher.ConfigLabel: configwatcher.ServiceAccountLabelValue,
			},
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: registryCredentialName,
			},
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			{
				Name: imagePullSecretName,
			},
		},
	}

	if additionalLabels != nil {
		for key, value := range additionalLabels {
			serviceAccount.Labels[key] = value
		}
	}

	return &serviceAccount
}

func getConfigMap(name, namespace string) (*corev1.ConfigMap, error) {
	return clientset.CoreV1().ConfigMaps(namespace).Get(name, v1.GetOptions{})
}

func deleteConfigMap(name, namespace string) error {
	return clientset.CoreV1().ConfigMaps(namespace).Delete(name, &v1.DeleteOptions{})
}

func getSecret(name, namespace string) (*corev1.Secret, error) {
	return clientset.CoreV1().Secrets(namespace).Get(name, v1.GetOptions{})
}

func deleteSecret(name, namespace string) error {
	return clientset.CoreV1().Secrets(namespace).Delete(name, &v1.DeleteOptions{})
}

func getServiceAccount(name, namespace string) (*corev1.ServiceAccount, error) {
	return clientset.CoreV1().ServiceAccounts(namespace).Get(name, v1.GetOptions{})
}

func deleteServiceAccount(name, namespace string) error {
	return clientset.CoreV1().ServiceAccounts(namespace).Delete(name, &v1.DeleteOptions{})
}
