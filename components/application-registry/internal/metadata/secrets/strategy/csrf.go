package strategy

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

func convertToModel(appCSRFInfo *applications.CSRFInfo) *model.CSRFInfo {
	if appCSRFInfo == nil {
		return nil
	}

	return &model.CSRFInfo{
		TokenEndpointURL: appCSRFInfo.TokenEndpointURL,
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

	if credentials.Oauth != nil {
		return convertFromModel(credentials.Oauth.CSRFInfo)
	}

	if credentials.Basic != nil {
		return convertFromModel(credentials.Basic.CSRFInfo)
	}

	if credentials.CertificateGen != nil {
		return convertFromModel(credentials.CertificateGen.CSRFInfo)
	}

	return nil
}
