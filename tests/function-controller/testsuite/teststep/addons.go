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
		log:         container.Log.WithField(step.LogStepKey, name),
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
	return errors.Wrap(a.addonConfig.Delete(), "while deleting addon configuration")
}

func (a Addons) OnError() error {
	return a.addonConfig.LogResource()
}

var _ step.Step = Addons{}
