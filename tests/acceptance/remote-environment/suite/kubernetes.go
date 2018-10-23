package suite

import (
	"fmt"
	"io/ioutil"
	"sort"
	"time"

	"github.com/kyma-project/kyma/tests/acceptance/pkg/repeat"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	authV1 "k8s.io/api/rbac/v1beta1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

const (
	containerPort = 8080

	envInjectedKey   = "envInjected"
	callSucceededKey = "callSucceeded"
	callForbiddenKey = "callForbidden"
)

func (ts *TestSuite) createKubernetesResources() {
	gwSelectorLabels := map[string]string{
		// label app value must match container name - used, when printing logs
		"app": "fake-gateway",
		"acceptance-test-app": "fake-gateway",
	}

	clientset, err := kubernetes.NewForConfig(ts.config)
	require.NoError(ts.t, err)

	nsClient := clientset.CoreV1().Namespaces()
	_, err = nsClient.Create(fixEnvironment(ts.namespace))
	require.NoError(ts.t, err)

	deploymentClient := clientset.AppsV1beta1().Deployments(ts.namespace)
	serviceClient := clientset.CoreV1().Services(ts.namespace)
	cfgMapClient := clientset.CoreV1().ConfigMaps(ts.namespace)

	_, err = clientset.CoreV1().ServiceAccounts(ts.namespace).Create(fixServiceAccount())
	require.NoError(ts.t, err)

	_, err = clientset.RbacV1beta1().Roles(ts.namespace).Create(fixRole())
	require.NoError(ts.t, err)

	_, err = clientset.RbacV1beta1().RoleBindings(ts.namespace).Create(fixRoleBinding(ts.namespace))
	require.NoError(ts.t, err)

	_, err = cfgMapClient.Create(fixConfigMap())
	require.NoError(ts.t, err)

	_, err = deploymentClient.Create(fixGatewayDeployment(gwSelectorLabels, ts.remoteEnvironmentName, ts.dockerImage))
	require.NoError(ts.t, err)

	_, err = deploymentClient.Create(fixGatewayClientDeployment(ts.dockerImage, ts.gwClientSvcDeploymentName, ts.gatewaySvcName))
	require.NoError(ts.t, err)

	_, err = serviceClient.Create(fixService(ts.gatewaySvcName, gwSelectorLabels, map[string]string{}))
	require.NoError(ts.t, err)
}

func fixRoleBinding(ns string) *authV1.RoleBinding {
	return &authV1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "acceptance-test",
		},
		Subjects: []authV1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "acceptance-test",
				Namespace: ns,
			},
		},
		RoleRef: authV1.RoleRef{
			Kind:     "Role",
			Name:     "acceptance-test",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func fixRole() *authV1.Role {
	return &authV1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: "acceptance-test",
		},
		Rules: []authV1.PolicyRule{
			{
				APIGroups: []string{""},
				Verbs:     []string{"create", "get", "update"},
				Resources: []string{"configmaps"},
			},
		},
	}
}
func fixServiceAccount() *apiv1.ServiceAccount {
	return &apiv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "acceptance-test",
		},
	}
}

func (ts *TestSuite) ensureNamespaceIsDeleted(timeout time.Duration) {
	clientset, err := kubernetes.NewForConfig(ts.config)
	require.NoError(ts.t, err)

	nsClient := clientset.CoreV1().Namespaces()

	err = nsClient.Delete(ts.namespace, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)

	waitForNsTermination := func() error {
		ns, err := nsClient.Get(ts.namespace, metav1.GetOptions{})
		switch {
		case err == nil:
			return fmt.Errorf("namespace %q still exists [phase: %q]", ts.namespace, ns.Status.Phase)
		case apiErrors.IsNotFound(err):
			return nil
		default:
			return errors.Wrap(err, "while getting namespace")
		}
	}

	repeat.FuncAtMost(ts.t, waitForNsTermination, timeout)
}

func (ts *TestSuite) deleteNamespace() {
	clientset, err := kubernetes.NewForConfig(ts.config)
	require.NoError(ts.t, err)

	nsClient := clientset.CoreV1().Namespaces()
	err = nsClient.Delete(ts.namespace, &metav1.DeleteOptions{})
	require.NoError(ts.t, err)
}

func (ts *TestSuite) WaitForPodsAreRunning(timeout time.Duration) {
	clientSet, err := kubernetes.NewForConfig(ts.config)
	require.NoError(ts.t, err)
	done := time.After(timeout)
	for {
		pods, err := clientSet.CoreV1().Pods(ts.namespace).List(metav1.ListOptions{})
		require.NoError(ts.t, err)

		// check, if pods are ready
		allPodsReady := true
		for _, pod := range pods.Items {
			if !ts.isPodReady(&pod) {
				allPodsReady = false
				break
			}
		}
		if allPodsReady {
			return
		}

		select {
		case <-done:
			require.Fail(ts.t, "Timeout for pods running exceeded.", ts.notReadyPodsReport(clientSet, pods.Items))
			return
		default:
			time.Sleep(time.Second)
		}
	}
}

func (ts *TestSuite) isPodReady(pod *apiv1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if !cs.Ready {
			return false
		}
	}
	return true
}

type eventsByCreationTimestamp []apiv1.Event

func (e eventsByCreationTimestamp) Len() int {
	return len(e)
}

func (e eventsByCreationTimestamp) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e eventsByCreationTimestamp) Less(i, j int) bool {
	return e[i].CreationTimestamp.Before(&e[j].CreationTimestamp)
}

func (ts *TestSuite) notReadyPodsReport(cs *kubernetes.Clientset, pods []apiv1.Pod) string {
	eventInterface := cs.CoreV1().Events(ts.namespace)
	report := ""
	for _, pod := range pods {
		if !ts.isPodReady(&pod) {
			pReport := fmt.Sprintf("Pod %s, phase %s:\n", pod.Name, pod.Status.Phase)

			selector := eventInterface.GetFieldSelector(nil, &ts.namespace, nil, str2Ptr(string(pod.UID)))
			eventList, err := eventInterface.List(metav1.ListOptions{
				FieldSelector: selector.String(),
			})
			require.NoError(ts.t, err)
			events := eventList.Items
			sort.Sort(eventsByCreationTimestamp(events))

			pReport = pReport + "  Events:\n"
			for _, event := range events {
				pReport = pReport + fmt.Sprintf("   - type: %s, reason: %s, message: %s\n", event.Type, event.Reason, event.Message)
			}

			switch pod.Status.Phase {
			case apiv1.PodRunning, apiv1.PodSucceeded, apiv1.PodFailed:
				pReport = pReport + "\n" + ts.podLogs(cs, &pod, pod.Labels["app"]) + "\n"
			}

			report = report + pReport
		}
	}
	return report
}

func (ts *TestSuite) printGatewayClientLogs() {
	// in case of errors, do not fail test, just log error. The method is executed when we know the test is failing.
	clientset, err := kubernetes.NewForConfig(ts.config)
	if err != nil {
		ts.t.Logf("error while creating client set: %s", err.Error())
		return
	}

	pods, err := clientset.CoreV1().Pods(ts.namespace).List(metav1.ListOptions{
		LabelSelector: "app=gateway-client",
	})
	if err != nil {
		ts.t.Logf("error while getting gateway-client pod: %s", err.Error())
		return
	}

	for _, pod := range pods.Items {
		ts.t.Log(ts.podLogs(clientset, &pod, "gateway-client"))
	}
}

func (ts *TestSuite) podLogs(clientset *kubernetes.Clientset, pod *apiv1.Pod, containerName string) string {
	req := clientset.CoreV1().Pods(ts.namespace).GetLogs(pod.Name, &apiv1.PodLogOptions{
		Container: containerName,
	})

	readCloser, err := req.Stream()
	if err != nil {
		ts.t.Logf("error while getting log stream: %s", err.Error())
		return ""
	}
	defer readCloser.Close()

	logs, err := ioutil.ReadAll(readCloser)
	if err != nil {
		ts.t.Logf("error while reading logs from pod %s, error: %s", pod.Name, err.Error())
		return ""
	}
	return fmt.Sprintf("Logs from pod %s (created at %v):\n%s", pod.Name, pod.CreationTimestamp.Format("15:04:05"), string(logs))
}

func fixEnvironment(name string) *apiv1.Namespace {
	return &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"env": "true", "istio-injection": "enabled"},
		},
	}
}

func fixNamespace(name string) *apiv1.Namespace {
	return &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func fixService(serviceName string, selectorLabels, annotations map[string]string) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        serviceName,
			Annotations: annotations,
		},
		Spec: apiv1.ServiceSpec{
			Selector: selectorLabels,
			Ports: []apiv1.ServicePort{
				{
					Name:       "http",
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: containerPort},
				},
			},
		},
	}
}

func fixGatewayDeployment(labels map[string]string, remoteEnvironmentName, image string) *appsv1beta1.Deployment {
	return &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fake-gateway",
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "fake-gateway",
							Image: image,

							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: containerPort,
								},
							},
							Env: []apiv1.EnvVar{
								{Name: "REMOTE_ENVIRONMENT_NAME", Value: remoteEnvironmentName},
							},
							Command: []string{"/go/bin/gateway.bin"},
						},
					},
				},
			},
		},
	}
}

func fixConfigMap() *apiv1.ConfigMap {
	return &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-output",
		},
		Data: map[string]string{
			envInjectedKey:   "not-defined",
			callSucceededKey: "not-defined",
			callForbiddenKey: "not-defined",
		},
	}
}

func fixGatewayClientDeployment(image, deploymentName, gatewaySvcName string) *appsv1beta1.Deployment {
	return &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						// label app value must match container name - used, when printing logs
						"app": "gateway-client",
					},
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "true",
					},
				},
				Spec: apiv1.PodSpec{
					ServiceAccountName:            "acceptance-test",
					TerminationGracePeriodSeconds: int64Ptr(5),
					Containers: []apiv1.Container{
						{
							Name:    "gateway-client",
							Image:   image,
							Command: []string{"/go/bin/client.bin"},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: containerPort,
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "TARGET_URL",
									Value: fmt.Sprintf("http://%s", gatewaySvcName),
								},
								{
									Name: "NAMESPACE",
									ValueFrom: &apiv1.EnvVarSource{
										FieldRef: &apiv1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}

func str2Ptr(s string) *string {
	return &s
}
