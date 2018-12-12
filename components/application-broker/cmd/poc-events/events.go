package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/scheme"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/reference"

	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"

	typedV1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func main() {
	var kubeconfig *string
	if home := os.Getenv("HOME"); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	reClient, err := versioned.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(func(format string, args ...interface{}) {
		fmt.Printf(format, args...)
	})
	broadcaster.StartRecordingToSink(&typedV1.EventSinkImpl{Interface: clientset.CoreV1().Events(metav1.NamespaceDefault)})
	eventRecorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "Remote-Environment-Broker"})

	re, err := reClient.ApplicationconnectorV1alpha1().RemoteEnvironments().Get("ec-prod", metav1.GetOptions{})
	if err != nil {
		panic(errors.Wrap(err, "on getting remote environment"))
	}

	ref, err := reference.GetReference(scheme.Scheme, re)
	if err != nil {
		panic(errors.Wrap(err, "on getting reference for Remote Environment"))
	}

	eventRecorder.Event(ref, v1.EventTypeWarning, "SomeReason", "Some additional message")
	eventRecorder.Event(ref, v1.EventTypeWarning, "SomeReason", "Some additional message")
	eventRecorder.Event(ref, v1.EventTypeWarning, "SomeReason", "Some additional message")

	time.Sleep(time.Second)
}
