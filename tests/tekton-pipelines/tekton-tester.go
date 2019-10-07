package main

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	tekton "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"os"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	ns       = "tekton-pipelines"
	name     = "test-build"
	label    = "test-build-func"
	image    = "alpine:3.8"
	response = "hello build"
)

var kubeConfig *rest.Config

func main() {
	kubeConfig = loadKubeConfigOrDie()
	tektonClient := tekton.NewForConfigOrDie(kubeConfig).TaskRuns(ns)
	if err := runTest(tektonClient); err != nil {
		log.Fatalf("Failed: %s", err)
	}
}

func runTest(tektonClient tekton.TaskRunInterface) error {
	labels := make(map[string]string)
	labels[label] = name
	tr, err := tektonClient.Create(&tektonv1alpha1.TaskRun{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Labels:    labels,
			Namespace: ns,
		},
		Spec: tektonv1alpha1.TaskRunSpec{
			TaskSpec: &tektonv1alpha1.TaskSpec{
				Steps: []tektonv1alpha1.Step{{
					Container: corev1.Container{
						Name:    name,
						Image:   image,
						Command: []string{"echo", "-n", response},
					},
				}},
			},
		},
		Status: tektonv1alpha1.TaskRunStatus{},
	})
	if err != nil {
		return errors.Wrap(err, "unable to create taskrun")
	}
	defer deleteBuild(tektonClient, tr)
	err = retry.Do(func() error {
		tr, err := tektonClient.Get(name, meta.GetOptions{})
		if err != nil {
			return err
		}

		if tr.Status.Conditions == nil {
			return fmt.Errorf("object not reconciled yet")
		}
		for _, condition := range tr.Status.Conditions {
			if condition.Type == apis.ConditionSucceeded && condition.Status == corev1.ConditionFalse {
				return retry.Unrecoverable(fmt.Errorf(condition.Message))
			} else if condition.Type == apis.ConditionSucceeded && condition.Status == corev1.ConditionUnknown {
				return fmt.Errorf(condition.Message)
			}
		}
		return nil
	}, retry.Attempts(15),
		retry.DelayType(retry.FixedDelay),
		retry.Delay(5*time.Second),
	)
	if err != nil {
		return errors.Wrap(err, "build failed")
	}
	return nil
}

func loadKubeConfigOrDie() *rest.Config {
	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalf("Cannot create in-cluster config: %v", err)
		}
		return cfg
	}

	var err error
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatalf("Cannot read kubeconfig: %s", err)
	}
	return kubeConfig
}

func deleteBuild(tektonClient tekton.TaskRunInterface, tr *tektonv1alpha1.TaskRun) {
	var deleteImmediately int64
	if err := tektonClient.Delete(tr.Name, &meta.DeleteOptions{GracePeriodSeconds: &deleteImmediately}); err != nil {
		log.Fatalf("Cannot delete TaskRun %v: %v", tr.Name, err)
	}
}
