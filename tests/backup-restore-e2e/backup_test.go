package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-project/kyma/tests/backup-restore-e2e/consts"
	"github.com/kyma-project/kyma/tests/backup-restore-e2e/framework"
	"github.com/kyma-project/kyma/tests/backup-restore-e2e/utils"
)

type brokerTestParams struct {
	isClusterScoped   bool
	instanceName      string
	className         string
	planName          string
	bindingUsedByType string
	bindingUsedByName string
	brokerType        string
	instanceParams    *runtime.RawExtension
	bindingsParams    *runtime.RawExtension
}

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

	logrus.Infof("Tests will be run for brokers: %s", strings.Join(f.Brokers, ","))

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

		rebBroker := brokerTestParams{
			isClusterScoped:   false,
			instanceName:      fmt.Sprintf("promotions-%s", uuid.NewV4().String()[0:7]),
			className:         envServiceName,
			planName:          "default",
			bindingUsedByType: "function",
			bindingUsedByName: fname,
			brokerType:        "reb",
		}

		rebBroker.createServiceInstanceAndBinding(t, f)
	}

	// HELM BROKER
	if framework.IsBrokerSelected(consts.BrokerHelm) {
		helmBroker := brokerTestParams{
			isClusterScoped:   true,
			instanceName:      "redis-helm",
			className:         "redis",
			planName:          "micro",
			bindingUsedByType: "function",
			bindingUsedByName: fname,
			brokerType:        "helmbroker",
		}
		helmBroker.createServiceInstanceAndBinding(t, f)
	}

	// GCP BROKER
	if framework.IsBrokerSelected(consts.BrokerGcp) {
		bucketName := fmt.Sprintf("ark-e2e-tests-%d", time.Now().Unix())

		gcpBroker := brokerTestParams{
			isClusterScoped:   true,
			instanceName:      "gcp-bucket",
			className:         "cloud-storage",
			planName:          "beta",
			bindingUsedByType: "function",
			bindingUsedByName: fname,
			brokerType:        "gcpbroker",
			instanceParams: &runtime.RawExtension{
				Raw: []byte(fmt.Sprintf(`{ "bucketId": "%s", "location":"EU" }`, bucketName)),
			},
			bindingsParams: &runtime.RawExtension{
				Raw: []byte(fmt.Sprintf(`{ "createServiceAccount": true, "roles":["roles/storage.objectAdmin"], "serviceAccount": "%s" }`, bucketName)),
			},
		}

		gcpBroker.createServiceInstanceAndBinding(t, f)
	}

	// AZURE BROKER
	if framework.IsBrokerSelected(consts.BrokerAzure) {

		azureBroker := brokerTestParams{
			isClusterScoped:   true,
			instanceName:      "ark-test-azure-sa",
			className:         "azure-storage",
			planName:          "blob-storage-account",
			bindingUsedByType: "function",
			bindingUsedByName: fname,
			brokerType:        "azurebroker",
			instanceParams: &runtime.RawExtension{
				Raw: []byte(`{ "location": "westeurope", "resourceGroup":"ark-test" }`),
			},
		}

		azureBroker.createServiceInstanceAndBinding(t, f)
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

func (b *brokerTestParams) createServiceInstanceAndBinding(t *testing.T, f *framework.Framework) error {
	logrus.Infof("Creating service instance for %s", b.brokerType)
	hbsi := utils.NewServiceInstance(b.isClusterScoped, b.instanceParams, f.Namespace, b.className, b.planName, b.instanceName)
	_, err := utils.CreateServiceInstance(t, hbsi, f.SBClient)
	if err != nil {
		t.Fatal(err)
	}

	logrus.Info("Waiting for service instance to be ready")
	if err := utils.WaitForServiceInstanceReady(f.SBClient, f.Namespace, b.instanceName); err != nil {
		t.Fatalf("failed to wait for service instance: %v", err)
	}

	logrus.Info("Creating service binding")
	sbName := fmt.Sprintf("%s-%s", b.brokerType, uuid.NewV4().String()[0:7])
	sb := utils.NewServiceBinding(sbName, f.Namespace, b.instanceName, b.bindingsParams)
	_, err = utils.CreateServiceBinding(t, sb, f.SBClient)
	if err != nil {
		t.Fatal(err)
	}

	logrus.Info("Creating service binding usage")
	sbuName := fmt.Sprintf("%s-%s", b.brokerType, uuid.NewV4().String()[0:7])
	sbu := utils.NewServiceBindingUsage(sbuName, f.Namespace, sbName, b.bindingUsedByType, b.bindingUsedByName)
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
