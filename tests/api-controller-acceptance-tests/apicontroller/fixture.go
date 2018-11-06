package apicontroller

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type Fixture struct {
	sampleAppCtrl    *sampleAppCtrl
	SampleAppDepl    *appv1.Deployment
	SampleAppService *corev1.Service
}

func setUpOrExit(k8sInterface kubernetes.Interface, namespace string, testId string) *Fixture {

	sampleAppCtrl := &sampleAppCtrl{
		k8sInterface: k8sInterface,
		namespace:    namespace,
		testId:       testId,
	}

	sampleAppDepl, deplErr := sampleAppCtrl.createDeployment()
	if deplErr != nil {

		if errors.IsAlreadyExists(deplErr) {
			log.Debug("SampleApp deployment already exists.")
			if sampleAppDepl2, getDeplErr := sampleAppCtrl.getDeployment(); getDeplErr != nil {
				log.Fatalf("error getting existing SampleApp deployment. Root cause: %v", getDeplErr)
			} else {
				sampleAppDepl = sampleAppDepl2
			}
		} else {
			log.Fatalf("error creating SampleApp deployment. Root cause: %v", deplErr)
		}
	}
	sampleAppService, svcErr := sampleAppCtrl.createService(&sampleAppDepl.Spec.Template)
	if svcErr != nil {
		if errors.IsAlreadyExists(deplErr) {
			log.Debug("SampleApp service already exists.")
			if sampleAppSvc2, getSvcErr := sampleAppCtrl.getService(); getSvcErr != nil {
				log.Fatalf("error getting existing SampleApp service. Root cause: %v", getSvcErr)
			} else {
				sampleAppService = sampleAppSvc2
			}
		} else {
			log.Fatalf("error creating SampleApp service. Root cause: %v", svcErr)
		}
	}

	return &Fixture{
		sampleAppCtrl:    sampleAppCtrl,
		SampleAppDepl:    sampleAppDepl,
		SampleAppService: sampleAppService,
	}
}

func (f *Fixture) tearDown() {

	if err := f.sampleAppCtrl.deleteDeployment(f.SampleAppDepl); err != nil {
		if errors.IsNotFound(err) {
			log.Warn("SampleApp deployment does not exist")
		} else {
			log.Errorf("Error while deleting SampleApp deployment. Root cause: %v", err)
		}
	}
	if err := f.sampleAppCtrl.deleteService(f.SampleAppService); err != nil {
		if errors.IsNotFound(err) {
			log.Warn("SampleApp service does not exist")
		} else {
			log.Errorf("Error while deleting SampleApp service. Root cause: %v", err)
		}
	}
}

type sampleAppCtrl struct {
	k8sInterface kubernetes.Interface
	namespace    string
	testId       string
}

func (c *sampleAppCtrl) createDeployment() (*appv1.Deployment, error) {

	labels := labels(c.testId)

	annotations := make(map[string]string)
	annotations["sidecar.istio.io/inject"] = "true"

	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.deplName(),
			Namespace: c.namespace,
			Labels:    labels,
		},
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            fmt.Sprintf("sample-app-cont-%s", c.testId),
							Image:           "eu.gcr.io/kyma-project/example/http-db-service:0.0.6",
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8017,
								},
							},
						},
					},
				},
			},
		},
	}

	return c.k8sInterface.AppsV1().Deployments(c.namespace).Create(deployment)
}

func (c *sampleAppCtrl) getDeployment() (*appv1.Deployment, error) {
	return c.k8sInterface.AppsV1().Deployments(c.namespace).Get(c.deplName(), metav1.GetOptions{})
}

func (c *sampleAppCtrl) deplName() string {
	return fmt.Sprintf("sample-app-depl-%s", c.testId)
}

func (c *sampleAppCtrl) createService(podTmpl *corev1.PodTemplateSpec) (*corev1.Service, error) {

	selectors := make(map[string]string)
	selectors["app"] = podTmpl.ObjectMeta.Labels["app"]

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.svcName(),
			Namespace: c.namespace,
			Labels:    labels(c.testId),
		},
		Spec: corev1.ServiceSpec{
			Selector: selectors,
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Port:       8888,
					Name:       "http",
					TargetPort: intstr.FromString(podTmpl.Spec.Containers[0].Ports[0].Name),
				},
			},
		},
	}

	return c.k8sInterface.CoreV1().Services(c.namespace).Create(svc)
}

func (c *sampleAppCtrl) getService() (*corev1.Service, error) {
	return c.k8sInterface.CoreV1().Services(c.namespace).Get(c.svcName(), metav1.GetOptions{})
}

func (c *sampleAppCtrl) svcName() string {
	return fmt.Sprintf("sample-app-svc-%s", c.testId)
}

func (c *sampleAppCtrl) deleteDeployment(depl *appv1.Deployment) error {
	return c.k8sInterface.AppsV1().Deployments(c.namespace).Delete(depl.Name, &metav1.DeleteOptions{})
}

func (c *sampleAppCtrl) deleteService(service *corev1.Service) error {
	return c.k8sInterface.CoreV1().Services(c.namespace).Delete(service.Name, &metav1.DeleteOptions{})
}

func labels(testId string) map[string]string {
	labels := make(map[string]string)
	labels["createdBy"] = "api-controller-acceptance-tests"
	labels["app"] = fmt.Sprintf("sample-app-%s", testId)
	labels["test"] = "true"
	return labels
}
