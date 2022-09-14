package istio

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cucumber/godog"
	"github.com/tidwall/pretty"
	"gitlab.com/rodrigoodhin/gocure/models"
	"gitlab.com/rodrigoodhin/gocure/pkg/gocure"
	"gitlab.com/rodrigoodhin/gocure/report/html"

	"github.com/mitchellh/mapstructure"
	istioOperator "istio.io/istio/operator/pkg/apis/istio/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"

	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	proxyName              = "istio-proxy"
	istioNamespace         = "istio-system"
	evalProfile            = "evaluation"
	prodProfile            = "production"
	deployedKymaProfileVar = "KYMA_PROFILE"
	exportResultVar        = "EXPORT_RESULT"
	cucumberFileName       = "cucumber-report.json"
)

var k8sClient kubernetes.Interface
var dynamicClient dynamic.Interface
var mapper *restmapper.DeferredDiscoveryRESTMapper

//go:embed test/httpbin.yaml
var httpbin []byte

func initK8sClient() (kubernetes.Interface, dynamic.Interface, *restmapper.DeferredDiscoveryRESTMapper) {
	var kubeconfig string
	if kConfig, ok := os.LookupEnv("KUBECONFIG"); !ok {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	} else {
		kubeconfig = kConfig
	}
	_, err := os.Stat(kubeconfig)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatalf("kubeconfig %s does not exist", kubeconfig)
		}
		log.Fatalf(err.Error())
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf(err.Error())
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	return k8sClient, dynClient, mapper
}

func readManifestToUnstructured() ([]unstructured.Unstructured, error) {
	var err error
	var unstructList []unstructured.Unstructured

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(httpbin), 200)
	for {
		var rawObj k8sruntime.RawExtension
		if err = decoder.Decode(&rawObj); err != nil {
			break
		}
		obj, _, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			break
		}
		unstructuredMap, err := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
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

func hasIstioProxy(containers []corev1.Container) bool {
	proxyImage := ""
	for _, container := range containers {
		if container.Name == proxyName {
			proxyImage = container.Image
		}
	}
	return proxyImage != ""
}

func getPodListReport(list *corev1.PodList) string {
	type returnedPodList struct {
		PodList []struct {
			Metadata struct {
				Name              string `json:"name"`
				CreationTimestamp string `json:"creationTimestamp"`
			} `json:"metadata"`
			Status struct {
				Phase string `json:"phase"`
			} `json:"status"`
		} `json:"items"`
	}

	p := returnedPodList{}
	toMarshal, _ := json.Marshal(list)
	_ = json.Unmarshal(toMarshal, &p)
	toPrint, _ := json.Marshal(p)
	return string(pretty.Pretty(toPrint))
}

func listPodsIstioNamespace(istiodPodsSelector metav1.ListOptions) (*corev1.PodList, error) {
	return k8sClient.CoreV1().Pods(istioNamespace).List(context.Background(), istiodPodsSelector)
}

func generateHTMLReport() {
	html := gocure.HTML{
		Config: html.Data{
			InputJsonPath:    cucumberFileName,
			OutputHtmlFolder: "reports/",
			Title:            "Kyma Istio component tests",
			Metadata: models.Metadata{
				TestEnvironment: os.Getenv(deployedKymaProfileVar),
				Platform:        runtime.GOOS,
				Parallel:        "Scenarios",
				Executed:        "Remote",
				AppVersion:      "main",
				Browser:         "default",
			},
		},
	}
	err := html.Generate()
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func getIstioOperatorFromCluster() (*istioOperator.IstioOperator, error) {
	item, err := dynamicClient.Resource(schema.GroupVersionResource{Group: "install.istio.io", Version: "v1alpha1", Resource: "istiooperators"}).Namespace("istio-system").Get(context.Background(), "installed-state-default-operator", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("default Kyma IstioOperator CR wasn't found err=%s", err)
	}

	jsonSlice, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	operator := istioOperator.IstioOperator{}

	json.Unmarshal(jsonSlice, &operator)
	return &operator, nil
}

type ResourceStruct struct {
	Cpu    string
	Memory string
}

func getResourcesForComponent(operator *istioOperator.IstioOperator, component, resourceType string) (*ResourceStruct, error) {
	res := ResourceStruct{}

	switch component {
	case "proxy_init":
		fallthrough
	case "proxy":
		jsonResources, err := json.Marshal(operator.Spec.Values.Fields["global"].GetStructValue().Fields[component].GetStructValue().
			Fields["resources"].GetStructValue().Fields[resourceType].GetStructValue())

		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(jsonResources, &res)
		if err != nil {
			return nil, err
		}
		return &res, nil
	case "ingress-gateway":
		if resourceType == "limits" {
			err := mapstructure.Decode(operator.Spec.Components.IngressGateways[0].K8S.Resources.Limits, &res)
			if err != nil {
				return nil, err
			}
		} else {
			err := mapstructure.Decode(operator.Spec.Components.IngressGateways[0].K8S.Resources.Requests, &res)
			if err != nil {
				return nil, err
			}
		}

		return &res, nil
	case "egress-gateway":
		if resourceType == "limits" {
			err := mapstructure.Decode(operator.Spec.Components.EgressGateways[0].K8S.Resources.Limits, &res)
			if err != nil {
				return nil, err
			}
		} else {
			err := mapstructure.Decode(operator.Spec.Components.EgressGateways[0].K8S.Resources.Requests, &res)
			if err != nil {
				return nil, err
			}
		}
		return &res, nil
	case "pilot":
		if resourceType == "limits" {
			err := mapstructure.Decode(operator.Spec.Components.Pilot.K8S.Resources.Limits, &res)
			if err != nil {
				return nil, err
			}
		} else {
			err := mapstructure.Decode(operator.Spec.Components.Pilot.K8S.Resources.Requests, &res)
			if err != nil {
				return nil, err
			}
		}
		return &res, nil
	default:
		return nil, godog.ErrPending
	}
}
