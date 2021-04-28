package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/icza/session"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/app/automock"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/model"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/uaa"

	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApp_HandleXSUAAMigrate(t *testing.T) {
	// GIVEN
	testAuthorizationURL := "https://test.endpoint"
	testAuthorizationURLWithParams := "https://test.endpoint/?param=set"
	testError := errors.New("test error")

	testCases := []struct {
		Name                   string
		ExpectedResponseStatus int
		ExpectedLocation       string
		FixUAAClient           func() *automock.UAAClient
	}{
		{
			Name:                   "Success",
			ExpectedResponseStatus: http.StatusFound,
			ExpectedLocation:       testAuthorizationURLWithParams,
			FixUAAClient: func() *automock.UAAClient {
				mockClient := &automock.UAAClient{}
				mockClient.On("GetAuthorizationEndpointWithParams", testAuthorizationURL, mock.Anything).
					Return(testAuthorizationURLWithParams, nil).Once()
				return mockClient
			},
		},
		{
			Name:                   "Error while getting authorization endpoint",
			ExpectedResponseStatus: http.StatusInternalServerError,
			FixUAAClient: func() *automock.UAAClient {
				mockClient := &automock.UAAClient{}
				mockClient.On("GetAuthorizationEndpointWithParams", testAuthorizationURL, mock.Anything).
					Return(testAuthorizationURLWithParams, testError).Once()
				return mockClient
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/xsuaa-migrate", nil)
			w := httptest.NewRecorder()

			app := App{
				uaaOIDConfig: uaa.OpenIDConfiguration{
					AuthorizationEndpoint: testAuthorizationURL,
				},
			}
			mockUAAClient := testCase.FixUAAClient()
			app.uaaClient = mockUAAClient
			handler := http.HandlerFunc(app.HandleXSUAAMigrate)

			// WHEN
			handler.ServeHTTP(w, r)
			res := w.Result()

			// THEN
			assert.Equal(t, testCase.ExpectedResponseStatus, res.StatusCode)
			assert.Equal(t, testCase.ExpectedLocation, res.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, mockUAAClient)
		})
	}
}

func TestHandleXSUAACallback(t *testing.T) {
	// GIVEN
	testDomain := "test.domain"
	testTokenEndpoint := "https://token.endpoint"
	testQueryCode := "code"
	testOAuthState := "state"
	testAccessToken := "token"
	testJWKSURI := "https://jwks.uri"
	testParsedToken := jwt.New()
	testUser := model.User{
		Email:       "user@example.com",
		IsDeveloper: true,
		IsAdmin:     false,
	}
	testError := errors.New("test error")

	testCases := []struct {
		Name                   string
		ExpectedResponseStatus int
		ExpectedLocation       string
		FixUAAClient           func() *automock.UAAClient
		FixJWTService          func() *automock.JWTService
		FixK8SClient           func() *automock.K8sClient
	}{
		{
			Name:                   "Success",
			ExpectedResponseStatus: http.StatusFound,
			ExpectedLocation:       fmt.Sprintf("https://console.%s/info/success.html", testDomain),
			FixUAAClient: func() *automock.UAAClient {
				mockUAAClient := &automock.UAAClient{}
				mockUAAClient.On("GetToken", testTokenEndpoint, testQueryCode).Return(map[string]interface{}{
					"access_token": testAccessToken,
				}, nil).Once()
				return mockUAAClient
			},
			FixJWTService: func() *automock.JWTService {
				mockJWTService := &automock.JWTService{}
				mockJWTService.On("ParseAndVerify", testAccessToken, testJWKSURI).Return(testParsedToken, nil).Once()
				mockJWTService.On("GetUser", testParsedToken).Return(testUser, nil).Once()
				return mockJWTService
			},
			FixK8SClient: func() *automock.K8sClient {
				mockK8SClient := &automock.K8sClient{}
				mockK8SClient.On("EnsureUserPermissions", testUser).Return(nil).Once()
				return mockK8SClient
			},
		},
		{
			Name:                   "Error while getting token",
			ExpectedResponseStatus: http.StatusInternalServerError,
			FixUAAClient: func() *automock.UAAClient {
				mockUAAClient := &automock.UAAClient{}
				mockUAAClient.On("GetToken", testTokenEndpoint, testQueryCode).Return(nil, testError).Once()
				return mockUAAClient
			},
			FixJWTService: func() *automock.JWTService {
				mockJWTService := &automock.JWTService{}
				return mockJWTService
			},
			FixK8SClient: func() *automock.K8sClient {
				mockK8SClient := &automock.K8sClient{}
				return mockK8SClient
			},
		},
		{
			Name:                   "Error when invalid token",
			ExpectedResponseStatus: http.StatusInternalServerError,
			FixUAAClient: func() *automock.UAAClient {
				mockUAAClient := &automock.UAAClient{}
				mockUAAClient.On("GetToken", testTokenEndpoint, testQueryCode).Return(map[string]interface{}{
					"access_token": 0,
				}, nil).Once()
				return mockUAAClient
			},
			FixJWTService: func() *automock.JWTService {
				mockJWTService := &automock.JWTService{}
				return mockJWTService
			},
			FixK8SClient: func() *automock.K8sClient {
				mockK8SClient := &automock.K8sClient{}
				return mockK8SClient
			},
		},
		{
			Name:                   "Error while parsing token",
			ExpectedResponseStatus: http.StatusInternalServerError,
			FixUAAClient: func() *automock.UAAClient {
				mockUAAClient := &automock.UAAClient{}
				mockUAAClient.On("GetToken", testTokenEndpoint, testQueryCode).Return(map[string]interface{}{
					"access_token": testAccessToken,
				}, nil).Once()
				return mockUAAClient
			},
			FixJWTService: func() *automock.JWTService {
				mockJWTService := &automock.JWTService{}
				mockJWTService.On("ParseAndVerify", testAccessToken, testJWKSURI).Return(nil, testError).Once()
				return mockJWTService
			},
			FixK8SClient: func() *automock.K8sClient {
				mockK8SClient := &automock.K8sClient{}
				return mockK8SClient
			},
		},
		{
			Name:                   "Error while getting user from token",
			ExpectedResponseStatus: http.StatusInternalServerError,
			FixUAAClient: func() *automock.UAAClient {
				mockUAAClient := &automock.UAAClient{}
				mockUAAClient.On("GetToken", testTokenEndpoint, testQueryCode).Return(map[string]interface{}{
					"access_token": testAccessToken,
				}, nil).Once()
				return mockUAAClient
			},
			FixJWTService: func() *automock.JWTService {
				mockJWTService := &automock.JWTService{}
				mockJWTService.On("ParseAndVerify", testAccessToken, testJWKSURI).Return(testParsedToken, nil).Once()
				mockJWTService.On("GetUser", testParsedToken).Return(testUser, testError).Once()
				return mockJWTService
			},
			FixK8SClient: func() *automock.K8sClient {
				mockK8SClient := &automock.K8sClient{}
				return mockK8SClient
			},
		},
		{
			Name:                   "Error while creating role bindings for user",
			ExpectedResponseStatus: http.StatusInternalServerError,
			FixUAAClient: func() *automock.UAAClient {
				mockUAAClient := &automock.UAAClient{}
				mockUAAClient.On("GetToken", testTokenEndpoint, testQueryCode).Return(map[string]interface{}{
					"access_token": testAccessToken,
				}, nil).Once()
				return mockUAAClient
			},
			FixJWTService: func() *automock.JWTService {
				mockJWTService := &automock.JWTService{}
				mockJWTService.On("ParseAndVerify", testAccessToken, testJWKSURI).Return(testParsedToken, nil).Once()
				mockJWTService.On("GetUser", testParsedToken).Return(testUser, nil).Once()
				return mockJWTService
			},
			FixK8SClient: func() *automock.K8sClient {
				mockK8SClient := &automock.K8sClient{}
				mockK8SClient.On("EnsureUserPermissions", testUser).Return(testError).Once()
				return mockK8SClient
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", fmt.Sprintf("https://console.%s/callback?code=%s&state=%s", testDomain, testQueryCode, testOAuthState), nil)
			w := httptest.NewRecorder()
			sess := session.NewSessionOptions(&session.SessOptions{
				Attrs: map[string]interface{}{
					oauthStateSessionAttribute: testOAuthState,
				},
			})
			session.Add(sess, w)
			r.AddCookie(&http.Cookie{
				Name:  "sessid",
				Value: sess.ID(),
			})

			app := App{
				uaaOIDConfig: uaa.OpenIDConfiguration{
					TokenEndpoint: testTokenEndpoint,
					JWKSURI:       testJWKSURI,
				},
			}
			mockUAAClient := testCase.FixUAAClient()
			app.uaaClient = mockUAAClient
			mockJWTService := testCase.FixJWTService()
			app.jwtService = mockJWTService
			mockK8SClient := testCase.FixK8SClient()
			app.k8sClient = mockK8SClient
			handler := http.HandlerFunc(app.HandleXSUAACallback)

			// WHEN
			handler.ServeHTTP(w, r)
			res := w.Result()

			// THEN
			assert.Equal(t, testCase.ExpectedResponseStatus, res.StatusCode)
			assert.Equal(t, testCase.ExpectedLocation, res.Header.Get("Location"))
			mock.AssertExpectationsForObjects(t, mockUAAClient, mockJWTService, mockK8SClient)
		})
	}

	t.Run("Success when accessing dex subdomain", func(t *testing.T) {
		// GIVEN
		testRequestURL := fmt.Sprintf("https://dex.%s/callback?code=%s", testDomain, testQueryCode)
		testExpectedURL := fmt.Sprintf("https://console.%s/callback?code=%s", testDomain, testQueryCode)
		r, _ := http.NewRequest("GET", testRequestURL, nil)
		w := httptest.NewRecorder()

		app := App{}
		handler := http.HandlerFunc(app.HandleXSUAACallback)

		// WHEN
		handler.ServeHTTP(w, r)
		res := w.Result()

		// THEN
		assert.Equal(t, http.StatusFound, res.StatusCode)
		assert.Equal(t, testExpectedURL, res.Header.Get("Location"))
	})
}
