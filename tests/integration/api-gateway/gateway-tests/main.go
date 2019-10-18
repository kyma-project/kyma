// Note: the example only works with the code within the same release/branch.
package main

// TODO: create a helper for fetching OAUTH2 token from hydra
// TODO: copy paste a helper for fetching JWT token from dex

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

var rawJSON []byte = []byte(`{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "httpbin",
    "namespace": "default",
    "labels": {
    	"foo": "bar"
    }
  },
  "spec": {
    "replicas": 1,
    "selector": {
      "matchLabels": {
        "app": "httpbin",
        "version": "v1"
      }
    },
    "template": {
      "metadata": {
        "labels": {
          "app": "httpbin",
          "version": "v1"
        }
      },
      "spec": {
        "containers": [
          {
            "image": "docker.io/kennethreitz/httpbin",
            "imagePullPolicy": "IfNotPresent",
            "name": "httpbin",
            "ports": [
              {
                "containerPort": 80
              }
            ]
          }
        ]
      }
    }
  }
}`)

var rawAPIRule = []byte(`
{
  "apiVersion": "gateway.kyma-project.io/v1alpha1",
  "kind": "APIRule",
  "metadata": {
    "name": "passthrough-unsecured",
    "namespace": "kyma-system"
  },
  "spec": {
    "service": {
      "host": "httpbin1.kyma.local",
      "name": "httpbin",
      "port": 8000
    },
    "gateway": "kyma-gateway.kyma-system.svc.cluster.local",
    "rules": [
      {
        "path": "/.*",
        "methods": [
          "GET"
        ],
        "accessStrategies": [
          {
            "handler": "noop"
          }
        ],
        "mutators": [

        ]
      }
    ]
  }
}`)

var rawYaml = []byte(`
---
apiVersion: gateway.kyma-project.io/v1alpha1
kind: APIRule
metadata:
  name: passthrough-unsecured
spec:
  service:
    host: httpbin1.kyma.local
    name: httpbin
    port: 8000
  gateway: kyma-gateway.kyma-system.svc.cluster.local
  rules:
    - path: /.*
      methods: ["GET"]
      accessStrategies:
        - handler: noop
      mutators: []
`)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	deployment, err := parseManifest(rawAPIRule)
	if err != nil {
		panic(err)
	}

	// TODO: load manifests from files and pass them to run function
	// TODO: parse yaml to json

	run(client, deployment)
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

func parseManifest(input []byte) (*unstructured.Unstructured, error) {
	var middleware map[string]interface{}
	err := json.Unmarshal(input, &middleware)
	if err != nil {
		return nil, err
	}

	resource := &unstructured.Unstructured{
		Object: middleware,
	}
	return resource, nil
}

func plurarForm(name string) string {
	return fmt.Sprintf("%ss", strings.ToLower(name))
}

// TODO: load more than one manifests
func run(client dynamic.Interface, manifest *unstructured.Unstructured) {
	metadata := manifest.Object["metadata"].(map[string]interface{})
	apiVersion := strings.Split(fmt.Sprintf("%s", manifest.Object["apiVersion"]), "/")
	namespace := "default"
	if metadata["namespace"] != nil {
		namespace = fmt.Sprintf("%s", metadata["namespace"])
	}
	resourceName := fmt.Sprintf("%s", manifest.Object["kind"])

	resourceSchema := schema.GroupVersionResource{Group: apiVersion[0], Version: apiVersion[1], Resource: plurarForm(resourceName)}
	fmt.Printf("---\n%v\n%v\n%v", resourceSchema, manifest, namespace)

	// TODO: apply create to all manifests
	// Create Deployment
	prompt()
	// TODO: move to generic function
	fmt.Println("Creating resource...")
	result, err := client.Resource(resourceSchema).Namespace(namespace).Create(manifest, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created resource %q.\n", result.GetName())

	// TODO: invoke tests here

	// TODO: apply delete to all manifests
	// Delete Deployment
	prompt()
	// TODO: move to generic function
	fmt.Println("Deleting resource...")
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	name := result.GetName()
	if err := client.Resource(resourceSchema).Namespace(namespace).Delete(name, deleteOptions); err != nil {
		panic(err)
	}

	fmt.Println("Deleted resource.")
}
