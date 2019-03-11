package helloworld

import "github.com/sirupsen/logrus"

type HelloWorld struct {
}

func (h *HelloWorld) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return nil
}

func (h *HelloWorld) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return nil
}

func (h *HelloWorld) Name() string {
	return "HelloWorld example"
}
