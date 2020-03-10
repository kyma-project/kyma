package director

import (
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	kymamodel "kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

func (app Application) ToApplication() kymamodel.Application {

	var packages []kymamodel.APIPackage
	if app.Packages != nil {
		packages = convertAPIPackages(app.Packages.Data)
	}

	description := ""
	if app.Description != nil {
		description = *app.Description
	}

	providerName := ""
	if app.ProviderName != nil {
		providerName = *app.ProviderName
	}

	var apis []kymamodel.APIDefinition
	if app.APIDefinitions != nil {
		apis = convertAPIs(app.APIDefinitions.Data)
	}

	var eventAPIs []kymamodel.EventAPIDefinition
	if app.EventDefinitions != nil {
		eventAPIs = convertEventAPIs(app.EventDefinitions.Data)
	}

	var documents []kymamodel.Document
	if app.Documents != nil {
		documents = convertDocuments(app.Documents.Data)
	}

	return kymamodel.Application{
		ID:                  app.ID,
		Name:                app.Name,
		ProviderDisplayName: providerName,
		Description:         description,
		Labels:              map[string]interface{}(app.Labels),
		SystemAuthsIDs:      extractSystemAuthIDs(app.Auths),
		APIPackages:         packages,
		APIs:                apis,
		EventAPIs:           eventAPIs,
		Documents:           documents,
	}

}

func convertAPIPackages(compassAPIs []*graphql.PackageExt) []kymamodel.APIPackage {
	var packages = make([]kymamodel.APIPackage, len(compassAPIs))

	for i, cAPI := range compassAPIs {
		packages[i] = convertAPIPackage(cAPI)
	}

	return packages
}

func convertAPIPackage(p *graphql.PackageExt) kymamodel.APIPackage {

	apis := convertAPIsExt(p.APIDefinitions.Data)
	eventAPIs := convertEventAPIsExt(p.EventDefinitions.Data)
	documents := convertDocumentsExt(p.Documents.Data)

	var authRequestInputSchema *string
	if p.InstanceAuthRequestInputSchema != nil {
		s := string(*p.InstanceAuthRequestInputSchema)
		authRequestInputSchema = &s
	}

	return kymamodel.APIPackage{
		ID:                             p.ID,
		Name:                           p.Name,
		Description:                    p.Description,
		InstanceAuthRequestInputSchema: authRequestInputSchema,
		APIDefinitions:                 apis,
		EventDefinitions:               eventAPIs,
		Documents:                      documents,
	}
}

func convertAPIs(compassAPIs []*graphql.APIDefinition) []kymamodel.APIDefinition {
	var apis = make([]kymamodel.APIDefinition, len(compassAPIs))

	for i, cAPI := range compassAPIs {
		apis[i] = convertAPI(cAPI)
	}

	return apis
}

func convertAPIsExt(compassAPIs []*graphql.APIDefinitionExt) []kymamodel.APIDefinition {
	var apis = make([]kymamodel.APIDefinition, len(compassAPIs))

	for i, cAPI := range compassAPIs {
		apis[i] = convertAPIExt(cAPI)
	}

	return apis
}

func convertEventAPIs(compassEventAPIs []*graphql.EventDefinition) []kymamodel.EventAPIDefinition {
	var eventAPIs = make([]kymamodel.EventAPIDefinition, len(compassEventAPIs))

	for i, cAPI := range compassEventAPIs {
		eventAPIs[i] = convertEventAPI(cAPI)
	}

	return eventAPIs
}

func convertEventAPIsExt(compassEventAPIs []*graphql.EventAPIDefinitionExt) []kymamodel.EventAPIDefinition {
	var eventAPIs = make([]kymamodel.EventAPIDefinition, len(compassEventAPIs))

	for i, cAPI := range compassEventAPIs {
		eventAPIs[i] = convertEventAPIExt(cAPI)
	}

	return eventAPIs
}

func convertDocuments(compassDocuments []*graphql.Document) []kymamodel.Document {
	var documents = make([]kymamodel.Document, len(compassDocuments))

	for i, cDoc := range compassDocuments {
		documents[i] = convertDocument(cDoc)
	}

	return documents
}

func convertDocumentsExt(compassDocuments []*graphql.DocumentExt) []kymamodel.Document {
	var documents = make([]kymamodel.Document, len(compassDocuments))

	for i, cDoc := range compassDocuments {
		documents[i] = convertDocument(&cDoc.Document)
	}

	return documents
}

func extractSystemAuthIDs(auths []*graphql.SystemAuth) []string {
	ids := make([]string, 0, len(auths))

	for _, auth := range auths {
		ids = append(ids, auth.ID)
	}

	return ids
}

func convertAPI(compassAPI *graphql.APIDefinition) kymamodel.APIDefinition {
	logrus.Infof("Convert API Definition")
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

	if compassAPI.Spec != nil {
		logrus.Infof("API for %s is not nil", compassAPI.ID)
		var data []byte
		if compassAPI.Spec.Data != nil {
			data = []byte(*compassAPI.Spec.Data)
		}

		api.APISpec = &kymamodel.APISpec{
			Type:   kymamodel.APISpecType(compassAPI.Spec.Type),
			Data:   data,
			Format: kymamodel.SpecFormat(compassAPI.Spec.Format),
		}
	} else {
		logrus.Infof("API for %s is nil", compassAPI.ID)
	}

	// TODO: implement RequestParameters after update in Compass

	// TODO: we should use compassAPI.Auth instead of compassAPI.DefaultAuth but it is not working yet in Director
	//if compassAPI.Auth != nil {
	//	credentials, err := convertAuth(compassAPI.Auth.Auth)
	//	if err != nil {
	//		logrus.Errorf("Failed to convert Compass Authentication to credentials: %s", err.Error())
	//	} else {
	//		api.Credentials = credentials
	//	}
	//}

	if compassAPI.DefaultAuth != nil {
		credentials, err := convertAuth(compassAPI.DefaultAuth)
		if err != nil {
			logrus.Errorf("Failed to convert Compass Authentication to credentials: %s", err.Error())
		} else {
			api.Credentials = credentials
		}
	}

	return api
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

	if compassAPI.Spec != nil {
		var data []byte
		if compassAPI.Spec.Data != nil {
			data = []byte(*compassAPI.Spec.Data)
		}

		api.APISpec = &kymamodel.APISpec{
			Type:   kymamodel.APISpecType(compassAPI.Spec.Type),
			Data:   data,
			Format: kymamodel.SpecFormat(compassAPI.Spec.Format),
		}
	}

	return api
}

func convertAuth(compassAuth *graphql.Auth) (*kymamodel.Credentials, error) {
	credentials := &kymamodel.Credentials{}

	switch cred := compassAuth.Credential.(type) {
	case *graphql.BasicCredentialData:
		credentials.Basic = &kymamodel.Basic{
			Username: cred.Username,
			Password: cred.Password,
		}
	case *graphql.OAuthCredentialData:
		credentials.Oauth = &kymamodel.Oauth{
			URL:          cred.URL,
			ClientID:     cred.ClientID,
			ClientSecret: cred.ClientSecret,
		}
	default:
		return nil, errors.New("invalid credentials type")
	}

	if compassAuth.RequestAuth != nil && compassAuth.RequestAuth.Csrf != nil {
		credentials.CSRFInfo = &kymamodel.CSRFInfo{
			TokenEndpointURL: compassAuth.RequestAuth.Csrf.TokenEndpointURL,
		}
	}

	return credentials, nil
}

func convertEventAPI(compassEventAPI *graphql.EventDefinition) kymamodel.EventAPIDefinition {
	description := ""
	if compassEventAPI.Description != nil {
		description = *compassEventAPI.Description
	}

	eventAPI := kymamodel.EventAPIDefinition{
		ID:          compassEventAPI.ID,
		Name:        compassEventAPI.Name,
		Description: description,
	}

	if compassEventAPI.Spec != nil {

		var data []byte
		if compassEventAPI.Spec.Data != nil {
			data = []byte(*compassEventAPI.Spec.Data)
		}

		eventAPI.EventAPISpec = &kymamodel.EventAPISpec{
			Type:   kymamodel.EventAPISpecType(compassEventAPI.Spec.Type),
			Data:   data,
			Format: kymamodel.SpecFormat(compassEventAPI.Spec.Format),
		}
	}

	return eventAPI
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

	if compassEventAPI.Spec != nil {

		var data []byte
		if compassEventAPI.Spec.Data != nil {
			data = []byte(*compassEventAPI.Spec.Data)
		}

		eventAPI.EventAPISpec = &kymamodel.EventAPISpec{
			Type:   kymamodel.EventAPISpecType(compassEventAPI.Spec.Type),
			Data:   data,
			Format: kymamodel.SpecFormat(compassEventAPI.Spec.Format),
		}
	}

	return eventAPI
}

func convertDocument(compassDoc *graphql.Document) kymamodel.Document {
	var data []byte
	if compassDoc.Data != nil {
		data = []byte(*compassDoc.Data)
	}

	return kymamodel.Document{
		ID:          compassDoc.ID,
		Title:       compassDoc.Title,
		Format:      kymamodel.DocumentFormat(string(compassDoc.Format)),
		Description: compassDoc.Description,
		DisplayName: compassDoc.DisplayName,
		Kind:        compassDoc.Kind,
		Data:        data,
	}
}
