package process

import "github.com/sirupsen/logrus"

type Process struct {
	Steps           []Step
	ReleaseName     string
	BEBNamespace    string
	EventingBackend string
	EventTypePrefix string
	Clients         Clients
	Logger          *logrus.Logger
	State           State
}
