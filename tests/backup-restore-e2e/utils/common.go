package utils

import (
	"fmt"
	"io/ioutil"
	"testing"

	arkv1 "github.com/heptio/ark/pkg/apis/ark/v1"
	arkclient "github.com/heptio/ark/pkg/generated/clientset/versioned"
	kubelessapi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	kubelesscli "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	sbuapi "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	sbuClient "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateNamespace creates a namespace
func CreateNamespace(t *testing.T, namespace *v1.Namespace, cli kubernetes.Interface) (*v1.Namespace, error) {
	res, err := cli.CoreV1().Namespaces().Create(namespace)

	if err != nil {
		return nil, err
	}
	t.Logf("Created namespace: %s", res.Name)

	return res, nil
}

// CreateRemoteEnvironment creates a RemoteEnvironment
func CreateRemoteEnvironment(t *testing.T, re *v1alpha1.RemoteEnvironment, cli versioned.Interface) (*v1alpha1.RemoteEnvironment, error) {
	res, err := cli.ApplicationconnectorV1alpha1().RemoteEnvironments().Create(re)
	if err != nil {
		return nil, err
	}
	t.Logf("Created remote environment: %s", res.Name)

	return res, nil
}

// CreateEnvironmentMapping creates an EnvironmentMapping
func CreateEnvironmentMapping(t *testing.T, em *v1alpha1.EnvironmentMapping, cli versioned.Interface) (*v1alpha1.EnvironmentMapping, error) {
	res, err := cli.ApplicationconnectorV1alpha1().EnvironmentMappings(em.Namespace).Create(em)
	if err != nil {
		return nil, err
	}
	t.Logf("Created environment mapping: %s", res.Name)

	return res, nil
}

// CreateServiceInstance creates a ServiceInstance
func CreateServiceInstance(t *testing.T, si *scv1beta1.ServiceInstance, cli scClient.Interface) (*scv1beta1.ServiceInstance, error) {
	res, err := cli.ServicecatalogV1beta1().ServiceInstances(si.Namespace).Create(si)
	if err != nil {
		return nil, err
	}
	t.Logf("Created service instance: %s", res.Name)

	return res, nil
}

// CreateServiceBinding creates a ServiceBinding
func CreateServiceBinding(t *testing.T, sb *scv1beta1.ServiceBinding, cli scClient.Interface) (*scv1beta1.ServiceBinding, error) {
	res, err := cli.ServicecatalogV1beta1().ServiceBindings(sb.Namespace).Create(sb)
	if err != nil {
		return nil, err
	}
	t.Logf("Created service binding: %s", res.Name)

	return res, nil
}

// CreateBackup creates a Backup
func CreateBackup(t *testing.T, b *arkv1.Backup, cli arkclient.Interface) (*arkv1.Backup, error) {
	res, err := cli.ArkV1().Backups(b.Namespace).Create(b)
	if err != nil {
		return nil, err
	}
	t.Logf("Created backup: %s", res.Name)

	return res, nil
}

// CreateFunction creates a Function
func CreateFunction(t *testing.T, f *kubelessapi.Function, cli kubelesscli.Interface) (*kubelessapi.Function, error) {
	res, err := cli.KubelessV1beta1().Functions(f.Namespace).Create(f)
	if err != nil {
		return nil, err
	}
	t.Logf("Created function: %s", res.Name)

	return res, nil
}

// CreateServiceBindingUsage creates a ServiceBindingUsage
func CreateServiceBindingUsage(t *testing.T, sbu *sbuapi.ServiceBindingUsage, cli sbuClient.Interface) (*sbuapi.ServiceBindingUsage, error) {
	res, err := cli.ServicecatalogV1alpha1().ServiceBindingUsages(sbu.Namespace).Create(sbu)
	if err != nil {
		return nil, err
	}
	t.Logf("Created service binding usage: %s", res.Name)

	return res, nil
}

// CreateRestore creates a Restore
func CreateRestore(t *testing.T, r *arkv1.Restore, cli arkclient.Interface) (*arkv1.Restore, error) {
	res, err := cli.ArkV1().Restores(r.Namespace).Create(r)
	if err != nil {
		return nil, err
	}
	t.Logf("Created restore: %s", res.Name)

	return res, nil
}

// DeleteAllServiceBindings deletes all Service Bindings from provided namespace
func DeleteAllServiceBindings(namespace string, cli scClient.Interface) error {
	sbl, err := cli.ServicecatalogV1beta1().ServiceBindings(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, sb := range sbl.Items {
		cli.ServicecatalogV1beta1().ServiceBindings(namespace).Delete(sb.Name, &metav1.DeleteOptions{})
	}

	return nil
}

// DeleteAllServiceInstances deletes all Service Instances from provided namespace
func DeleteAllServiceInstances(namespace string, cli scClient.Interface) error {
	sil, err := cli.ServicecatalogV1beta1().ServiceInstances(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, si := range sil.Items {
		cli.ServicecatalogV1beta1().ServiceInstances(namespace).Delete(si.Name, &metav1.DeleteOptions{})
	}

	return nil
}

func PodLogs(t *testing.T, cli kubernetes.Interface, podName string, podNamespace string, containerName string) string {
	plo := &v1.PodLogOptions{}
	if containerName != "" {
		plo.Container = containerName
	}
	req := cli.CoreV1().Pods(podNamespace).GetLogs(podName, plo)

	readCloser, err := req.Stream()
	if err != nil {
		t.Logf("error while getting log stream: %s", err.Error())
		return ""
	}
	defer readCloser.Close()

	logs, err := ioutil.ReadAll(readCloser)
	if err != nil {
		t.Logf("error while reading logs from pod %s, error: %s", podName, err.Error())
		return ""
	}
	return fmt.Sprintf("Logs from pod %s:\n%s", podName, string(logs))
}
