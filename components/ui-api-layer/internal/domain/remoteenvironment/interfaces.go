package remoteenvironment

import (
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/resource"
)

// EventActivation

//go:generate mockery -name=eventActivationLister -output=automock -outpkg=automock -case=underscore
type eventActivationLister interface {
	List(environment string) ([]*v1alpha1.EventActivation, error)
}

//go:generate mockery -name=AsyncApiSpecGetter -output=automock -outpkg=automock -case=underscore
type AsyncApiSpecGetter interface {
	Find(kind, id string) (*storage.AsyncApiSpec, error)
}

// Notifier

type notifier interface {
	AddListener(observer resource.Listener)
	DeleteListener(observer resource.Listener)
}
