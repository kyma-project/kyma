package istio

import (
	"bytes"
	_ "embed"
	"io"
	"log"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

const (
	kubeSystemNamespace    = "kube-system"
	proxyName              = "istio-proxy"
	istioNamespace         = "istio-system"
	evalProfile            = "evaluation"
	prodProfile            = "production"
	deployedKymaProfileVar = "KYMA_PROFILE"
	exportResultVar        = "EXPORT_RESULT"
	junitFileName          = "junit-report.xml"
	cucumberFileName       = "cucumber-report.json"
)

var k8sClient kubernetes.Interface
var dynamicClient dynamic.Interface
var mapper *restmapper.DeferredDiscoveryRESTMapper

//go:embed test/httpbin.yaml
var httpbin []byte

func readManifestToUnstructured() ([]unstructured.Unstructured, error) {
	var err error
	var unstructList []unstructured.Unstructured

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(httpbin), 200)
	for {
		var rawObj runtime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}
		obj, _, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			break
		}
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			break
		}
		unstructuredObj := unstructured.Unstructured{Object: unstructuredMap}
		unstructList = append(unstructList, unstructuredObj)
	}
	if err != io.EOF {
		return unstructList, err
	}

	return unstructList, nil
}

func getGroupVersionResource(resource unstructured.Unstructured) schema.GroupVersionResource {
	mapping, err := mapper.RESTMapping(resource.GroupVersionKind().GroupKind(), resource.GroupVersionKind().Version)
	if err != nil {
		log.Fatal(err)
	}
	return mapping.Resource
}
