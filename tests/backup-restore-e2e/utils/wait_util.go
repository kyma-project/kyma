package utils

import (
	"fmt"
	"time"

	arkv1 "github.com/heptio/ark/pkg/apis/ark/v1"
	arkclient "github.com/heptio/ark/pkg/generated/clientset/versioned"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	scClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	sbuapi "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	sbuClient "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionFunc - type for func returns (bool, error)
type ConditionFunc func() (bool, error)

// Retry function run provided function maxRetries times with interval between rerun
func Retry(interval time.Duration, maxRetries int, f ConditionFunc) error {
	if maxRetries <= 0 {
		return fmt.Errorf("maxRetries (%d) should be > 0", maxRetries)
	}
	tick := time.NewTicker(interval)
	defer tick.Stop()

	for i := 0; ; i++ {
		ok, err := f()
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		if i == maxRetries {
			break
		}
		<-tick.C
	}
	return fmt.Errorf("Still failed after %d retries", maxRetries)
}

// WaitServiceBrokerReady waits until service broker is in ready state
func WaitServiceBrokerReady(sbClient clientset.Interface, namespace string, name string) error {
	err := Retry(5*time.Second, 24, func() (bool, error) {
		sb, _ := sbClient.ServicecatalogV1beta1().ServiceBrokers(namespace).Get(name, metav1.GetOptions{})

		for _, s := range sb.Status.Conditions {
			if s.Type == scv1beta1.ServiceBrokerConditionReady && s.Status == scv1beta1.ConditionTrue {
				return true, nil
			}
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("failed to wait for servicebroker %s in namespace %s to become ready: %v", namespace, name, err)
	}

	return nil
}

// WaitForServiceInstanceReady waits until service instance is in ready state
func WaitForServiceInstanceReady(sbClient clientset.Interface, namespace string, name string) error {
	err := Retry(5*time.Second, 24, func() (bool, error) {
		si, _ := sbClient.ServicecatalogV1beta1().ServiceInstances(namespace).Get(name, metav1.GetOptions{})

		for _, s := range si.Status.Conditions {
			if s.Type == scv1beta1.ServiceInstanceConditionReady && s.Status == scv1beta1.ConditionTrue {
				return true, nil
			} else if s.Type == scv1beta1.ServiceInstanceConditionReady && s.Status == scv1beta1.ConditionFalse && s.Reason != "Provisioning" {
				return false, fmt.Errorf("Service instance %s failed: %s", name, s.Message)
			}
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("failed to wait for serviceinstance %s in namespace %s to become ready: %v", name, namespace, err)
	}

	return nil
}

// WaitForServiceBindingUsageReady waits until service binding usage is in ready state
func WaitForServiceBindingUsageReady(sbuClient sbuClient.Interface, namespace string, name string) error {
	err := Retry(5*time.Second, 24, func() (bool, error) {
		sbu, _ := sbuClient.ServicecatalogV1alpha1().ServiceBindingUsages(namespace).Get(name, metav1.GetOptions{})

		for _, s := range sbu.Status.Conditions {
			if s.Type == sbuapi.ServiceBindingUsageReady && s.Status == sbuapi.ConditionTrue {
				return true, nil
			}
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("failed to wait for bindingusage %s in namespace %s to become ready: %v", namespace, name, err)
	}

	return nil
}

// WaitForServiceBindingReady waits until service binding is in ready state
func WaitForServiceBindingReady(sbClient clientset.Interface, namespace string, name string) error {
	err := Retry(5*time.Second, 24, func() (bool, error) {
		if st, _ := getServiceBindingConditionReady(sbClient, namespace, name); st == scv1beta1.ConditionTrue {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		_, msg := getServiceBindingConditionReady(sbClient, namespace, name)
		return fmt.Errorf("failed to wait for servicebinding %s in namespace %s to become ready: %v. StatusMessage: %s", namespace, name, err, msg)
	}

	return nil
}

func getServiceBindingConditionReady(sbClient clientset.Interface, namespace string, name string) (status scv1beta1.ConditionStatus, message string) {
	si, _ := sbClient.ServicecatalogV1beta1().ServiceBindings(namespace).Get(name, metav1.GetOptions{})

	for _, s := range si.Status.Conditions {
		if s.Type == scv1beta1.ServiceBindingConditionReady {
			return s.Status, s.Message
		}
	}
	return "", ""
}

// WaitForBackupCompleted waits until heptio ark backup is completed
func WaitForBackupCompleted(arkCli arkclient.Interface, namespace string, name string) error {
	err := Retry(5*time.Second, 24, func() (bool, error) {

		b, _ := arkCli.ArkV1().Backups(namespace).Get(name, metav1.GetOptions{})

		if b.Status.Phase == arkv1.BackupPhaseCompleted {
			return true, nil
		} else if b.Status.Phase == arkv1.BackupPhaseFailed || b.Status.Phase == arkv1.BackupPhaseFailedValidation {
			return false, fmt.Errorf("Backup has failed with status: %s", b.Status.Phase)
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("failed to wait for backup %s to complete: %v", name, err)
	}

	return nil
}

// WaitForRestoreCompleted waits until heptio ark restore is completed
func WaitForRestoreCompleted(arkCli arkclient.Interface, namespace string, name string) error {
	err := Retry(5*time.Second, 24, func() (bool, error) {

		b, _ := arkCli.ArkV1().Restores(namespace).Get(name, metav1.GetOptions{})

		if b.Status.Phase == arkv1.RestorePhaseCompleted {
			if b.Status.Errors > 0 {
				//return false, fmt.Errorf("Restore completed with %d errors", b.Status.Errors)
			}
			return true, nil
		} else if b.Status.Phase == arkv1.RestorePhaseFailedValidation {
			return false, fmt.Errorf("Restore has failed validation with errors: %s. ", b.Status.ValidationErrors)
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("failed to wait for restore %s to complete: %v", name, err)
	}

	return nil
}

// WaitForAllInstancesDeleted waits until all service instances are deleted
func WaitForAllInstancesDeleted(cli scClient.Interface, namespace string) error {
	err := Retry(5*time.Second, 24, func() (bool, error) {
		sil, err := cli.ServicecatalogV1beta1().ServiceInstances(namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		if len(sil.Items) == 0 {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("failed to wait for all service instances are deleted: %v", err)
	}

	return nil
}

// WaitForAllBindingsDeleted waits until all service bindings are deleted
func WaitForAllBindingsDeleted(cli scClient.Interface, namespace string) error {
	err := Retry(5*time.Second, 24, func() (bool, error) {
		sbl, err := cli.ServicecatalogV1beta1().ServiceBindings(namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		if len(sbl.Items) == 0 {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("failed to wait for all service bindings are deleted: %v", err)
	}

	return nil
}
