package k8s

//go:generate go run gen.go -- -group=Core -types-package=k8s.io/api/core/v1 -resources=Namespaces,Pods
//go:generate go run gen.go -- -group=Ui -types-package=github.com/kyma-project/kyma/components/console-backend-service2/pkg/apis/ui/v1alpha1 -resources=BackendModules
//go:generate go run gen.go -- -group=ApplicationConnector -types-package=github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1 -resources=Applications,ApplicationMappings
