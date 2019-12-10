package mockice

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
)

var (
	podPort      int32  = 8080
	svcPort      int32  = 80
	defaultImage string = "hudymi/mockice:0.1.3"
)

func Start(client dynamic.Interface, namespace, name string) (string, error) {
	_, err := createConfigMap(client, namespace, name)
	if err != nil {
		return "", err
	}

	_, err = createPod(client, namespace, name)
	if err != nil {
		Stop(client, namespace, name)
		return "", err
	}

	_, err = createService(client, namespace, name)
	if err != nil {
		Stop(client, namespace, name)
		return "", err
	}

	return fmt.Sprintf("%s.%s.svc.cluster.local:%d", name, namespace, svcPort), nil
}

func Stop(client dynamic.Interface, namespace, name string) {
	logOnDeleteError(deleteResource(client, "configmaps", namespace, name), "ConfigMap", namespace, name)
	logOnDeleteError(deleteResource(client, "pods", namespace, name), "Pod", namespace, name)
	logOnDeleteError(deleteResource(client, "services", namespace, name), "Service", namespace, name)
}

func logOnDeleteError(err error, kind, namespace, name string) {
	if err != nil {
		log.Println(fmt.Sprintf("Cannot delete %s %s/%s, because: %v", kind, namespace, name, err))
	}
}

func ReadmeURL(host string) string {
	return fmt.Sprintf("http://%s/README.md", host)
}

func AsynAPIFileURL(host string) string {
	return fmt.Sprintf("http://%s/streetlights.yml", host)
}

func deleteResource(client dynamic.Interface, resource, namespace, name string) error {
	groupVersion := schema.GroupVersionResource{Group: "", Version: "v1", Resource: resource}
	return client.Resource(groupVersion).Namespace(namespace).Delete(name, nil)
}

func createConfigMap(client dynamic.Interface, namespace, name string) (*v1.ConfigMap, error) {
	configMap := fixConfigMap(namespace, name)
	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&configMap)
	if err != nil {
		return nil, errors.Wrap(err, "while converting ConfigMap to map[string]interface{}")
	}

	configMap = v1.ConfigMap{}
	err = create(client, resource, namespace, obj, &configMap)
	return &configMap, err
}

func createPod(client dynamic.Interface, namespace, name string) (*v1.Pod, error) {
	pod, err := fixPod(namespace, name)
	if err != nil {
		return nil, err
	}

	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Pod to map[string]interface{}")
	}

	pod = v1.Pod{}
	err = create(client, resource, namespace, obj, &pod)
	return &pod, err
}

func createService(client dynamic.Interface, namespace, name string) (*v1.Service, error) {
	svc := fixService(namespace, name)
	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&svc)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Service to map[string]interface{}")
	}

	svc = v1.Service{}
	err = create(client, resource, namespace, obj, &svc)
	return &svc, err
}

func create(client dynamic.Interface, resource schema.GroupVersionResource, namespace string, unstructuredMap map[string]interface{}, obj interface{}) error {
	result, err := client.Resource(resource).Namespace(namespace).Create(&unstructured.Unstructured{Object: unstructuredMap}, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "while creating resource")
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, obj)
	if err != nil {
		return errors.Wrap(err, "while converting Unstructured resource")
	}

	return nil
}

func getResources(memory, cpu string) (map[v1.ResourceName]resource.Quantity, error) {
	memQ, err := resource.ParseQuantity(memory)
	if err != nil {
		return nil, err
	}

	cpuQ, err := resource.ParseQuantity(cpu)
	if err != nil {
		return nil, err
	}

	return map[v1.ResourceName]resource.Quantity{
		v1.ResourceCPU:    cpuQ,
		v1.ResourceMemory: memQ,
	}, nil
}

func fixPod(namespace, name string) (v1.Pod, error) {
	image := os.Getenv("MOCKICE_IMAGE")
	if image == "" {
		image = defaultImage
	}

	requests, err := getResources("2Mi", "1m")
	if err != nil {
		return v1.Pod{}, err
	}

	limits, err := getResources("8Mi", "2m")
	if err != nil {
		return v1.Pod{}, err
	}

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{"sidecar.istio.io/inject": "false"},
			Labels:      map[string]string{"owner": "console-backend-service-tests", "app": name},
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: "config",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						}},
					},
				},
			},
			Containers: []v1.Container{
				{
					Name:            "mockice",
					Image:           image,
					ImagePullPolicy: v1.PullIfNotPresent,
					Args:            []string{"--verbose", "--config", "/app/config.yaml"},
					VolumeMounts: []v1.VolumeMount{{
						Name:      "config",
						MountPath: "/app/config.yaml",
						ReadOnly:  true,
						SubPath:   "config.yaml",
					}},
					Ports: []v1.ContainerPort{{
						Name:          "http",
						ContainerPort: podPort,
						Protocol:      v1.ProtocolTCP,
					}},
					Resources: v1.ResourceRequirements{
						Requests: requests,
						Limits:   limits,
					},
				},
			},
		},
	}, nil
}

func fixConfigMap(namespace, name string) v1.ConfigMap {
	return v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{"owner": "console-backend-service-tests", "app": name},
		},
		Data: map[string]string{
			"config.yaml": fmt.Sprintf(`
address: :%d
endpoints:
- name: README.md
  defaultResponseCode: 200
  defaultResponseContent: "# Test markdown"  
  defaultResponseContentType: text/markdown; charset=utf-8
- name: streetlights.yml
  defaultResponseCode: 200
  defaultResponseContent: %q  
  defaultResponseContentType: text/plain; charset=utf-8
`, podPort, AsyncAPIFile),
		},
	}
}

func fixService(namespace, name string) v1.Service {
	return v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{"auth.istio.io/80": "NONE"},
			Labels:      map[string]string{"owner": "console-backend-service-tests", "app": name},
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Port:       svcPort,
				TargetPort: intstr.IntOrString{IntVal: podPort},
				Protocol:   v1.ProtocolTCP,
				Name:       "http",
			}},
			Selector: map[string]string{"owner": "console-backend-service-tests", "app": name},
		},
	}
}

const AsyncAPIFile = `
asyncapi: '1.0.0'
info:
  title: Streetlights API
  version: '1.0.0'
  description: |
    The Smartylighting Streetlights API allows you to remotely manage the city lights.

    ### Check out its awesome features:

    * Turn a specific streetlight on/off 🌃
    * Dim a specific streetlight 😎
    * Receive real-time information about environmental lighting conditions 📈
  license:
    name: Apache 2.0
    url: https://www.apache.org/licenses/LICENSE-2.0
baseTopic: smartylighting.streetlights.1.0

servers:
  - url: api.streetlights.smartylighting.com:{port}
    scheme: mqtt
    description: Test broker
    variables:
      port:
        description: Secure connection (TLS) is available through port 8883.
        default: '1883'
        enum:
          - '1883'
          - '8883'

security:
  - apiKey: []

topics:
  event.{streetlightId}.lighting.measured:
    parameters:
      - $ref: '#/components/parameters/streetlightId'
    publish:
      $ref: '#/components/messages/lightMeasured'

  action.{streetlightId}.turn.on:
    parameters:
      - $ref: '#/components/parameters/streetlightId'
    subscribe:
      $ref: '#/components/messages/turnOnOff'

  action.{streetlightId}.turn.off:
    parameters:
      - $ref: '#/components/parameters/streetlightId'
    subscribe:
      $ref: '#/components/messages/turnOnOff'

  action.{streetlightId}.dim:
    parameters:
      - $ref: '#/components/parameters/streetlightId'
    subscribe:
      $ref: '#/components/messages/dimLight'

components:
  messages:
    lightMeasured:
      summary: Inform about environmental lighting conditions for a particular streetlight.
      payload:
        $ref: "#/components/schemas/lightMeasuredPayload"
    turnOnOff:
      summary: Command a particular streetlight to turn the lights on or off.
      payload:
        $ref: "#/components/schemas/turnOnOffPayload"
    dimLight:
      summary: Command a particular streetlight to dim the lights.
      payload:
        $ref: "#/components/schemas/dimLightPayload"

  schemas:
    lightMeasuredPayload:
      type: object
      properties:
        lumens:
          type: integer
          minimum: 0
          description: Light intensity measured in lumens.
        sentAt:
          $ref: "#/components/schemas/sentAt"
    turnOnOffPayload:
      type: object
      properties:
        command:
          type: string
          enum:
            - on
            - off
          description: Whether to turn on or off the light.
        sentAt:
          $ref: "#/components/schemas/sentAt"
    dimLightPayload:
      type: object
      properties:
        percentage:
          type: integer
          description: Percentage to which the light should be dimmed to.
          minimum: 0
          maximum: 100
        sentAt:
          $ref: "#/components/schemas/sentAt"
    sentAt:
      type: string
      format: date-time
      description: Date and time when the message was sent.

  securitySchemes:
    apiKey:
      type: apiKey
      in: user
      description: Provide your API key as the user and leave the password empty.

  parameters:
    streetlightId:
      name: streetlightId
      description: The ID of the streetlight.
      schema:
        type: string
`
