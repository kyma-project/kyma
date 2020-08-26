package teststep

import (
	"net/url"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/addons"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
)

type Addons struct {
	name        string
	addonConfig *addons.AddonConfiguration
	url         *url.URL
}

func NewAddonConfiguration(name string, addonConfig *addons.AddonConfiguration, url *url.URL) step.Step {
	return &Addons{
		name:        name,
		addonConfig: addonConfig,
		url:         url,
	}
}

func (a Addons) Name() string {
	return a.name
}

func (a Addons) Run() error {
	err := a.addonConfig.Create(a.url.String())
	if err != nil {
		return err
	}

	return errors.Wrap(a.addonConfig.WaitForStatusRunning(), "while checking if addon configuration is ready")
}

func (a Addons) Cleanup() error {
	if err := a.addonConfig.LogResource(); err != nil {
		return errors.Wrapf(err, "while logging resource")
	}

	return errors.Wrap(a.addonConfig.Delete(), "while deleting addong configuration")
}

var _ step.Step = Addons{}
