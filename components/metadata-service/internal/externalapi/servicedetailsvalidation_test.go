package externalapi

import (
	"testing"

	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/stretchr/testify/assert"
)

func TestServiceDetailsValidator(t *testing.T) {
	t.Run("should accept service details with API", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})

	t.Run("should accept service details with events", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Events: &Events{
				Spec: eventsRawSpec,
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})

	t.Run("should accept service details with API and events", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
			},
			Events: &Events{
				Spec: eventsRawSpec,
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not accept service details without API and Events", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept service details without name", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Provider:    "provider",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept service details without provider", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept service details without description", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:     "name",
			Provider: "provider",
			Api: &API{
				TargetUrl: "http://target.com",
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})
}

func TestServiceDetailsValidator_API(t *testing.T) {
	t.Run("should not accept API without targetUrl", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api:         &API{},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept API spec other than json object", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
				Spec:      []byte("\"{\\\"wrong_string_json_object\\\":true}\""),
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept API spec with more than 1 type of auth", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
				Credentials: &Credentials{
					Basic: &BasicAuth{
						Username: "username",
						Password: "password",
					},
					Oauth: &Oauth{
						URL:          "http://test.com/token",
						ClientID:     "client",
						ClientSecret: "secret",
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})
}

func TestServiceDetailsValidator_API_OAuth(t *testing.T) {
	t.Run("should accept OAuth credentials", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
				Credentials: &Credentials{
					Oauth: &Oauth{
						URL:          "http://test.com/token",
						ClientID:     "client",
						ClientSecret: "secret",
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not accept OAuth credentials with empty oauth", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
				Credentials: &Credentials{
					Oauth: &Oauth{},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept OAuth credentials with incomplete oauth", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
				Credentials: &Credentials{
					Oauth: &Oauth{
						URL:      "http://test.com/token",
						ClientID: "client",
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept OAuth credentials with wrong oauth url", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
				Credentials: &Credentials{
					Oauth: &Oauth{
						URL:          "test_com/token",
						ClientID:     "client",
						ClientSecret: "secret",
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})
}

func TestServiceDetailsValidator_API_Basic(t *testing.T) {
	t.Run("should accept Basic Auth credentials", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
				Credentials: &Credentials{
					Basic: &BasicAuth{
						Username: "username",
						Password: "password",
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.NoError(t, err)
	})

	t.Run("should not accept Basic Auth credentials with empty basic", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
				Credentials: &Credentials{
					Basic: &BasicAuth{},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})

	t.Run("should not accept Basic Auth credentials with incomplete basic", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Api: &API{
				TargetUrl: "http://target.com",
				Credentials: &Credentials{
					Basic: &BasicAuth{
						Username: "username",
					},
				},
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})
}

func TestServiceDetailsValidator_Events(t *testing.T) {
	t.Run("should not accept events spec other than json object", func(t *testing.T) {
		// given
		serviceDetails := ServiceDetails{
			Name:        "name",
			Provider:    "provider",
			Description: "description",
			Events: &Events{
				Spec: []byte("\"{\\\"wrong_string_json_object\\\":true}\""),
			},
		}

		validator := NewServiceDetailsValidator()

		// when
		err := validator.Validate(serviceDetails)

		// then
		assert.Error(t, err)
		assert.Equal(t, apperrors.CodeWrongInput, err.Code())
	})
}
