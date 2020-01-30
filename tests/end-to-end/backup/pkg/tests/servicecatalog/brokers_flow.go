package servicecatalog

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/sirupsen/logrus"

	bu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"

	appsTypes "k8s.io/api/apps/v1"

	k8sCoreTypes "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/kubectl/genericclioptions/printers"
)

const (
	envTesterName = "env-tester"
)

type brokersFlow struct {
	namespace string
	log       logrus.FieldLogger

	k8sInterface kubernetes.Interface
	scInterface  clientset.Interface
	buInterface  bu.Interface
}

func (f *brokersFlow) createEnvTester(testedEnv string) error {
	f.log.Infof("Creating environment variable tester")

	labels := map[string]string{
		"app": envTesterName,
	}
	_, err := f.k8sInterface.AppsV1().Deployments(f.namespace).Create(f.envTesterDeployment(labels, testedEnv))
	if err != nil {
		return err
	}
	_, err = f.k8sInterface.CoreV1().Services(f.namespace).Create(f.envTesterService(labels))
	if err != nil {
		return err
	}
	return nil
}

func (f *brokersFlow) envTesterService(labels map[string]string) *k8sCoreTypes.Service {
	return &k8sCoreTypes.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      envTesterName,
			Namespace: f.namespace,
		},
		Spec: k8sCoreTypes.ServiceSpec{
			Selector: labels,
			Ports: []k8sCoreTypes.ServicePort{
				{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 8080},
				},
			},
			Type: k8sCoreTypes.ServiceTypeNodePort,
		},
	}
}

// envTesterDeployment creates a deployment which starts Alpine Pod, which prints value of the environment variable testedEnv
func (f *brokersFlow) envTesterDeployment(labels map[string]string, testedEnv string) *appsTypes.Deployment {
	var replicas int32 = 1

	return &appsTypes.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      envTesterName,
			Namespace: f.namespace,
		},
		Spec: appsTypes.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Replicas: &replicas,
			Template: k8sCoreTypes.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: k8sCoreTypes.PodSpec{
					Containers: []k8sCoreTypes.Container{
						{
							Name:    "app",
							Image:   "alpine:3.8",
							Command: []string{"/bin/sh", "-c", "--"},
							Args:    []string{fmt.Sprintf("echo \"value=$%s\"; echo \"done\"; while true; do sleep 30; done;", testedEnv)},
						},
					},
				},
			},
		},
	}
}

func (f *brokersFlow) createBindingAndWaitForReadiness(bindingName, instanceName string) error {
	f.log.Infof("Creating binding %s for service instance %s", bindingName, instanceName)
	_, err := f.scInterface.ServicecatalogV1beta1().ServiceBindings(f.namespace).Create(&v1beta1.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBinding",
			APIVersion: "servicecatalog.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: bindingName,
		},
		Spec: v1beta1.ServiceBindingSpec{
			InstanceRef: v1beta1.LocalObjectReference{
				Name: instanceName,
			},
		},
	})
	if err != nil {
		return err
	}

	f.log.Infof("Waiting for binding %s to be ready", bindingName)
	return f.wait(time.Minute, func() (done bool, err error) {
		sb, err := f.scInterface.ServicecatalogV1beta1().ServiceBindings(f.namespace).Get(bindingName, metav1.GetOptions{})
		if err != nil {
			f.log.Error(err.Error())
			return false, err
		}
		for _, cnd := range sb.Status.Conditions {
			if cnd.Type == v1beta1.ServiceBindingConditionReady && cnd.Status == v1beta1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
}
func (f *brokersFlow) createBindingUsage(usageName, bindingName string) error {
	f.log.Infof("Creating %s binding usage", usageName)
	_, err := f.buInterface.ServicecatalogV1alpha1().ServiceBindingUsages(f.namespace).Create(&v1alpha1.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingUsage",
			APIVersion: "servicecatalog.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: usageName,
		},
		Spec: v1alpha1.ServiceBindingUsageSpec{
			ServiceBindingRef: v1alpha1.LocalReferenceByName{
				Name: bindingName,
			},
			UsedBy: v1alpha1.LocalReferenceByKindAndName{
				Kind: "deployment",
				Name: envTesterName,
			},
		},
	})
	return err
}

func (f *brokersFlow) createBindingUsageAndWaitForReadiness(usageName, bindingName string) error {
	if err := f.createBindingUsage(usageName, bindingName); err != nil {
		return err
	}

	f.log.Infof("Waiting for binding usage %s to be ready", usageName)
	return f.wait(time.Minute, func() (done bool, err error) {
		si, err := f.buInterface.ServicecatalogV1alpha1().ServiceBindingUsages(f.namespace).Get(usageName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		for _, cnd := range si.Status.Conditions {
			if cnd.Type == v1alpha1.ServiceBindingUsageReady && cnd.Status == v1alpha1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
}

func (f *brokersFlow) deleteBindingUsage(name string) error {
	f.log.Infof("Deleting ServiceBindingUsage %s", name)
	return f.buInterface.ServicecatalogV1alpha1().ServiceBindingUsages(f.namespace).Delete(name, &metav1.DeleteOptions{})
}

func (f *brokersFlow) wait(timeout time.Duration, conditionFunc wait.ConditionFunc) error {
	timeoutCh := time.After(timeout)
	stopCh := make(chan struct{})
	go func() {
		select {
		case <-timeoutCh:
			close(stopCh)
		}
	}()
	return wait.PollUntil(500*time.Millisecond, conditionFunc, stopCh)
}

func (f *brokersFlow) waitForEnvInjected(expectedEnvName, expectedEnvValue string) error {
	return f.waitForEnvTesterValue(expectedEnvName, expectedEnvValue)
}

func (f *brokersFlow) waitForEnvNotInjected(expectedEnvName string) error {
	return f.waitForEnvTesterValue(expectedEnvName, "")
}

func (f *brokersFlow) waitForEnvTesterValue(expectedEnvName, expectedEnvValue string) error {
	f.log.Infof("Waiting for env: %s value: %s", expectedEnvName, expectedEnvValue)
	return f.wait(3*time.Minute, func() (bool, error) {
		var podName string

		// wait for running single env tester pod
		err := f.wait(time.Minute, func() (bool, error) {
			pods, err := f.k8sInterface.CoreV1().Pods(f.namespace).List(metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", envTesterName),
			})
			if err != nil {
				return false, err
			}

			//extract running pods
			var runningPods []*k8sCoreTypes.Pod
			for _, p := range pods.Items {
				var containerRunning bool
				for _, c := range p.Status.ContainerStatuses {
					if c.Name == "app" && c.State.Running != nil {
						containerRunning = true
					}
				}
				if p.Status.Phase == k8sCoreTypes.PodRunning && containerRunning {
					runningPods = append(runningPods, &p)
				}
			}

			// expect only one running env-tester pod
			if len(runningPods) == 1 {
				podName = runningPods[0].Name
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			return false, nil
		}

		req := f.k8sInterface.CoreV1().Pods(f.namespace).GetLogs(podName, &k8sCoreTypes.PodLogOptions{Container: "app"})
		readCloser, err := req.Stream()
		if err != nil {
			// it can happen, when pod is initializing
			return false, nil
		}
		defer readCloser.Close()
		logs, err := ioutil.ReadAll(readCloser)
		if err != nil {
			f.log.Warnf("error while reading logs: %s", err.Error())
			return false, nil
		}
		if strings.Contains(string(logs), fmt.Sprintf("value=%s", expectedEnvValue)) {
			return true, nil
		}
		// the "done" string is sent just after the value, it means the value was printed
		if strings.Contains(string(logs), "done") {
			f.log.Errorf("unexpected environment variable value: %s", string(logs))
			return false, nil
		}
		return false, nil
	})
}

func (f *brokersFlow) waitForInstance(name string) error {
	f.log.Infof("Waiting for %s instance to be ready", name)
	return f.wait(3*time.Minute, func() (done bool, err error) {
		si, err := f.scInterface.ServicecatalogV1beta1().ServiceInstances(f.namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if si.Status.ProvisionStatus == v1beta1.ServiceInstanceProvisionStatusProvisioned {
			return true, nil
		}
		return false, nil
	})
}

func (f *brokersFlow) waitForInstanceRemoved(name string) error {
	f.log.Infof("Waiting for %s instance to be removed", name)
	return f.wait(3*time.Minute, func() (bool, error) {
		_, err := f.scInterface.ServicecatalogV1beta1().ServiceInstances(f.namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			f.log.Warnf(err.Error())
			return false, err
		}
		return false, nil
	})
}

func (f *brokersFlow) logK8SReport() {
	deployments, err := f.k8sInterface.AppsV1().Deployments(f.namespace).List(metav1.ListOptions{})
	f.report("Deployments", deployments, err)

	pods, err := f.k8sInterface.CoreV1().Pods(f.namespace).List(metav1.ListOptions{})
	f.report("Pods", pods, err)

	secrets, err := f.k8sInterface.CoreV1().Secrets(f.namespace).List(metav1.ListOptions{})
	f.report("Secrets", secrets, err)
}

func (f *brokersFlow) logServiceCatalogAndBindingUsageReport() {
	clusterServiceBrokers, err := f.scInterface.ServicecatalogV1beta1().ClusterServiceBrokers().List(metav1.ListOptions{})
	f.report("ClusterServiceBrokers", clusterServiceBrokers, err)

	serviceBrokers, err := f.scInterface.ServicecatalogV1beta1().ServiceBrokers(f.namespace).List(metav1.ListOptions{})
	f.report("ServiceBrokers", serviceBrokers, err)

	clusterServiceClass, err := f.scInterface.ServicecatalogV1beta1().ClusterServiceClasses().List(metav1.ListOptions{})
	f.report("ClusterServiceClasses", clusterServiceClass, err)

	serviceClass, err := f.scInterface.ServicecatalogV1beta1().ServiceClasses(f.namespace).List(metav1.ListOptions{})
	f.report("ServiceClasses", serviceClass, err)

	serviceInstances, err := f.scInterface.ServicecatalogV1beta1().ServiceInstances(f.namespace).List(metav1.ListOptions{})
	f.report("ServiceInstances", serviceInstances, err)

	serviceBindings, err := f.scInterface.ServicecatalogV1beta1().ServiceBindings(f.namespace).List(metav1.ListOptions{})
	f.report("ServiceBindings", serviceBindings, err)

	serviceBindingUsages, err := f.buInterface.ServicecatalogV1alpha1().ServiceBindingUsages(f.namespace).List(metav1.ListOptions{})
	f.report("ServiceBindingUsages", serviceBindingUsages, err)
}

func (f *brokersFlow) report(kind string, obj runtime.Object, err error) {
	printer := &printers.JSONPrinter{}
	logger := &logWriter{log: f.log}
	obj.(printerObject).SetGroupVersionKind(schema.GroupVersionKind{Kind: kind})
	if err != nil {
		f.log.Errorf("Could not fetch resources: %v", err)
		return
	}
	err = printer.PrintObj(obj, logger)
	if err != nil {
		f.log.Errorf("Could not print objects: %v", err)
	}
}

type logWriter struct {
	log logrus.FieldLogger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.log.Infof(string(p))
	return len(p), nil
}

type printerObject interface {
	SetGroupVersionKind(gvk schema.GroupVersionKind)
}
