package model

import corev1 "k8s.io/api/core/v1"
import applicationoperatorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
import applicationbrokerv1alpha1 "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"

type Pod corev1.Pod
type Namespace corev1.Namespace
type Application applicationoperatorv1alpha1.Application
type ApplicationMapping applicationbrokerv1alpha1.ApplicationMapping