package main

import (
	"fmt"
	"sync"
	"testing"

	arkclient "github.com/heptio/ark/pkg/generated/clientset/versioned"
	scClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/backup-restore-e2e/consts"
	"github.com/kyma-project/kyma/tests/backup-restore-e2e/utils"

	"github.com/kyma-project/kyma/tests/backup-restore-e2e/framework"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestBackupCreate is the main test for backup phase
func TestRestoreCreate(t *testing.T) {
	logrus.Info("Starting restore test")
	f := framework.Global

	createAndWaitForRestore(t, f.ArkClient, fmt.Sprintf("%s-re", f.Namespace))
	createAndWaitForRestore(t, f.ArkClient, fmt.Sprintf("%s-ns", f.Namespace))

	// Let's check everything...
	var siWg sync.WaitGroup
	serviceInstances, _ := f.SBClient.ServicecatalogV1beta1().ServiceInstances(f.Namespace).List(metav1.ListOptions{})
	siWg.Add(len(serviceInstances.Items))
	for _, si := range serviceInstances.Items {
		go func(sbClient scClient.Interface, namespace string, name string) {
			defer siWg.Done()
			if err := utils.WaitForServiceInstanceReady(sbClient, namespace, name); err != nil {
				t.Fail()
				t.Logf("ServiceInstance: %s is not ready with error message: %v", name, err)
			}
		}(f.SBClient, si.Namespace, si.Name)
	}

	var sbWg sync.WaitGroup
	serviceBindings, _ := f.SBClient.ServicecatalogV1beta1().ServiceBindings(f.Namespace).List(metav1.ListOptions{})
	sbWg.Add(len(serviceBindings.Items))
	for _, sb := range serviceBindings.Items {
		go func(sbClient scClient.Interface, namespace string, name string) {
			defer sbWg.Done()
			if err := utils.WaitForServiceBindingReady(sbClient, namespace, name); err != nil {
				t.Fail()
				t.Logf("ServiceBinding: %s is not ready with error message: %v", name, err)
			}
		}(f.SBClient, sb.Namespace, sb.Name)
	}

	logrus.Info("Waiting for service instances and bindings to become ready...")
	siWg.Wait()
	sbWg.Wait()

}

func createAndWaitForRestore(t *testing.T, arkCli arkclient.Interface, restoreName string) {
	logrus.Infof("Creating restore %s", restoreName)
	r := utils.NewRestore(restoreName, restoreName)
	_, err := utils.CreateRestore(t, r, arkCli)
	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}

	logrus.Info("Waiting for restore to complete")
	if err := utils.WaitForRestoreCompleted(arkCli, consts.HeptioArkNamespace, restoreName); err != nil {
		t.Fatalf("Failed to wait for restore to complete: %v", err)
	}
}
