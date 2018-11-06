package main

import (
	"fmt"
	"strings"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/kyma/tests/backup-restore-e2e/consts"
	"github.com/kyma-project/kyma/tests/backup-restore-e2e/framework"
	"github.com/kyma-project/kyma/tests/backup-restore-e2e/utils"
)

// TestBackupCreate is the main test for backup phase
func TestBackupCreate(t *testing.T) {

	logrus.Info("Starting backup test")
	f := framework.Global

	logrus.Infof("Creating environment namespace: %s", f.Namespace)
	n := utils.NewEnvironmentNamespace(f.Namespace)
	_, err := utils.CreateNamespace(t, n, f.KubeClient)
	if err != nil {
		t.Fatal(err)
	}

	fname := fmt.Sprintf("function-%s", f.Namespace)
	logrus.Infof("Creating function %s", fname)
	fn := utils.NewFunction(fname, f.Namespace)
	_, err = utils.CreateFunction(t, fn, f.KubelessClient)
	if err != nil {
		t.Fatal(err)
	}

	logrus.Infof("Tests will be run for those brokers: %s", strings.Join(f.Brokers, ","))

	// REMOTE ENVIRONMENT BROKER
	if framework.IsBrokerSelected(consts.BrokerReb) {
		envServiceName := fmt.Sprintf("promotions-%s", uuid.NewV4().String()[0:7])
		logrus.Infof("Creating remote environment, envServiceName: %s", envServiceName)
		re := utils.NewRemoteEnvironment(f.Namespace, envServiceName)
		_, err = utils.CreateRemoteEnvironment(t, re, f.REClient)
		if err != nil {
			t.Fatal(err)
		}

		logrus.Info("Creating environment mapping")
		em := utils.NewEnvironmentMapping(f.Namespace)
		_, err = utils.CreateEnvironmentMapping(t, em, f.REClient)
		if err != nil {
			t.Fatal(err)
		}

		logrus.Infof("Waiting for service broker %s to be ready", consts.RemoteEnvBrokerName)
		if err := utils.WaitServiceBrokerReady(f.SBClient, f.Namespace, consts.RemoteEnvBrokerName); err != nil {
			t.Fatalf("failed to wait for service broker %s: %v", consts.RemoteEnvBrokerName, err)
		}

		instanceName := fmt.Sprintf("promotions-%s", uuid.NewV4().String()[0:7])
		className := envServiceName
		planName := "default"

		createServiceInstanceAndBinding(t, f, false, nil, className, planName, instanceName, fname, "reb")
	}

	// HELM BROKER
	if framework.IsBrokerSelected(consts.BrokerHelm) {
		createServiceInstanceAndBinding(t, f, true, nil, "redis", "micro", "redis-helm", fname, "helmbroker")
	}

	// GCP BROKER
	if framework.IsBrokerSelected(consts.BrokerGcp) {
		createServiceInstanceAndBinding(t, f, true, nil, "gcp-example", "gcp-example", "gcp-example", fname, "gcpbroker")

	}

	// AZURE BROKER
	if framework.IsBrokerSelected(consts.BrokerAzure) {
		instanceParams := &runtime.RawExtension{
			Raw: []byte(`{ "location": "westeurope", "resourceGroup":"ark-test" }`),
		}
		createServiceInstanceAndBinding(t, f, true, instanceParams, "azure-storage", "blob-storage-account", "ark-test-azure-sa", fname, "azurebroker")
	}

	// BACKUP
	backupName := fmt.Sprintf("%s-ns", f.Namespace)
	logrus.Infof("Creating backup %s", backupName)
	backup := utils.NewBackup(backupName, []string{f.Namespace, consts.IntegrationNamespace}, nil)
	_, err = utils.CreateBackup(t, backup, f.ArkClient)
	if err != nil {
		t.Fatal(err)
	}

	logrus.Infof("Waiting for backup %s to complete", backupName)
	if err := utils.WaitForBackupCompleted(f.ArkClient, consts.HeptioArkNamespace, backupName); err != nil {
		t.Fatalf("Failed to wait for backup to complete: %v", err)
	}

	backupName = fmt.Sprintf("%s-re", f.Namespace)
	logrus.Infof("Creating backup %s", backupName)
	backup = utils.NewBackup(backupName, nil, []string{"remoteenvironment"})
	_, err = utils.CreateBackup(t, backup, f.ArkClient)
	if err != nil {
		t.Fatal(err)
	}

	logrus.Infof("Waiting for backup %s to complete", backupName)
	if err := utils.WaitForBackupCompleted(f.ArkClient, consts.HeptioArkNamespace, backupName); err != nil {
		t.Fatalf("Failed to wait for backup to complete: %v", err)
	}
}

func createServiceInstanceAndBinding(t *testing.T, f *framework.Framework, isClusterScoped bool, params *runtime.RawExtension, className string, planName string, instanceName string, fname string, brokerName string) error {
	logrus.Infof("Creating service instance for %s", brokerName)
	hbsi := utils.NewServiceInstance(isClusterScoped, params, f.Namespace, className, planName, instanceName)
	_, err := utils.CreateServiceInstance(t, hbsi, f.SBClient)
	if err != nil {
		t.Fatal(err)
	}

	logrus.Info("Waiting for service instance to be ready")
	if err := utils.WaitForServiceInstanceReady(f.SBClient, f.Namespace, instanceName); err != nil {
		t.Fatalf("failed to wait for service instance: %v", err)
	}

	logrus.Info("Creating service binding")
	sbName := fmt.Sprintf("%s-%s", brokerName, uuid.NewV4().String()[0:7])
	sb := utils.NewServiceBinding(sbName, f.Namespace, instanceName)
	_, err = utils.CreateServiceBinding(t, sb, f.SBClient)
	if err != nil {
		t.Fatal(err)
	}

	logrus.Info("Creating service binding usage")
	sbuName := fmt.Sprintf("%s-%s", brokerName, uuid.NewV4().String()[0:7])
	sbu := utils.NewServiceBindingUsage(sbuName, f.Namespace, sbName, "function", fname)
	_, err = utils.CreateServiceBindingUsage(t, sbu, f.SbuClient)
	if err != nil {
		t.Fatal(err)
	}

	logrus.Info("Waiting for servicebindingusage to be ready")
	if err := utils.WaitForServiceBindingUsageReady(f.SbuClient, f.Namespace, sbuName); err != nil {
		t.Fatalf("failed to wait for service binding usage: %v", err)
	}

	return nil
}
