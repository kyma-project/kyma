package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/actionmanager"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/conditionmanager"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/consts"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/finalizer"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/installation"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymainstallation"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/servicecatalog"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/toolkit"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/statusmanager"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/steps"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned"
	informers "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/informers/externalversions"
)

var gitCommitHash string

const STDOUT = "/dev/stdout"

func main() {

	log.Println("Starting operator...")

	if gitCommitHash != "" {
		log.Println("Git commit hash:", gitCommitHash)
	}

	stop := make(chan struct{})

	kubeconfig := flag.String("kubeconfig", "", "Path to a kubeconfig file")
	helmHost := flag.String("helmhost", "tiller-deploy.kube-system.svc.cluster.local:44134", "Helm host")
	kymaDir := flag.String("kymadir", "/kyma", "Directory where kyma packages will be extracted")
	tlsKey := flag.String("tillerTLSKey", "/etc/certs/tls.key", "Path to TLS key file")
	tlsCrt := flag.String("tillerTLSCrt", "/etc/certs/tls.crt", "Path to TLS cert file")
	TLSInsecureSkipVerify := flag.Bool("tillerTLSInsecureSkipVerify", false, "Disable verification of Tiller TLS cert")
	backoffIntervalsRaw := flag.String("backoffIntervals", "10,20,40,60,80", "Number of seconds to wait before subsequent retries")
	overrideLogFile := flag.String("overrideLogFile", STDOUT, "Log File to Print Installation overrides. (Default: /dev/stdout)")
	overrideLogFormat := flag.String("overrideLogFormat", "text", "Installation Override Log format (Accepted values: text or json)")

	flag.Parse()

	backoffIntervals, err := parseBackoffIntervals(*backoffIntervalsRaw)
	if err != nil {
		log.Fatalf("Unable to parse backoff intervals configuration. Error: %v", err)
	}

	config, err := getClientConfig(*kubeconfig)
	if err != nil {
		log.Fatalf("Unable to build kubernetes configuration. Error: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create kubernetes client. Error: %v", err)
	}

	internalClient, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create internal client. Error: %v", err)
	}

	overridesLogger, closeFn, err := setupLogrus(*overrideLogFile, getLogrusFormatter(*overrideLogFormat))
	if err != nil {
		log.Fatalf("Unable to create logrus Instance. Error: %v", err)
	}

	helmClient, err := kymahelm.NewClient(*helmHost, *tlsKey, *tlsCrt, *TLSInsecureSkipVerify, overridesLogger)
	if err != nil {
		log.Fatalf("Unable create helm client. Error: %v", err)
	}

	serviceCatalogClient := servicecatalog.NewClient(config)
	kymaCommandExecutor := &toolkit.KymaCommandExecutor{}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	internalInformerFactory := informers.NewSharedInformerFactory(internalClient, time.Second*30)
	installationLister := internalInformerFactory.Installer().V1alpha1().Installations().Lister()

	kymaStatusManager := statusmanager.NewKymaStatusManager(internalClient, installationLister)
	kymaActionManager := actionmanager.NewKymaActionManager(internalClient, installationLister)
	conditionManager := conditionmanager.New(internalClient, installationLister)

	installationFinalizerManager := finalizer.NewManager(consts.InstFinalizer)

	fsWrapper := kymasources.NewFilesystemWrapper()

	kymaPackages := kymasources.NewKymaPackages(fsWrapper, kymaCommandExecutor, *kymaDir)
	stepFactoryCreator := kymainstallation.NewStepFactoryCreator(helmClient, kymaPackages, fsWrapper, *kymaDir)
	installationSteps := steps.New(serviceCatalogClient, kymaStatusManager, kymaActionManager, stepFactoryCreator, backoffIntervals)

	installationController := installation.NewController(kubeClient, kubeInformerFactory, internalInformerFactory, installationSteps, conditionManager, installationFinalizerManager, internalClient)

	kubeInformerFactory.Start(stop)
	internalInformerFactory.Start(stop)

	installationController.Run(stop)

	closeFn(stop)
}

func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func parseBackoffIntervals(backoffIntervals string) ([]uint, error) {
	backoffIntervalsParsed := []uint{}
	intervalsWithoutTrailingCommas := strings.TrimRight(backoffIntervals, ",")
	backoffIntervalsSplit := strings.Split(intervalsWithoutTrailingCommas, ",")
	for _, interval := range backoffIntervalsSplit {
		trimmed := strings.TrimSpace(interval)
		parsedInterval, err := strconv.ParseUint(trimmed, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("not able to parse value: \"%v\" as uint", trimmed)
		}
		backoffIntervalsParsed = append(backoffIntervalsParsed, uint(parsedInterval))
	}

	return backoffIntervalsParsed, nil
}

func setupLogrus(logFile string, formatter logrus.Formatter) (*logrus.Logger, func(ch <-chan struct{}), error) {
	// create the logger
	logger := logrus.New()
	logger.SetFormatter(formatter)

	//open the log file
	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return nil, nil, err
	}
	logger.SetOutput(file)
	if logFile == STDOUT { //if the output file is /dev/stdout, we should not return a function to close the file
		return logger, func(ch <-chan struct{}) {}, nil
	}

	// return logger instance and function which can close the log file at recieving signal from the channel
	return logger, func(ch <-chan struct{}) {
		defer func() {
			log.Println("Closing log file")
			if err := file.Close(); err != nil {
				log.Fatalf("Unable to close the logfile with error: %+v", err)
			}
		}()
		// Wait for Channel to send stop signal
		<-ch
	}, nil
}

func getLogrusFormatter(format string) logrus.Formatter {
	switch format {
	case "json":
		return new(logrus.JSONFormatter)
	}
	return new(logrus.TextFormatter)
}
