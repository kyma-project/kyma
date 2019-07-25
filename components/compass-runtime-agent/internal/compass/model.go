package compass

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ApplicationsForRuntimeResponse struct {
	Result *ApplicationPage `json:"result"`
}

type ApplicationPage struct {
	Data       []*Application    `json:"data"`
	PageInfo   *graphql.PageInfo `json:"pageInfo"`
	TotalCount int               `json:"totalCount"`
}

type Application struct {
	ID             string                          `json:"id"`
	Name           string                          `json:"name"`
	Description    *string                         `json:"description"`
	Labels         Labels                          `json:"labels"`
	Status         *graphql.ApplicationStatus      `json:"status"`
	Webhooks       []*graphql.Webhook              `json:"webhooks"`
	APIs           *graphql.APIDefinitionPage      `json:"apis"`
	EventAPIs      *graphql.EventAPIDefinitionPage `json:"eventAPIs"`
	Documents      *graphql.DocumentPage           `json:"documents"`
	HealthCheckURL *string                         `json:"healthCheckURL"`
}

type ApplicationData struct {
	ID             string
	Name           string
	Description    *string
	Labels         Labels
	Webhooks       []*graphql.Webhook
	APIs           []graphql.APIDefinition
	EventAPIs      []graphql.EventAPIDefinition
	Documents      []graphql.Document
	HealthCheckURL *string
}

type Labels map[string][]string

func (app Application) ToApplication() synchronization.Application {

	return synchronization.Application{
		ID:          app.ID,
		Name:        app.Name,
		Description: app.Description,
		Labels:      map[string][]string(app.Labels),
		APIs:        convertAPIs(app.APIs.Data),
		EventAPIs:   convertEventAPIs(app.EventAPIs.Data),
		Documents:   convertDocuments(app.Documents.Data),
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

// TODO - do we need api.APIType if we have api.APISpec.Type
// TODO - event api spec type

// TODO - move to separate file
// TODO - heavy tests
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
			ClientID:     cred.ClientID,
			ClientSecret: cred.ClientSecret,
		}
	default:
		return nil, errors.New("Invalid credentials type!")
	}

	credentials.CSRFInfo = &synchronization.CSRFInfo{
		TokenEndpointURL: compassAuth.RequestAuth.Csrf.TokenEndpointURL,
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
		// TODO - event api spec type
		//eventAPI.APISpec = &synchronization.APISpec{
		//	Type: synchronization.EventAPISpecType(string(compassEventAPI.Spec.Type)),
		//	Data: []byte(*compassAPI.Spec.Data),
		//}
	}

	return eventAPI
}

func convertDocument(compassDoc *graphql.Document) synchronization.Document {
	return synchronization.Document{
		ID:          compassDoc.ID,
		Title:       compassDoc.Title,
		Format:      synchronization.DocumentFormat(string(compassDoc.Format)),
		Description: compassDoc.Description,
		DisplayName: compassDoc.DisplayName,
		Kind:        compassDoc.Kind,
		Data:        []byte(*compassDoc.Data),
	}
}
