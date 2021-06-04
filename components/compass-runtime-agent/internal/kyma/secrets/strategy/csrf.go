package strategy

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

func convertToModelCSRInfo(appCredentials *applications.Credentials) *model.CSRFInfo {
	if appCredentials == nil || appCredentials.CSRFInfo == nil {
		return nil
	}

	return &model.CSRFInfo{
		TokenEndpointURL: appCredentials.CSRFInfo.TokenEndpointURL,
	}
}

func toAppCSRFInfo(credentials *model.Credentials) *applications.CSRFInfo {

	convertFromModel := func(csrfInfo *model.CSRFInfo) *applications.CSRFInfo {
		if csrfInfo == nil {
			return nil
		}

		return &applications.CSRFInfo{
			TokenEndpointURL: csrfInfo.TokenEndpointURL,
		}
	}

	if credentials == nil {
		return nil
	}

	return convertFromModel(credentials.CSRFInfo)
}
