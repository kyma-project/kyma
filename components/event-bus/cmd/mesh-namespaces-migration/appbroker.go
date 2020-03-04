package main

import (
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	// allow client authentication against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const (
	appBrokerNamespace = "kyma-integration"
	appBrokerName      = "application-broker"
)

// appBrokerShutdownAndWait shuts down the Application broker and blocks until all its replicas have been terminated.
// The undo function can be used to revert to the original state.
func appBrokerShutdownAndWait(cli kubernetes.Interface) (undo func() error, err error) {
	appBrokerKey := fmt.Sprintf("%s/%s", appBrokerNamespace, appBrokerName)

	undo = func() error { return nil }

	currentScale, err := cli.AppsV1().Deployments(appBrokerNamespace).GetScale(appBrokerName, metav1.GetOptions{})
	switch {
	case apierrors.IsNotFound(err):
		return undo, nil
	case err != nil:
		return nil, errors.Wrapf(err, "getting scale of Deployment %q: %s", appBrokerKey)
	}

	currentScale.ResourceVersion = ""

	newScale := currentScale.DeepCopy()
	newScale.Spec.Replicas = 0

	log.Print("Shutting down Application broker")

	if _, err := cli.AppsV1().Deployments(appBrokerNamespace).UpdateScale(appBrokerName, newScale); err != nil {
		return nil, errors.Wrapf(err, "scaling Deployment %q", appBrokerKey)
	}

	if err := waitForDeploymentShutdown(cli, appBrokerNamespace, appBrokerName); err != nil {
		return nil, errors.Wrapf(err, "waiting for shutdown of Deployment %q", appBrokerKey)
	}

	undo = func() error {
		log.Print("Re-starting Application broker")

		if err := scaleDeploymentWithRetry(cli, appBrokerNamespace, appBrokerName, currentScale); err != nil {
			return errors.Wrapf(err, "scaling Deployment %q", appBrokerKey)
		}
		return nil
	}

	return undo, nil
}

// waitForDeploymentShutdown waits until all replicas of a Deployment have been terminated.
func waitForDeploymentShutdown(cli kubernetes.Interface, ns, name string) error {
	var expectNoReplica wait.ConditionFunc = func() (bool, error) {
		dep, err := cli.AppsV1().Deployments(ns).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if dep.Status.Replicas != 0 {
			return false, nil
		}

		return true, nil
	}

	return wait.PollImmediateUntil(time.Second, expectNoReplica, make(<-chan struct{}))
}

// scaleDeploymentWithRetry scales a Deployment and retries in case of failure.
func scaleDeploymentWithRetry(cli kubernetes.Interface, ns, name string, scale *autoscalingv1.Scale) error {
	var expectSuccessfulDeploymentScale wait.ConditionFunc = func() (bool, error) {
		_, err := cli.AppsV1().Deployments(ns).UpdateScale(name, scale)
		switch {
		case apierrors.IsConflict(err), apierrors.IsInvalid(err):
			return false, err
		case err != nil:
			return false, nil
		}
		return true, nil
	}

	return wait.PollImmediateUntil(5*time.Second, expectSuccessfulDeploymentScale, make(<-chan struct{}))
}
