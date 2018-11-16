package specification

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/model"
)

type SpecData struct {
	Id         string
	API        *model.API
	Events     *model.Events
	GatewayUrl string
	Docs       []byte
}
