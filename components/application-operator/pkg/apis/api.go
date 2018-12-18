package apis

import (
	remoteenvs "github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func AddREToScheme(s *runtime.Scheme) error {
	return remoteenvs.SchemeBuilder.AddToScheme(s)
}
