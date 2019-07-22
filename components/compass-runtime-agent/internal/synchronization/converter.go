package synchronization

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
)

type Converter interface {
	Do(application compass.Application) v1alpha1.Application
}

type converter struct {
}

func (c converter) Do(application compass.Application) v1alpha1.Application {
	return v1alpha1.Application{}
}

func NewConverter() Converter {
	return converter{}
}
