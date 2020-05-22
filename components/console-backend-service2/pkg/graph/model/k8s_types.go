package model

import (
	applicationbrokerv1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	applicationoperatorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	uiv1alpha1 "github.com/kyma-project/kyma/components/console-backend-service2/pkg/apis/ui/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

//go:generate genny -in=k8s_types.genny -out=k8s_types_gen.go gen "Value=Namespace,Application,ApplicationMapping,Pod,BackendModule"

type Pod corev1.Pod
type Namespace corev1.Namespace

type Application applicationoperatorv1alpha1.Application
type ApplicationMapping applicationbrokerv1alpha1.ApplicationMapping

type BackendModule uiv1alpha1.BackendModule
