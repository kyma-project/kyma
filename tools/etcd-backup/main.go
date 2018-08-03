package main

import (
	"time"

	"github.com/kyma-project/kyma/tools/etcd-backup/internal/azure"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/backup"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/cleaner"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/platform/logger"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/platform/signal"

	etcdOperatorClientset "github.com/coreos/etcd-operator/pkg/generated/clientset/versioned"
	etcdOperatorInformers "github.com/coreos/etcd-operator/pkg/generated/informers/externalversions"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// informerResyncPeriod defines how often informer will execute relist action. Setting to zero disable resync.
// BEWARE: too short period time will increase the CPU load.
const informerResyncPeriod = time.Minute

// Config holds application configuration
type Config struct {
	Logger           logger.Config
	KubeconfigPath   string `envconfig:"optional"`
	WorkingNamespace string
	ABS              struct {
		ContainerName string
		SecretName    string
	}
	BlobPrefix string

	Backup  backup.Config
	Cleaner cleaner.Config
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(err, "while reading configuration from environment variables")

	log := logger.New(&cfg.Logger)
	stopCh := signal.SetupChannel()

	k8sConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	fatalOnError(err, "while creating k8s rest client config")

	// k8s client
	k8sCli, err := kubernetes.NewForConfig(k8sConfig)
	fatalOnError(err, "while creating k8s client")
	var (
		cfgMapNsScopedCli = k8sCli.CoreV1().ConfigMaps(cfg.WorkingNamespace)
		secretNsScopedCli = k8sCli.CoreV1().Secrets(cfg.WorkingNamespace)
	)

	// etcd-operator informers
	etcdOperatorCli, err := etcdOperatorClientset.NewForConfig(k8sConfig)
	fatalOnError(err, "while creating etcd operator clientset")
	var (
		etcdOpInformerFactory = etcdOperatorInformers.NewFilteredSharedInformerFactory(etcdOperatorCli, informerResyncPeriod, cfg.WorkingNamespace, nil)
		etcdOpInformersGroup  = etcdOpInformerFactory.Etcd().V1beta2()
		etcdOpNsScopedLister  = etcdOpInformersGroup.EtcdBackups().Lister().EtcdBackups(cfg.WorkingNamespace)
		etcdOpNsScopedCli     = etcdOperatorCli.EtcdV1beta2().EtcdBackups(cfg.WorkingNamespace)
	)

	// start informers
	etcdOpInformerFactory.Start(stopCh)

	// wait for cache sync
	etcdOpInformerFactory.WaitForCacheSync(stopCh)

	// STEP 1: backup service
	backupExecutor := backup.NewExecutor(cfg.Backup, cfg.ABS.SecretName, cfg.ABS.ContainerName, etcdOpNsScopedCli, etcdOpNsScopedLister, log)
	recBackupExecutor := backup.NewRecordedExecutor(backupExecutor, cfg.Backup.ConfigMapNameForTracing, cfgMapNsScopedCli)

	_, err = recBackupExecutor.SingleBackup(stopCh, cfg.BlobPrefix)
	fatalOnError(err, "while executing single etcd cluster backup")

	// STEP 2: rotate backup service
	azureCreds, err := azure.ExtractCredsFromSecret(cfg.ABS.SecretName, secretNsScopedCli)
	fatalOnError(err, "while extracting ABS credentials from secret")

	azBlobCli, err := azure.NewBlobContainerClient(azureCreds.AccountName, azureCreds.AccountKey, cfg.ABS.ContainerName)
	fatalOnError(err, "while creating Azure Blob client")

	backupCleaner := cleaner.NewAzure(cfg.Cleaner, azBlobCli, log)
	err = backupCleaner.Clean(stopCh, cfg.BlobPrefix)
	fatalOnError(err, "while removing old backups from ABS")
}

func newRestClientConfig(kubeConfigPath string) (*restclient.Config, error) {
	if kubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}

	return restclient.InClusterConfig()
}

func fatalOnError(err error, context string) {
	if err != nil {
		logrus.Fatal(errors.Wrap(err, context).Error())
	}
}
