package compass

import (
	"errors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization"
	"github.com/sirupsen/logrus"
)

func (app Application) ToApplication() synchronization.Application {

	var apis []synchronization.APIDefinition
	if app.APIs != nil {
		apis = convertAPIs(app.APIs.Data)
	}

	var eventAPIs []synchronization.EventAPIDefinition
	if app.EventAPIs != nil {
		eventAPIs = convertEventAPIs(app.EventAPIs.Data)
	}

	var documents []synchronization.Document
	if app.Documents != nil {
		documents = convertDocuments(app.Documents.Data)
	}

	return synchronization.Application{
		ID:          app.ID,
		Name:        app.Name,
		Description: app.Description,
		Labels:      map[string][]string(app.Labels),
		APIs:        apis,
		EventAPIs:   eventAPIs,
		Documents:   documents,
	}

}

func convertAPIs(compassAPIs []*graphql.APIDefinition) []synchronization.APIDefinition {
	var apis = make([]synchronization.APIDefinition, len(compassAPIs))

	for i, cAPI := range compassAPIs {
		apis[i] = convertAPI(cAPI)
	}

	return apis
}

func convertEventAPIs(compassEventAPIs []*graphql.EventAPIDefinition) []synchronization.EventAPIDefinition {
	var eventAPIs = make([]synchronization.EventAPIDefinition, len(compassEventAPIs))

	for i, cAPI := range compassEventAPIs {
		eventAPIs[i] = convertEventAPI(cAPI)
	}

	return eventAPIs
}

func convertDocuments(compassDocuments []*graphql.Document) []synchronization.Document {
	var documents = make([]synchronization.Document, len(compassDocuments))

	for i, cDoc := range compassDocuments {
		documents[i] = convertDocument(cDoc)
	}

	return documents
}

func convertAPI(compassAPI *graphql.APIDefinition) synchronization.APIDefinition {
	api := synchronization.APIDefinition{
		ID:          compassAPI.ID,
		Name:        compassAPI.Name,
		Description: *compassAPI.Description,
		TargetUrl:   compassAPI.TargetURL,
	}

	if compassAPI.Spec != nil {
		api.APISpec = &synchronization.APISpec{
			Type: synchronization.APISpecType(string(compassAPI.Spec.Type)),
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

func convertAuth(compassAuth *graphql.Auth) (*synchronization.Credentials, error) {
	credentials := &synchronization.Credentials{}

	switch cred := compassAuth.Credential.(type) {
	case graphql.BasicCredentialData:
		credentials.Basic = &synchronization.Basic{
			Username: cred.Username,
			Password: cred.Password,
		}
	case graphql.OAuthCredentialData:
		credentials.Oauth = &synchronization.Oauth{
			URL:          cred.URL,
			ClientID:     cred.ClientID,
			ClientSecret: cred.ClientSecret,
		}
	default:
		return nil, errors.New("invalid credentials type")
	}

	if compassAuth.RequestAuth != nil && compassAuth.RequestAuth.Csrf != nil {
		credentials.CSRFInfo = &synchronization.CSRFInfo{
			TokenEndpointURL: compassAuth.RequestAuth.Csrf.TokenEndpointURL,
		}
	}

	return credentials, nil
}

func convertEventAPI(compassEventAPI *graphql.EventAPIDefinition) synchronization.EventAPIDefinition {
	eventAPI := synchronization.EventAPIDefinition{
		ID:          compassEventAPI.ID,
		Name:        compassEventAPI.Name,
		Description: *compassEventAPI.Description,
	}

	if compassEventAPI.Spec != nil {
		eventAPI.EventAPISpec = &synchronization.EventAPISpec{
			Type: synchronization.EventAPISpecType(string(compassEventAPI.Spec.Type)),
			Data: []byte(*compassEventAPI.Spec.Data),
		}
	}

	return eventAPI
}

func convertDocument(compassDoc *graphql.Document) synchronization.Document {
	var data []byte
	if compassDoc.Data != nil {
		data = []byte(*compassDoc.Data)
	}

	return synchronization.Document{
		ID:          compassDoc.ID,
		Title:       compassDoc.Title,
		Format:      synchronization.DocumentFormat(string(compassDoc.Format)),
		Description: compassDoc.Description,
		DisplayName: compassDoc.DisplayName,
		Kind:        compassDoc.Kind,
		Data:        data,
	}
}
