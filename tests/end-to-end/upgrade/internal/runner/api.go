package runner

import "github.com/sirupsen/logrus"

type (
	// UpgradeTest allows to execute test in a generic way
	UpgradeTest interface {
		CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error
		TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error
	}
)
