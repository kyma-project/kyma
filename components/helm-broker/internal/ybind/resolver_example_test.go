//+build integration

package ybind_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/ybind"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func ExampleNewResolver() {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	fatalOnErr(err)

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	fatalOnErr(err)

	// create namespace for test
	nsSpec := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "resolver-example-test"}}
	_, err = clientset.CoreV1().Namespaces().Create(nsSpec)
	defer clientset.CoreV1().Namespaces().Delete(nsSpec.ObjectMeta.Name, &metav1.DeleteOptions{})
	fatalOnErr(err)

	_, err = clientset.CoreV1().Secrets(nsSpec.Name).Create(&v1.Secret{
		Type: v1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name: "single-secret-test-redis",
		},
		StringData: map[string]string{
			// The serialized form of the secret data is a base64 encoded string, so we need to pass here raw data
			"redis-password": "gopherek",
		},
	})
	fatalOnErr(err)

	_, err = clientset.CoreV1().Secrets(nsSpec.Name).Create(&v1.Secret{
		Type: v1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name: "all-secret-test-redis",
		},
		StringData: map[string]string{
			// The serialized form of the secret data is a base64 encoded string, so we need to pass here raw data
			"secret-key-no-1": "piko",
			"secret-key-no-2": "bello",
		},
	})
	fatalOnErr(err)

	_, err = clientset.CoreV1().ConfigMaps(nsSpec.Name).Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "single-cfg-map-test-redis",
		},
		Data: map[string]string{
			"username": "redisMaster",
		},
	})
	fatalOnErr(err)

	_, err = clientset.CoreV1().ConfigMaps(nsSpec.Name).Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "all-cfg-map-test-redis",
		},
		Data: map[string]string{
			"cfg-key-no-1": "margarita",
			"cfg-key-no-2": "capricciosa",
		},
	})
	fatalOnErr(err)

	_, err = clientset.CoreV1().Services(nsSpec.Name).Create(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-renderer-test-redis",
		},
		Spec: v1.ServiceSpec{
			Type:     v1.ServiceTypeNodePort,
			Selector: map[string]string{"app": "some-app"},
			Ports: []v1.ServicePort{
				{
					Name: "redis",
					Port: 123,
				},
			},
		},
	})

	fatalOnErr(err)

	resolver := ybind.NewResolver(clientset.CoreV1())
	out, err := resolver.Resolve(fixBindYAML(), internal.Namespace(nsSpec.Name))
	fatalOnErr(err)

	printSorted(out.Credentials)

	// Output:
	// key: HOST_PORT, value: 123
	// key: REDIS_PASSWORD, value: gopherek
	// key: REDIS_USERNAME, value: redisMaster
	// key: URL, value: host1-example-renderer-test-redis.ns-name.svc.cluster.local:6379
	// key: cfg-key-no-1, value: override-value
	// key: cfg-key-no-2, value: capricciosa
	// key: secret-key-no-1, value: piko
	// key: secret-key-no-2, value: bello
}

func printSorted(m map[string]string) {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf("key: %s, value: %s\n", key, m[key])
	}

}

func fixBindYAML() []byte {
	return []byte(`
credential:
  - name: cfg-key-no-1
    value: override-value
  - name: URL
    value: host1-example-renderer-test-redis.ns-name.svc.cluster.local:6379
  - name: HOST_PORT
    valueFrom:
      serviceRef:
        name: example-renderer-test-redis
        jsonpath: '{ .spec.ports[?(@.name=="redis")].port }'
  - name: REDIS_PASSWORD
    valueFrom:
      secretKeyRef:
        name: single-secret-test-redis
        key: redis-password
  - name: REDIS_USERNAME
    valueFrom:
      configMapKeyRef:
        name: single-cfg-map-test-redis
        key: username

credentialFrom:
  - configMapRef:
     name: all-cfg-map-test-redis
  - secretRef:
      name:  all-secret-test-redis
`)
}
