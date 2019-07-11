package bus

import (
	"fmt"
)

type configurationData struct {
	SourceID string
}

//Conf Event-Service configuration data
var Conf *configurationData

var eventsTargetURLV1 string
var eventsTargetURLV2 string

// Init should be used to initialize the "source" related configuration data
func Init(sourceID string, targetURLV1 string, targetURLV2 string) {
	Conf = &configurationData{
		SourceID: sourceID,
	}
	eventsTargetURLV1 = targetURLV1
	eventsTargetURLV2 = targetURLV2
}

//CheckConf assert the configuration initialization
func CheckConf() (err error) {
	if Conf == nil {
		return fmt.Errorf("configuration data not initialized")
	}
	return nil
}
