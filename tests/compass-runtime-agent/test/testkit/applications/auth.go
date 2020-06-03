package applications

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

func NewAuth() *AuthInput {
	auth := AuthInput(graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{},
	})
	return &auth
}

func (in *AuthInput) ToCompassInput() *graphql.AuthInput {
	auth := graphql.AuthInput(*in)
	return &auth
}

func (in *AuthInput) WithOAuth(clientId, clientSecret, url string) *AuthInput {
	in.Credential = &graphql.CredentialDataInput{
		Oauth: &graphql.OAuthCredentialDataInput{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			URL:          url,
		},
	}

	return in
}

func (in *AuthInput) WithBasicAuth(username, password string) *AuthInput {
	in.Credential = &graphql.CredentialDataInput{
		Basic: &graphql.BasicCredentialDataInput{
			Username: username,
			Password: password,
		},
	}

	return in
}

func (in *AuthInput) WithHeaders(headers map[string][]string) *AuthInput {
	cHeaders := graphql.HttpHeaders(headers)
	in.AdditionalHeaders = &cHeaders

	return in
}

func (in *AuthInput) WithQueryParams(queryParams map[string][]string) *AuthInput {
	cQuery := graphql.QueryParams(queryParams)
	in.AdditionalQueryParams = &cQuery

	return in
}

func (in *AuthInput) WithCSRF(tokenURL string) *AuthInput {
	in.RequestAuth = &graphql.CredentialRequestAuthInput{
		Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{
			TokenEndpointURL: tokenURL,
			Credential:       in.Credential,
		},
	}

	return in
}
