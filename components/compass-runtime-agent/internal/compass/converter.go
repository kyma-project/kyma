package compass

import (
	"errors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	kymamodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/sirupsen/logrus"
)

func (app Application) ToApplication() kymamodel.Application {

	var apis []kymamodel.APIDefinition
	if app.APIs != nil {
		apis = convertAPIs(app.APIs.Data)
	}

	var eventAPIs []kymamodel.EventAPIDefinition
	if app.EventAPIs != nil {
		eventAPIs = convertEventAPIs(app.EventAPIs.Data)
	}

	var documents []kymamodel.Document
	if app.Documents != nil {
		documents = convertDocuments(app.Documents.Data)
	}

	description := ""
	if app.Description != nil {
		description = *app.Description
	}

	return kymamodel.Application{
		ID:          app.ID,
		Name:        app.Name,
		Description: description,
		Labels:      map[string][]string(app.Labels),
		APIs:        apis,
		EventAPIs:   eventAPIs,
		Documents:   documents,
	}

}

func convertAPIs(compassAPIs []*graphql.APIDefinition) []kymamodel.APIDefinition {
	var apis = make([]kymamodel.APIDefinition, len(compassAPIs))

	for i, cAPI := range compassAPIs {
		apis[i] = convertAPI(cAPI)
	}

	return apis
}

func convertEventAPIs(compassEventAPIs []*graphql.EventAPIDefinition) []kymamodel.EventAPIDefinition {
	var eventAPIs = make([]kymamodel.EventAPIDefinition, len(compassEventAPIs))

	for i, cAPI := range compassEventAPIs {
		eventAPIs[i] = convertEventAPI(cAPI)
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

func convertAPI(compassAPI *graphql.APIDefinition) kymamodel.APIDefinition {
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
		api.APISpec = &kymamodel.APISpec{
			Type: kymamodel.APISpecType(string(compassAPI.Spec.Type)),
			Data: []byte(*compassAPI.Spec.Data),
		}
	}

	// TODO: implement RequestParameters after update in Compass

	if compassAPI.Auth != nil {
		credentials, err := convertAuth(compassAPI.Auth.Auth)
		if err != nil {
			logrus.Errorf("Failed to convert Compass Authentication to credentials: %s", err.Error())
		} else {
			api.Credentials = credentials
		}
	}

	return api
}

func convertAuth(compassAuth *graphql.Auth) (*kymamodel.Credentials, error) {
	credentials := &kymamodel.Credentials{}

	switch cred := compassAuth.Credential.(type) {
	case graphql.BasicCredentialData:
		credentials.Basic = &kymamodel.Basic{
			Username: cred.Username,
			Password: cred.Password,
		}
	case graphql.OAuthCredentialData:
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

func convertEventAPI(compassEventAPI *graphql.EventAPIDefinition) kymamodel.EventAPIDefinition {
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
		eventAPI.EventAPISpec = &kymamodel.EventAPISpec{
			Type: kymamodel.EventAPISpecType(string(compassEventAPI.Spec.Type)),
			Data: []byte(*compassEventAPI.Spec.Data),
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
