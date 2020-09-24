package teststep

import (
	"net/url"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/addons"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
)

type Addons struct {
	name        string
	addonConfig *addons.AddonConfiguration
	url         *url.URL
	log         *logrus.Entry
}

func NewAddonConfiguration(name string, addonConfig *addons.AddonConfiguration, url *url.URL, container shared.Container) step.Step {
	return &Addons{
		name:        name,
		addonConfig: addonConfig,
		url:         url,
		log:         container.Log,
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
		a.log.Warn(errors.Wrapf(err, "while logging resource"))
	}

	return errors.Wrap(a.addonConfig.Delete(), "while deleting addon configuration")
}

var _ step.Step = Addons{}
