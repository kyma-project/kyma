package synchronization

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
)

const (
	specAPIType                   = "API"
	specEventsType                = "Events"
	CredentialsOAuthType          = "OAuth"
	CredentialsBasicType          = "Basic"
	CredentialsCertificateGenType = "CertificateGen"
)

type Converter interface {
	Do(application Application) v1alpha1.Application
}

type converter struct {
}

func (c converter) Do(application Application) v1alpha1.Application {
	return v1alpha1.Application{}
}

func NewConverter() Converter {
	return converter{}
}
