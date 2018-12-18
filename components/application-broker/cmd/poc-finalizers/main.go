package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"path/filepath"

	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/informers/externalversions"
	"github.com/sirupsen/logrus"
	restclient "k8s.io/client-go/rest"

	"k8s.io/client-go/tools/clientcmd"
)

// informerResyncPeriod defines how often informer will execute relist action. Setting to zero disable resync.
// BEWARE: too short period time will increase the CPU load.
const informerResyncPeriod = 30 * time.Minute

func main() {
	var kubeconfig *string
	if home := os.Getenv("HOME"); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	log := (&logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.StampMicro,
		},
		Hooks: make(logrus.LevelHooks),
		Level: logrus.InfoLevel,
	}).WithField("service", "main")

	// create sync-job
	k8sConfig := newRestClientConfig(*kubeconfig)

	reClient, err := versioned.NewForConfig(k8sConfig)
	fatalOnError(err)

	// Always prefer using an informer factory to get a shared informer instead of getting an independent
	// one. This reduces memory footprint and number of connections to the server.
	informerFactory := externalversions.NewSharedInformerFactory(reClient, informerResyncPeriod)

	v1alpha1Interface := informerFactory.Applicationconnector().V1alpha1()

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	stopCh := make(chan struct{})
	cancelOnInterrupt(ctx, stopCh, cancelFunc)

	/* protection controller */
	protectionController := NewProtectionController(v1alpha1Interface.RemoteEnvironments(),
		reClient.ApplicationconnectorV1alpha1().EnvironmentMappings("default"),
		reClient.ApplicationconnectorV1alpha1().RemoteEnvironments(), log)
	protectionController.Run(1, stopCh)

	informerFactory.Start(stopCh)

	<-stopCh
}

func fatalOnError(err error) {
	if err != nil {
		logrus.Fatal(err.Error())
	}
}

// cancelOnInterrupt closes given channel and also calls cancel func when os.Interrupt or SIGTERM is received
func cancelOnInterrupt(ctx context.Context, ch chan<- struct{}, cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
			close(ch)
		case <-c:
			close(ch)
			cancel()
		}
	}()
}

func newRestClientConfig(kubeconfigPath string) *restclient.Config {
	var config *restclient.Config
	var err error
	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = restclient.InClusterConfig()
	}
	fatalOnError(err)
	return config
}
