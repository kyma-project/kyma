package jwt

import (
	"testing"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/model"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/jwt/automock"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_ParseAndVerify(t *testing.T) {
	// GIVEN
	testTokenSrc := "token_src"
	testJWKSURI := "https://jwks.uri"
	testJWKSet := jwk.NewSet()
	testKey, _ := jwk.New(`-----BEGIN PUBLIC KEY-----
MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAIIzg7xiwpkyETZ10GEO9Dg3LEt2u5CR
Pl1X0ZVL8n2h3LmuP7n1MXkmNDou7aFFGF8e7VVMcgcH/Fc2CP5prgkCAwEAAQ==
-----END PUBLIC KEY-----`)
	testJWKSet.Add(testKey)
	testToken := jwt.New()
	testError := errors.New("test error")

	testCases := []struct {
		Name           string
		FixJWX         func() *automock.JWX
		ExpectedResult jwt.Token
		ExpectedError  error
	}{
		{
			Name: "Success",
			FixJWX: func() *automock.JWX {
				m := &automock.JWX{}
				m.On("Fetch", mock.Anything, testJWKSURI).Return(testJWKSet, nil).Once()
				m.On("Parse", []byte(testTokenSrc), jwt.WithValidate(true), jwt.WithVerify(alg, testKey)).Return(testToken, nil).Once()
				return m
			},
			ExpectedResult: testToken,
			ExpectedError:  nil,
		},
		{
			Name: "Error when key set empty",
			FixJWX: func() *automock.JWX {
				m := &automock.JWX{}
				m.On("Fetch", mock.Anything, testJWKSURI).Return(jwk.NewSet(), nil).Once()
				return m
			},
			ExpectedResult: nil,
			ExpectedError:  errors.New("JWK key not found in key set"),
		},
		{
			Name: "Error while fetching jwk set",
			FixJWX: func() *automock.JWX {
				m := &automock.JWX{}
				m.On("Fetch", mock.Anything, testJWKSURI).Return(nil, testError).Once()
				return m
			},
			ExpectedResult: nil,
			ExpectedError:  testError,
		},
		{
			Name: "Error parsing token",
			FixJWX: func() *automock.JWX {
				m := &automock.JWX{}
				m.On("Fetch", mock.Anything, testJWKSURI).Return(testJWKSet, nil).Once()
				m.On("Parse", []byte(testTokenSrc), jwt.WithValidate(true), jwt.WithVerify(alg, testKey)).Return(nil, testError).Once()
				return m
			},
			ExpectedResult: nil,
			ExpectedError:  testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			service := NewService()
			mockFetcher := testCase.FixJWX()
			service.jwx = mockFetcher

			// WHEN
			res, err := service.ParseAndVerify(testTokenSrc, testJWKSURI)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, res, testCase.ExpectedResult)
			mock.AssertExpectationsForObjects(t, mockFetcher)
		})
	}
}

func TestService_GetUser(t *testing.T) {
	// GIVEN
	testEmail := "test@example.com"
	service := NewService()

	testCases := []struct {
		Name           string
		Token          func() jwt.Token
		ExpectedResult model.User
		ExpectedError  error
	}{
		{
			Name: "Success when Admin and Dev",
			Token: func() jwt.Token {
				token := jwt.New()
				token.Set(claimEmail, testEmail)
				token.Set(claimScope, []interface{}{
					"unknownScope",
					scopeAdmin,
					1234,
					scopeDeveloper,
					"someOtherScope",
				})
				return token
			},
			ExpectedResult: model.User{
				Email:       testEmail,
				IsDeveloper: true,
				IsAdmin:     true,
			},
			ExpectedError: nil,
		},
		{
			Name: "Success when Admin",
			Token: func() jwt.Token {
				token := jwt.New()
				token.Set(claimEmail, testEmail)
				token.Set(claimScope, []interface{}{
					scopeAdmin,
				})
				return token
			},
			ExpectedResult: model.User{
				Email:       testEmail,
				IsDeveloper: false,
				IsAdmin:     true,
			},
			ExpectedError: nil,
		},
		{
			Name: "Success when Dev",
			Token: func() jwt.Token {
				token := jwt.New()
				token.Set(claimEmail, testEmail)
				token.Set(claimScope, []interface{}{
					scopeDeveloper,
				})
				return token
			},
			ExpectedResult: model.User{
				Email:       testEmail,
				IsDeveloper: true,
				IsAdmin:     false,
			},
			ExpectedError: nil,
		},
		{
			Name: "Success when no known scopes",
			Token: func() jwt.Token {
				token := jwt.New()
				token.Set(claimEmail, testEmail)
				token.Set(claimScope, []interface{}{
					"unknownScope",
				})
				return token
			},
			ExpectedResult: model.User{
				Email:       testEmail,
				IsDeveloper: false,
				IsAdmin:     false,
			},
			ExpectedError: nil,
		},
		{
			Name: "Error when no email claim",
			Token: func() jwt.Token {
				token := jwt.New()
				return token
			},
			ExpectedResult: model.User{},
			ExpectedError:  errors.New("Token does not contain valid email private claim"),
		},
		{
			Name: "Error when no scopes claim",
			Token: func() jwt.Token {
				token := jwt.New()
				token.Set(claimEmail, testEmail)
				return token
			},
			ExpectedResult: model.User{},
			ExpectedError:  errors.New("Token does not contain valid scope private claim"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {

			// WHEN
			res, err := service.GetUser(testCase.Token())

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, res, testCase.ExpectedResult)
		})
	}
}
