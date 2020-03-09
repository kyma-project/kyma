package eventing

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
)

//go:generate mockery -name=gqlAssetConverter -output=automock -outpkg=automock -case=underscore
//go:generate failery -name=gqlAssetConverter -case=underscore -output disabled -outpkg disabled
type gqlTriggerConverter interface {
	ToGQL(in *v1alpha1.Trigger) (*gqlschema.Trigger, error)
	ToTrigger(in *gqlschema.Trigger) (*v1alpha1.Trigger, error)
}

type triggerConverter struct {
}

func newTriggerConverter() gqlTriggerConverter {
	return &triggerConverter{}
}

func (c *triggerConverter) ToGQL(in *v1alpha1.Trigger) (*gqlschema.Trigger, error) {
	return nil, nil
}

func (c *triggerConverter) ToTrigger(in *gqlschema.Trigger) (*v1alpha1.Trigger, error) {
	return nil, nil
}
