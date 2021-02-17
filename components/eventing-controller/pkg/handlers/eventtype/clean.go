package eventtype

import "github.com/kyma-project/kyma/components/eventing-controller/pkg/application"

type Cleaner interface {
	Clean(eventType string) (string, error)
}

type cleaner struct {
	eventTypePrefix   string
	applicationLister *application.Lister
}

// compile-time check
var _ Cleaner = &cleaner{}

func NewCleaner(eventTypePrefix string, applicationLister *application.Lister) Cleaner {
	return &cleaner{eventTypePrefix: eventTypePrefix, applicationLister: applicationLister}
}

// Clean cleans the application name segment in the event-type from none-alphanumeric characters
func (e cleaner) Clean(eventType string) (string, error) {
	applicationName, event, version, err := parse(eventType, e.eventTypePrefix)
	if err != nil {
		return "", err
	}

	app, err := e.applicationLister.Get(applicationName)
	if err != nil {
		return "", err
	}

	applicationNameClean := application.CleanName(app)
	eventType = build(e.eventTypePrefix, applicationNameClean, event, version)

	return eventType, nil
}
