package cleaner

import "github.com/kyma-project/kyma/components/eventing-controller/logger"

type Cleaner interface {
	CleanSource(source string) (string, error)

	CleanEventType(eventType string) (string, error)
}

type JetStreamCleaner struct {
	logger *logger.Logger
}

type EventMeshCleaner struct {
	logger *logger.Logger
}
