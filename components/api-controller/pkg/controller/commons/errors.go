package commons

import (
	"fmt"

	"github.com/gofrs/uuid"
	log "github.com/sirupsen/logrus"
)

//HandleError Logs root cause and returns new error to hide implementation details
func HandleError(rootCause error, msg string) error {
	errID, err := uuid.NewV4()
	if err != nil {
		log.Fatalf("failed to generate UUID: %v", err)
	}
	log.Errorf("[Error '%s']: %v", errID, rootCause)
	return fmt.Errorf("%s (error code = '%s')", msg, errID)
}
