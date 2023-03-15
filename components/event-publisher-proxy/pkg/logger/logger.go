package logger

import (
	"log"
	"strings"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/api"
	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/watcher"
	kymalogger "github.com/kyma-project/kyma/components/eventing-controller/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// compile-time check for interfaces implementation.
var (
	_ watcher.UpdateNotifiable = &Logger{}
)

type Logger struct {
	config         *watcher.Config
	v1alpha1Client *api.Client
	k8sClient      kubernetes.Interface
}

func New(config *watcher.Config, k8sConfig *rest.Config) *Logger {
	dynamicClient := dynamic.NewForConfigOrDie(k8sConfig)
	subscriptionClient := api.NewClient(dynamicClient)
	return &Logger{
		config:         config,
		v1alpha1Client: subscriptionClient,
		k8sClient:      kubernetes.NewForConfigOrDie(k8sConfig),
	}
}

func (l *Logger) NotifyUpdate(cm *corev1.ConfigMap) {
	l.apply(cm)
}

func (l *Logger) apply(cm *corev1.ConfigMap) {
	if err := l.configureLogger(cm); err != nil {
		log.Fatalf("error:[%v]", err)
	}
}

func (l *Logger) configureLogger(cm *corev1.ConfigMap) error {

	for key, levelValue := range cm.Data {
		if !strings.EqualFold(key, "APP_LOG_LEVEL") {
			log.Fatalf("error:[%v]")
		}
		logger, err := kymalogger.New("json", levelValue)
		{
			log.Fatalf("error:[%v]", err)
		}
	}
}
