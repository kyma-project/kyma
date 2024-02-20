package director

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	kymamodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

func (app Application) ToApplication() kymamodel.Application {
	var bundles []kymamodel.APIBundle
	if app.Bundles != nil {
		bundles = convertAPIBundles(app.Bundles.Data)
	}

	description := ""
	if app.Description != nil {
		description = *app.Description
	}

	providerName := ""
	if app.ProviderName != nil {
		providerName = *app.ProviderName
	}

	return kymamodel.Application{
		ID:                  app.ID,
		Name:                app.Name,
		ProviderDisplayName: providerName,
		Description:         description,
		Labels:              app.Labels,
		SystemAuthsIDs:      extractSystemAuthIDs(app.Auths),
		ApiBundles:          bundles,
	}
}

func convertAPIBundles(apiBundles []*graphql.BundleExt) []kymamodel.APIBundle {
	bundles := make([]kymamodel.APIBundle, len(apiBundles))

	for i, apiBundle := range apiBundles {
		bundles[i] = convertAPIBundle(apiBundle)
	}

	return bundles
}

func convertAPIBundle(apiBundle *graphql.BundleExt) kymamodel.APIBundle {
	apis := convertAPIsExt(apiBundle.APIDefinitions.Data)
	eventAPIs := convertEventAPIsExt(apiBundle.EventDefinitions.Data)
	defaultInstanceAuth := convertAuth(apiBundle.DefaultInstanceAuth)

	var authRequestInputSchema *string
	if apiBundle.InstanceAuthRequestInputSchema != nil {
		s := string(*apiBundle.InstanceAuthRequestInputSchema)
		authRequestInputSchema = &s
	}
	return kymamodel.APIBundle{
		ID:                             apiBundle.ID,
		Name:                           apiBundle.Name,
		Description:                    apiBundle.Description,
		InstanceAuthRequestInputSchema: authRequestInputSchema,
		APIDefinitions:                 apis,
		EventDefinitions:               eventAPIs,
		DefaultInstanceAuth:            defaultInstanceAuth,
	}
}

func convertAPIsExt(compassAPIs []*graphql.APIDefinitionExt) []kymamodel.APIDefinition {
	apis := make([]kymamodel.APIDefinition, len(compassAPIs))

	for i, cAPI := range compassAPIs {
		apis[i] = convertAPIExt(cAPI)
	}

	return apis
}

func convertEventAPIsExt(compassEventAPIs []*graphql.EventAPIDefinitionExt) []kymamodel.EventAPIDefinition {
	eventAPIs := make([]kymamodel.EventAPIDefinition, len(compassEventAPIs))

	for i, cAPI := range compassEventAPIs {
		eventAPIs[i] = convertEventAPIExt(cAPI)
	}

	return eventAPIs
}

func extractSystemAuthIDs(auths []*graphql.AppSystemAuth) []string {
	ids := make([]string, 0, len(auths))

	for _, auth := range auths {
		ids = append(ids, auth.ID)
	}

	return ids
}

func convertAPIExt(compassAPI *graphql.APIDefinitionExt) kymamodel.APIDefinition {
	description := ""
	if compassAPI.Description != nil {
		description = *compassAPI.Description
	}

	api := kymamodel.APIDefinition{
		ID:          compassAPI.ID,
		Name:        compassAPI.Name,
		Description: description,
		TargetUrl:   compassAPI.TargetURL,
	}

	return api
}

func convertEventAPIExt(compassEventAPI *graphql.EventAPIDefinitionExt) kymamodel.EventAPIDefinition {
	description := ""
	if compassEventAPI.Description != nil {
		description = *compassEventAPI.Description
	}

	eventAPI := kymamodel.EventAPIDefinition{
		ID:          compassEventAPI.ID,
		Name:        compassEventAPI.Name,
		Description: description,
	}

	return eventAPI
}

func convertAuth(compassAuth *graphql.Auth) *kymamodel.Auth {
	if compassAuth == nil {
		return nil
	}
	return &kymamodel.Auth{
		Credentials:       convertCredentials(compassAuth),
		RequestParameters: convertRequestParameters(compassAuth),
	}
}

func convertCredentials(compassAuth *graphql.Auth) *kymamodel.Credentials {
	if compassAuth == nil {
		return nil
	}
	switch compassAuth.Credential.(type) {
	case *graphql.OAuthCredentialData:
		v := compassAuth.Credential.(*graphql.OAuthCredentialData)
		if v != nil {
			return &kymamodel.Credentials{
				Oauth: &kymamodel.Oauth{
					URL:          v.URL,
					ClientID:     v.ClientID,
					ClientSecret: v.ClientSecret,
				},
				CSRFInfo: convertCSRFInfo(compassAuth),
			}
		}
	case *graphql.BasicCredentialData:
		v := compassAuth.Credential.(*graphql.BasicCredentialData)
		if v != nil {
			return &kymamodel.Credentials{
				Basic: &kymamodel.Basic{
					Username: v.Username,
					Password: v.Password,
				},
				CSRFInfo: convertCSRFInfo(compassAuth),
			}
		}
	}
	return nil
}

func convertRequestParameters(compassAuth *graphql.Auth) *kymamodel.RequestParameters {
	if compassAuth.AdditionalHeaders != nil || compassAuth.AdditionalQueryParams != nil {
		result := &kymamodel.RequestParameters{}
		if compassAuth.AdditionalHeaders != nil {
			v := map[string][]string(compassAuth.AdditionalHeaders)
			result.Headers = &v
		}
		if compassAuth.AdditionalQueryParams != nil {
			v := map[string][]string(compassAuth.AdditionalQueryParams)
			result.QueryParameters = &v
		}
		return result
	}
	return nil
}

func convertCSRFInfo(compassAuth *graphql.Auth) *kymamodel.CSRFInfo {
	if compassAuth.RequestAuth != nil && compassAuth.RequestAuth.Csrf != nil {
		return &kymamodel.CSRFInfo{
			TokenEndpointURL: compassAuth.RequestAuth.Csrf.TokenEndpointURL,
		}
	}
	return nil
}
