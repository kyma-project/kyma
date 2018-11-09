package utils

import (
	"fmt"

	arkapi "github.com/heptio/ark/pkg/apis/ark/v1"
	kubelessapi "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	scapi "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sbuapi "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	acapi "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"

	uuid "github.com/satori/go.uuid"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewEnvironmentNamespace returns a namespace labeled as kyma's environment
func NewEnvironmentNamespace(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"env":             "true",
				"istio-injection": "enabled",
			},
		},
	}
}

// NewRemoteEnvironment returns a RemoteEnvironment with one service class containing both service and events
func NewRemoteEnvironment(namespace string, envServiceName string) *acapi.RemoteEnvironment {
	return &acapi.RemoteEnvironment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RemoteEnvironment",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("re-%s", namespace),
		},
		Spec: acapi.RemoteEnvironmentSpec{
			AccessLabel: "access-label-1",
			Description: "Remote Environment",
			Services: []acapi.Service{
				{
					ID:   uuid.NewV4().String(),
					Name: envServiceName,
					Labels: map[string]string{
						"connected-app": "ec-prod",
					},
					ProviderDisplayName: "HakunaMatata",
					DisplayName:         "Promotions",
					Description:         "Remote Environment Service Class",
					Tags:                []string{"occ", "promotions"},
					Entries: []acapi.Entry{
						{
							Type:        "API",
							AccessLabel: "access-label-1",
							GatewayUrl:  fmt.Sprintf("http://promotions-gateway.%s.svc.cluster.local/", namespace),
						},
						{
							Type: "Events",
						},
					},
				},
			},
		},
	}
}

// NewEnvironmentMapping returns EnvironmentMapping for provided namespace
func NewEnvironmentMapping(namespace string) *acapi.EnvironmentMapping {
	return &acapi.EnvironmentMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EnvironmentMapping",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("re-%s", namespace),
			Namespace: namespace,
		},
	}
}

// NewFunction returns a Function with empty handler and a service port
func NewFunction(name string, namespace string) *kubelessapi.Function {
	return &kubelessapi.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: "kubeless.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kubelessapi.FunctionSpec{
			Handler:             "handler.main",
			Runtime:             "nodejs6",
			Function:            "module.exports = { main: function (event, context) {} }",
			FunctionContentType: "text",
			ServiceSpec: v1.ServiceSpec{
				Selector: map[string]string{
					"created-by": "kubeless",
					"function":   name,
				},
				Ports: []v1.ServicePort{
					{
						Name:       "http-function-port",
						Port:       8080,
						Protocol:   "TCP",
						TargetPort: intstr.FromInt(8080),
					},
				},
			},
		},
	}
}

// NewServiceInstance returns ServiceInstance
// use flag isClusterScoped to control whether serviceClass/servicePlan should be set as cluster or non-cluster
// use params to provide additional parameters for the instance
func NewServiceInstance(isClusterScoped bool, params *runtime.RawExtension, namespace, serviceClass, servicePlan, instanceName string) *scapi.ServiceInstance {
	si := scapi.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceInstance",
			APIVersion: "servicecatalog.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instanceName,
			Namespace: namespace,
		},
		Spec: scapi.ServiceInstanceSpec{
			Parameters: params,
		},
	}

	if isClusterScoped {
		si.Spec.PlanReference.ClusterServiceClassExternalName = serviceClass
		si.Spec.PlanReference.ClusterServicePlanExternalName = servicePlan
	} else {
		si.Spec.PlanReference.ServiceClassExternalName = serviceClass
		si.Spec.PlanReference.ServicePlanExternalName = servicePlan
	}

	return &si
}

// NewServiceBinding returns a ServiceBindings for provided service instance instanceName
func NewServiceBinding(name string, namespace string, instanceName string, params *runtime.RawExtension) *scapi.ServiceBinding {
	return &scapi.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBinding",
			APIVersion: "servicecatalog.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: scapi.ServiceBindingSpec{
			ServiceInstanceRef: scapi.LocalObjectReference{
				Name: instanceName,
			},
			Parameters: params,
		},
	}
}

// NewServiceBindingUsage returns a ServiceBindingUsage
func NewServiceBindingUsage(name, namespace, sbName, usageKind, usageName string) *sbuapi.ServiceBindingUsage {
	return &sbuapi.ServiceBindingUsage{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceBindingUsage",
			APIVersion: "servicecatalog.kyma.cx/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: sbuapi.ServiceBindingUsageSpec{
			ServiceBindingRef: sbuapi.LocalReferenceByName{
				Name: sbName,
			},
			UsedBy: sbuapi.LocalReferenceByKindAndName{
				Name: usageName,
				Kind: usageKind,
			},
		},
	}
}

// NewBackup returns a heptio ark's Backup with includedNamespaces and includedResources
func NewBackup(name string, includedNamespaces []string, includedResources []string) *arkapi.Backup {
	return &arkapi.Backup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Backup",
			APIVersion: "ark.heptio.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "heptio-ark",
		},
		Spec: arkapi.BackupSpec{
			IncludedNamespaces: includedNamespaces,
			IncludedResources:  includedResources,
		},
	}
}

// NewRestore returns a heptio ark's Restore for provided backupName
func NewRestore(name string, backupName string) *arkapi.Restore {
	return &arkapi.Restore{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Restore",
			APIVersion: "ark.heptio.com/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "heptio-ark",
		},
		Spec: arkapi.RestoreSpec{
			BackupName: backupName,
		},
	}
}
