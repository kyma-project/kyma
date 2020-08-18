//TODO: It's better to test from outside - how much work would be to put this in the endpoints_test package?
package endpoints

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/coreos/go-oidc"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	hydraAPI "github.com/ory/hydra-client-go/models"
	"golang.org/x/oauth2"
)

var _ = Describe("Login Endpoint", func() {

	Describe("Should accept login requests", func() {
		It("In a happy path scenario", func() {

			//given
			var redirectToCallback = "https://login-consent.kyma.local/callback"
			var skip = false
			var hydraClientMock = &mockHydraLoginConsentClient{
				//response from "GetLoginRequest"f
				getLoginRequestResValue: &hydraAPI.LoginRequest{
					Skip: &skip,
				},
			}

			//authenticator, err := NewAuthenticator("https://dex.kyma.local", "clientID", "clientSecret", "https://login-consent.kyma.local/callback", []string{"email", "openid", "profile", "groups"})
			mts := mockTokenSupport{
				clientIDResValue:    "someClientID",
				authCodeURLResValue: redirectToCallback,
			}

			authenticator := &Authenticator{
				tokenSupport: &mts,
				ctx:          context.Background(),
			}

			var cfg = Config{
				hydraClient:   hydraClientMock,
				authenticator: authenticator,
			}

			//when
			req, err := http.NewRequest("GET", "/login?login_challenge=3456", nil)
			if err != nil {
				Fail(err.Error())
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(cfg.Login)
			handler.ServeHTTP(rr, req)

			//then
			//Assert hydra endpoint "GetLoginRequest" was invoked with correct argments
			Expect(hydraClientMock.getLoginRequestArgChallenge).To(Equal("3456"))
			//Hydra client responds with a Skip==false value
			Expect(*hydraClientMock.getLoginRequestResValue.Skip).To(Equal(false))

			//Assert we return correct response to the user
			Expect(rr.Code).To(Equal(http.StatusFound))
			Expect(rr.Body.String()).To(ContainSubstring(redirectToCallback))
			//Assert "Location" header is properly set (we have a redirect)
			loc, err := rr.Result().Location()
			Expect(err).To(BeNil())
			Expect(loc).ToNot(BeNil())
			Expect(loc.String()).To(Equal(redirectToCallback))
		})
	})

	Describe("Should reject login requests", func() {
		It("when user doesn't provide login_challenge", func() {

			//given
			var redirectTo = "http://redirect.to.local:8080"
			var hydraClientMock = &mockHydraLoginConsentClient{
				//response from "RejectLoginRequest"
				rejectLoginRequestResValue: &hydraAPI.CompletedRequest{
					RedirectTo: &redirectTo,
				},
			}

			var cfg = Config{
				hydraClient: hydraClientMock,
			}

			//when
			req, err := http.NewRequest("GET", "/login?login_challenge=", nil)
			if err != nil {
				Fail(err.Error())
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(cfg.Login)
			handler.ServeHTTP(rr, req)

			//then
			//Assert hydra endpoint RejectLoginRequest was invoked with correct arguments
			Expect(hydraClientMock.rejectLoginRequestArgChallenge).To(Equal(""))
			Expect(hydraClientMock.rejectLoginRequestArgBody).To(Equal(
				&hydraAPI.RejectRequest{
					Error:      "login_challenge not found",
					StatusCode: http.StatusBadRequest,
				}))
			//Assert we return correct response to the user
			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			Expect(rr.Body.String()).To(ContainSubstring(redirectTo))
			//NOTE: "Location" header is not relevant for 4xx status codes
		})

		It("when hydra client returns error on getting login request", func() {

			//given
			var redirectTo = "http://redirect.to.local:8081"
			var hydraClientMock = &mockHydraLoginConsentClient{
				//error on "GetLoginRequest"
				getLoginRequestResError: errors.New("getLoginRequestError"),

				//response from "RejectLoginRequest"
				rejectLoginRequestResValue: &hydraAPI.CompletedRequest{
					RedirectTo: &redirectTo,
				},
			}

			var cfg = Config{
				hydraClient: hydraClientMock,
			}

			//when
			req, err := http.NewRequest("GET", "/login?login_challenge=1234", nil)
			if err != nil {
				Fail(err.Error())
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(cfg.Login)
			handler.ServeHTTP(rr, req)

			//then
			//Assert hydra endpoint "GetLoginRequest" was invoked with correct argments
			Expect(hydraClientMock.getLoginRequestArgChallenge).To(Equal("1234"))
			//Hydra client responds with error
			Expect(hydraClientMock.getLoginRequestResError.Error()).To(Equal("getLoginRequestError"))
			//Assert hydra client "RejectLoginRequest" arguments
			Expect(hydraClientMock.rejectLoginRequestArgChallenge).To(Equal("1234"))
			Expect(hydraClientMock.rejectLoginRequestArgBody).To(Equal(
				&hydraAPI.RejectRequest{
					Error:      "getLoginRequestError",
					StatusCode: http.StatusBadRequest,
				}))
			//Assert we return correct response to the user
			Expect(rr.Code).To(Equal(http.StatusBadRequest))
			Expect(rr.Body.String()).To(ContainSubstring(redirectTo))
			//NOTE: "Location" header is not relevant for 4xx status codes
		})

		It("when hydra client returns error on accepting login request", func() {

			//given
			var redirectTo = "http://redirect.to.local:8083"
			var skip = true
			var hydraClientMock = &mockHydraLoginConsentClient{
				//response from "GetLoginRequest"f
				getLoginRequestResValue: &hydraAPI.LoginRequest{
					Skip: &skip,
				},

				//error on "AcceptLoginRequest"
				acceptLoginRequestResError: errors.New("acceptLoginRequestError"),

				//response from "RejectLoginRequest"
				rejectLoginRequestResValue: &hydraAPI.CompletedRequest{
					RedirectTo: &redirectTo,
				},
			}

			var cfg = Config{
				hydraClient: hydraClientMock,
			}

			//when
			req, err := http.NewRequest("GET", "/login?login_challenge=2345", nil)
			if err != nil {
				Fail(err.Error())
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(cfg.Login)
			handler.ServeHTTP(rr, req)

			//then
			//Assert hydra endpoint "GetLoginRequest" was invoked with correct argments
			Expect(hydraClientMock.getLoginRequestArgChallenge).To(Equal("2345"))
			//Hydra client responds with a Skip==true value
			Expect(*hydraClientMock.getLoginRequestResValue.Skip).To(Equal(true))

			//Assert hydra client "AcceptLoginRequest" arguments
			Expect(hydraClientMock.acceptLoginRequestArgChallenge).To(Equal("2345"))
			//Assert hydra client responds with an error
			Expect(hydraClientMock.acceptLoginRequestResError.Error()).To(Equal("acceptLoginRequestError"))

			//Assert hydra client "RejectLoginRequest" arguments
			Expect(hydraClientMock.rejectLoginRequestArgChallenge).To(Equal("2345"))
			Expect(hydraClientMock.rejectLoginRequestArgBody).To(Equal(
				&hydraAPI.RejectRequest{
					Error:      "acceptLoginRequestError",
					StatusCode: http.StatusBadRequest,
				}))
			//Assert we return correct response to the user
			Expect(rr.Code).To(Equal(http.StatusInternalServerError))
			Expect(rr.Body.String()).To(ContainSubstring(redirectTo))
			//NOTE: "Location" header is not relevant for 5xx status codes
		})
	})
})

//Simple mock for hydra client
//Fetches actual function arguments and allows to setup return values
type mockHydraLoginConsentClient struct {
	getLoginRequestArgChallenge string
	getLoginRequestResValue     *hydraAPI.LoginRequest
	getLoginRequestResError     error

	acceptLoginRequestArgChallenge string
	acceptLoginRequestResValue     *hydraAPI.CompletedRequest
	acceptLoginRequestResError     error

	rejectLoginRequestArgChallenge string
	rejectLoginRequestArgBody      *hydraAPI.RejectRequest
	rejectLoginRequestResValue     *hydraAPI.CompletedRequest
	rejectLoginRequestResError     error

	getConsentRequestArgChallenge string
	getConsentRequestResValue     *hydraAPI.ConsentRequest
	getConsentRequestResError     error

	acceptConsentRequestArgChallenge string
	acceptConsentRequestResValue     *hydraAPI.CompletedRequest
	acceptConsentRequestResError     error
}

func (m *mockHydraLoginConsentClient) GetLoginRequest(challenge string) (*hydraAPI.LoginRequest, error) {
	m.getLoginRequestArgChallenge = challenge
	return m.getLoginRequestResValue, m.getLoginRequestResError
}

func (m *mockHydraLoginConsentClient) AcceptLoginRequest(challenge string, body *hydraAPI.AcceptLoginRequest) (*hydraAPI.CompletedRequest, error) {
	m.acceptLoginRequestArgChallenge = challenge
	return m.acceptLoginRequestResValue, m.acceptLoginRequestResError
}

func (m *mockHydraLoginConsentClient) RejectLoginRequest(challenge string, body *hydraAPI.RejectRequest) (*hydraAPI.CompletedRequest, error) {
	m.rejectLoginRequestArgChallenge = challenge
	m.rejectLoginRequestArgBody = body
	return m.rejectLoginRequestResValue, m.rejectLoginRequestResError
}

func (m *mockHydraLoginConsentClient) GetConsentRequest(challenge string) (*hydraAPI.ConsentRequest, error) {
	m.getConsentRequestArgChallenge = challenge
	return m.getConsentRequestResValue, m.getConsentRequestResError
}

func (m *mockHydraLoginConsentClient) AcceptConsentRequest(challenge string, body *hydraAPI.AcceptConsentRequest) (*hydraAPI.CompletedRequest, error) {
	m.acceptConsentRequestArgChallenge = challenge
	return m.acceptConsentRequestResValue, m.acceptConsentRequestResError
}

//Simple mock for endpoints.TokenSupport interface
type mockTokenSupport struct {
	clientIDResValue    string
	authCodeURLResValue string
}

func (mts *mockTokenSupport) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return mts.authCodeURLResValue
}

func (mts *mockTokenSupport) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return nil, nil
}

func (mts *mockTokenSupport) ClientID() string {
	return mts.clientIDResValue
}

func (mts *mockTokenSupport) Verify(config *oidc.Config, ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return nil, nil
}
