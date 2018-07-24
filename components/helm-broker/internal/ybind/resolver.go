package ybind

import (
	"bytes"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/jsonpath"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

// Resolver implements resolver for chart values.
type Resolver struct {
	clientCoreV1 corev1.CoreV1Interface
}

// NewResolver returns new instance of Resolver.
func NewResolver(clientCoreV1 corev1.CoreV1Interface) *Resolver {
	return &Resolver{
		clientCoreV1: clientCoreV1,
	}
}

// ResolveOutput represents results of Resolve.
type ResolveOutput struct {
	Credentials internal.InstanceCredentials
}

// Resolve determines the final value of credentials
//
// Resolve policy rules
// 1.  When a key exists in multiple sources defined by `credentialFrom` section, then the value associated with the last source will take precedence
// 2.  When you duplicate a key in `credential` section then error will be returned
// 3.  Values defined by `credentialFrom` section will be overridden by values from `credential` section if keys will be duplicated
func (r *Resolver) Resolve(bindYAML RenderedBindYAML, ns internal.Namespace) (*ResolveOutput, error) {
	var bind BindYAML
	if err := yaml.Unmarshal(bindYAML, &bind); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling bind yaml")
	}

	credFrom := credentials{}
	for _, v := range bind.CredentialFrom {
		envs, err := r.getCredFromAllRefValues(ns, v)
		if err != nil {
			return nil, err
		}
		//Policy no. 1: Use Set method to allow override value.
		for k, v := range envs {
			credFrom.Set(k, v)
		}
	}

	cred := credentials{}
	for _, v := range bind.Credential {
		if v.Value != "" {
			// Policy no. 2: Use Insert method to return error on duplication.
			if err := cred.Insert(v.Name, v.Value); err != nil {
				return nil, err
			}
		} else if v.ValueFrom != nil {
			val, err := r.getCredVarKeyRefValue(ns, *v.ValueFrom)
			if err != nil {
				return nil, err
			}
			// Policy no. 2: Use Insert method to return error on duplication.
			if err := cred.Insert(v.Name, val); err != nil {
				return nil, err
			}
		}
	}

	// Policy no. 3: Merge `credential` vars to `credentialFrom` vars to allow replace existing keys.
	for k, v := range cred {
		credFrom.Set(k, v)
	}

	return &ResolveOutput{
		Credentials: internal.InstanceCredentials(credFrom),
	}, nil
}

// getCredFromAllRefValues returns the key-value pairs referenced by the given CredentialFromSource in the supplied namespace
func (r *Resolver) getCredFromAllRefValues(ns internal.Namespace, from CredentialFromSource) (map[string]string, error) {
	if from.ConfigMapRef != nil {
		return getConfigMapAllValues(r.clientCoreV1, ns, *from.ConfigMapRef)
	}

	if from.SecretRef != nil {
		return getSecretAllValues(r.clientCoreV1, ns, *from.SecretRef)
	}

	return map[string]string{}, fmt.Errorf("invalid credentialFrom")
}

// getCredVarKeyRefValue returns the value referenced by the given CredentialVarSource in the supplied namespace
func (r *Resolver) getCredVarKeyRefValue(ns internal.Namespace, from CredentialVarSource) (string, error) {
	if from.SecretKeyRef != nil {
		return getSecretKeyValue(r.clientCoreV1, ns, *from.SecretKeyRef)
	}

	if from.ConfigMapKeyRef != nil {
		return getConfigMapKeyValue(r.clientCoreV1, ns, *from.ConfigMapKeyRef)
	}

	if from.ServiceRef != nil {
		return getServiceJSONPathValue(r.clientCoreV1, ns, *from.ServiceRef)
	}
	return "", fmt.Errorf("invalid valueFrom")
}

// getSecretAllValues returns key-value pairs populated from secret in the supplied namespace
func getSecretAllValues(client corev1.CoreV1Interface, namespace internal.Namespace, secretSelector NameSelector) (map[string]string, error) {
	secret, err := client.Secrets(string(namespace)).Get(secretSelector.Name, metav1.GetOptions{})
	if err != nil {
		return map[string]string{}, errors.Wrapf(err, "while getting secrets [%s] from namespace [%s]", secretSelector.Name, namespace)
	}

	envs := map[string]string{}
	for k, v := range secret.Data {
		envs[k] = string(v)
	}

	return envs, nil
}

// getConfigMapAllValues returns key-value pairs populated from configmap in the supplied namespace
func getConfigMapAllValues(client corev1.CoreV1Interface, namespace internal.Namespace, configMapSelector NameSelector) (map[string]string, error) {
	configMap, err := client.ConfigMaps(string(namespace)).Get(configMapSelector.Name, metav1.GetOptions{})
	if err != nil {
		return map[string]string{}, errors.Wrapf(err, "while getting configmap [%s] from namespace [%s]", configMapSelector.Name, namespace)
	}

	return configMap.Data, nil
}

// getSecretKeyValue returns the value of a secret in the supplied namespace
func getSecretKeyValue(client corev1.CoreV1Interface, namespace internal.Namespace, secretSelector KeySelector) (string, error) {
	secret, err := client.Secrets(string(namespace)).Get(secretSelector.Name, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "while getting secrets [%s] from namespace [%s]", secretSelector.Name, namespace)
	}

	data, ok := secret.Data[secretSelector.Key]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret %s in namespace %s", secretSelector.Key, secretSelector.Name, namespace)
	}

	return string(data), nil
}

// getConfigMapKeyValue returns the value of a configmap in the supplied namespace
func getConfigMapKeyValue(client corev1.CoreV1Interface, namespace internal.Namespace, configMapSelector KeySelector) (string, error) {
	configMap, err := client.ConfigMaps(string(namespace)).Get(configMapSelector.Name, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "while getting configmap [%s] from namespace [%s]", configMapSelector.Name, namespace)
	}
	data, ok := configMap.Data[configMapSelector.Key]
	if !ok {
		return "", fmt.Errorf("key %s not found in config map %s in namespace %s", configMapSelector.Key, configMapSelector.Name, namespace)
	}
	return string(data), nil
}

func getServiceJSONPathValue(kubeClient corev1.CoreV1Interface, namespace internal.Namespace, selector JSONPathSelector) (string, error) {
	service, err := kubeClient.Services(string(namespace)).Get(selector.Name, metav1.GetOptions{})
	if err != nil {
		return "", errors.Wrapf(err, "while getting service [%s] from namespace [%s]", selector.Name, namespace)
	}
	pathParser := jsonpath.New("service pathParser parser")
	if err := pathParser.Parse(selector.JSONPath); err != nil {
		return "", errors.Wrapf(err, "while parsing json path [%s] for service [%s] in namespace [%s]", selector.JSONPath, selector.Name, namespace)
	}
	out := bytes.Buffer{}
	if err := pathParser.Execute(&out, service); err != nil {
		return "", errors.Wrapf(err, "while selecting json path [%s] for service [%s] in namespace [%s]", selector.JSONPath, selector.Name, namespace)
	}
	return out.String(), nil
}

type credentials map[string]string

func (e *credentials) Insert(name, value string) error {
	if _, exists := (*e)[name]; exists {
		return fmt.Errorf("conflict: found credentials with the same name %q", name)
	}
	(*e)[name] = value
	return nil
}

func (e *credentials) Set(name, value string) {
	(*e)[name] = value
}
