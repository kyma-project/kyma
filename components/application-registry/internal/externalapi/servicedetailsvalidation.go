package externalapi

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata"

	"github.com/asaskevich/govalidator"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
)

type ServiceDetailsValidator interface {
	Validate(application string, details ServiceDetails) apperrors.AppError
}

type ServiceDetailsValidatorFunc func(application string, details ServiceDetails) apperrors.AppError

func (f ServiceDetailsValidatorFunc) Validate(application string, details ServiceDetails) apperrors.AppError {
	return f(application, details)
}

func NewServiceDetailsValidator(serviceDefinitionService metadata.ServiceDefinitionService) ServiceDetailsValidator {
	return ServiceDetailsValidatorFunc(func(application string, details ServiceDetails) apperrors.AppError {
		_, err := govalidator.ValidateStruct(details)
		if err != nil {
			return apperrors.WrongInput("Incorrect structure of service definition, %s", err.Error())
		}

		if details.Api == nil && details.Events == nil {
			return apperrors.WrongInput(
				"At least one of service definition attributes: 'api', 'events' have to be provided")
		}

		apperr := validateServiceName(serviceDefinitionService, application, details.Name)

		if apperr != nil {
			return apperr
		}

		if details.Api != nil {
			apperr = validateApiSpec(details.Api.Spec)
			if apperr != nil {
				return apperr
			}

			apperr = validateApiCredentials(details.Api.Credentials)
			if apperr != nil {
				return apperr
			}

			apperr = validateSpecificationCredentials(details.Api.SpecificationCredentials)
			if apperr != nil {
				return apperr
			}
		}

		apperr = validateEventsSpec(details.Events)
		if apperr != nil {
			return apperr
		}

		return nil
	})
}

func validateServiceName(serviceDefinitionService metadata.ServiceDefinitionService, application, serviceName string) apperrors.AppError {

	isAlreadyUsed, err := serviceDefinitionService.IsServiceNameUsed(application, serviceName)

	if err != nil {
		return apperrors.WrongInput("Could not validate service name, %s", err.Error())
	}

	if isAlreadyUsed == true {
		return apperrors.WrongInput("Service definition is invalid: Provided service name %s is already used in application %s", serviceName, application)
	}

	return nil
}

func validateApiSpec(spec json.RawMessage) apperrors.AppError {
	if spec != nil {
		err := validateSpec(spec)
		if err != nil {
			return apperrors.WrongInput("api.Spec is not a proper json object, %s", err.Error())
		}
	}

	return nil
}

func validateEventsSpec(events *Events) apperrors.AppError {
	if events != nil && events.Spec != nil {
		err := validateSpec(events.Spec)
		if err != nil {
			return apperrors.WrongInput("events.Spec is not a proper json object, %s", err.Error())
		}
	}

	return nil
}

func validateApiCredentials(credentials *CredentialsWithCSRF) apperrors.AppError {
	if credentials != nil {
		var basic *BasicAuth
		var oauth *Oauth
		var cert *CertificateGen

		if credentials.BasicWithCSRF != nil {
			basic = &credentials.BasicWithCSRF.BasicAuth
		}

		if credentials.OauthWithCSRF != nil {
			oauth = &credentials.OauthWithCSRF.Oauth
		}

		if credentials.CertificateGenWithCSRF != nil {
			cert = &credentials.CertificateGenWithCSRF.CertificateGen
		}

		if validateCredentials(basic, oauth, cert) {
			return apperrors.WrongInput("api.CredentialsWithCSRF is invalid: to many authentication methods provided")
		}
	}

	return nil
}

func validateSpecificationCredentials(credentials *Credentials) apperrors.AppError {
	if credentials != nil {
		basic := credentials.Basic
		oauth := credentials.Oauth

		if validateCredentials(basic, oauth, nil) {
			return apperrors.WrongInput("api.CredentialsWithCSRF is invalid: to many authentication methods provided")
		}
	}

	return nil
}

func validateSpec(rawMessage json.RawMessage) error {
	var m map[string]*json.RawMessage
	return json.Unmarshal(rawMessage, &m)
}

func validateCredentials(basic *BasicAuth, oauth *Oauth, cert *CertificateGen) bool {
	credentialsCount := 0

	if basic != nil {
		credentialsCount++
	}

	if oauth != nil {
		credentialsCount++
	}

	if cert != nil {
		credentialsCount++
	}

	return credentialsCount > 1
}
