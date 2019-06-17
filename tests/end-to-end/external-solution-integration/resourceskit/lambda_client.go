package resourceskit

import (
	"encoding/json"

	kubelessV1 "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubeless "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/consts"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const lambdaFunction = `
	const request = require('request')
	module.exports = { main: function (event, context) {
		return new Promise((resolve, reject) => {
			const url = process.env.GATEWAY_URL + "/counter"
			sendReq(url, resolve, reject)
		})
	} }
	function sendReq(url, resolve, reject) {
        request.post(url, { json: true }, (error, response, body) => {
            if (error) {
                resolve(error)
            }
            resolve(response) 
        })
    }`

type LambdaClient interface {
	DeployLambda() error
	DeleteLambda() error
	IsLambdaReady() (bool, error)
	IsLambdaReadyWithSBU() (bool, error)
	IsFunctionAnnotated() (bool, error)
}

type lambdaClient struct {
	coreClient        *kubernetes.Clientset
	kubelessClientSet *kubeless.Clientset
	namespace         string
}

type injectLabelInfo struct {
	InjectedLabels map[string]string `json:"injectedLabels"`
}

type tracingInformation map[string]injectLabelInfo

func NewLambdaClient(config *rest.Config, namespace string) (LambdaClient, error) {
	coreClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	kubelessClientSet, err := kubeless.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &lambdaClient{
		coreClient:        coreClientSet,
		kubelessClientSet: kubelessClientSet,
		namespace:         namespace,
	}, nil
}

func (c *lambdaClient) DeployLambda() error {
	log.WithFields(log.Fields{"AppName": consts.AppName}).Debug("Creating Lambda")
	lambda := c.createLambda()

	_, err := c.kubelessClientSet.KubelessV1beta1().Functions(c.namespace).Create(lambda)
	if err != nil {
		return err
	}

	return nil
}

func (c *lambdaClient) DeleteLambda() error {
	log.WithFields(log.Fields{"AppName": consts.AppName}).Debug("Deleting Lambda")
	return c.kubelessClientSet.KubelessV1beta1().Functions(c.namespace).Delete(consts.AppName, &metav1.DeleteOptions{})
}

func (c *lambdaClient) createLambda() *kubelessV1.Function {
	lambdaSpec := kubelessV1.FunctionSpec{
		Handler:             "handler.main",
		Function:            lambdaFunction,
		FunctionContentType: "text",
		Runtime:             "nodejs8",
		Deps:                `{"dependencies":{"request": "^2.88.0"}}`,
		Deployment: v1beta1.Deployment{
			TypeMeta:   metav1.TypeMeta{Kind: "Deployment", APIVersion: v1beta1.SchemeGroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{Name: consts.AppName, Namespace: c.namespace, Labels: map[string]string{"function": consts.AppName}},
			Spec: v1beta1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: consts.AppName,
							},
						},
					},
				},
			},
		},
		ServiceSpec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http-function-port",
					Port:       8080,
					Protocol:   "TCP",
					TargetPort: intstr.FromInt(8080),
				},
			},
			Selector: map[string]string{"created-by": "kubeless", "function": consts.AppName},
		},
	}

	return &kubelessV1.Function{
		TypeMeta:   metav1.TypeMeta{Kind: "Function", APIVersion: kubelessV1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: consts.AppName, Namespace: c.namespace},
		Spec:       lambdaSpec,
	}
}

func (c *lambdaClient) GetLamda() (*kubelessV1.Function, error) {
	log.WithFields(log.Fields{"name": consts.AppName}).Debug("Retrieving Lamda")
	return c.kubelessClientSet.KubelessV1beta1().Functions(c.namespace).Get(consts.AppName, metav1.GetOptions{})
}

func (c *lambdaClient) getSBULabel(sbuName string) (map[string]string, error) {
	function, err := c.GetLamda()
	if err != nil {
		return nil, err
	}

	if val, ok := function.GetObjectMeta().GetAnnotations()["servicebindingusages.servicecatalog.kyma-project.io/tracing-information"]; ok {
		info := tracingInformation{}
		err := json.Unmarshal([]byte(val), &info)
		if err != nil {
			log.Debug(err)
			return nil, err
		}
		if sbulabel, ok := info[sbuName]; ok {
			return sbulabel.InjectedLabels, nil
		}
	}
	return map[string]string{}, nil

}

func (c *lambdaClient) IsLambdaReadyWithSBU() (bool, error) {
	log.WithFields(log.Fields{}).Debug("Checking Lambda(SBU)")

	labelMap, err := c.getSBULabel(consts.ServiceBindingUsageName)
	if err != nil {
		return false, err
	}

	labelMap["function"] = consts.AppName
	labelMap["created-by"] = "kubeless"

	labelSelector := &metav1.LabelSelector{
		MatchLabels: labelMap,
	}
	return c.isLambdaReadySelector(labelSelector)
}

func (c *lambdaClient) IsLambdaReady() (bool, error) {
	log.WithFields(log.Fields{}).Debug("Checking Lambda")
	labelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"function":   consts.AppName,
			"created-by": "kubeless",
		},
	}
	return c.isLambdaReadySelector(labelSelector)

}

func (c *lambdaClient) isLambdaReadySelector(labelSelector *metav1.LabelSelector) (bool, error) {
	log.WithFields(log.Fields{}).Debug("Checking Lambda")
	labelMap, err := metav1.LabelSelectorAsMap(labelSelector)
	if err != nil {
		return false, err
	}

	listOptions := metav1.ListOptions{LabelSelector: labels.SelectorFromSet(labelMap).String()}
	podList, e := c.ListPods(listOptions)

	pods := podList.Items

	if e != nil {
		return false, e
	}

	if len(pods) == 0 {
		log.Debug("No Functionpods found")
		return false, nil
	}

	for _, pod := range pods {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady {
				if condition.Status != corev1.ConditionTrue {
					return false, nil
				}
			}
		}
	}

	return true, nil
}

func (c *lambdaClient) IsFunctionAnnotated() (bool, error) {
	labelMap, err := c.getSBULabel(consts.ServiceBindingUsageName)
	if err != nil {
		return false, err
	}
	if len(labelMap) == 0 {
		return false, nil
	}
	return true, nil
}

func (c *lambdaClient) ListPods(options metav1.ListOptions) (*corev1.PodList, error) {
	return c.coreClient.CoreV1().Pods(c.namespace).List(options)
}
