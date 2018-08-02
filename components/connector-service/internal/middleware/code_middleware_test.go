package middleware

import (
	"testing"
	"github.com/stretchr/testify/require"
	"net/http"
	"bytes"
	"net/http/httptest"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"context"
)

const testHandlerStatus = http.StatusOK
const routeKey = 1

type testHandler struct {
}

func (th testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)  {
	//r.WithContext(context.WithValue())
	w.WriteHeader(testHandlerStatus)
}

func ServeHTTP(w http.ResponseWriter, r *http.Request)  {
	//r.WithContext(context.WithValue())
	w.WriteHeader(testHandlerStatus)
}

func testRouter(path string) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc(path, ServeHTTP)
	//(...)
	return r
}

func TestCodeMiddleware_Handle(t *testing.T) {

	t.Run("should observe status", func(t *testing.T) {
		// given
		url := "http://connector-service-internal-api:8081/v1/remoteenvironments/{reName}/tokens"

		labels := map[string]string{
			"endpoint":url,
			"method":http.MethodPost,
		}

		req, err := http.NewRequest(http.MethodPost, "/v1/remoteenvironments/{reName}/tokens", bytes.NewReader(nil))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": "ec-default"})
		//req = req.WithContext(context.WithValue(req.Context(), routeKey, "/v1/remoteenvironments/ec-default/tokens"))

		codeMiddleware, err := NewCodeMiddleware("code")
		require.NoError(t, err)

		router := mux.NewRouter()
		router.Use(codeMiddleware.Handle)
		router.HandleFunc("/v1/remoteenvironments/ec-default/tokens", ServeHTTP).Methods(http.MethodPost)

		// when
		// codeMiddleware.Handle(testHandler{})
		//handler.ServeHTTP(rr, req)
		//router := testRouter("/v1/remoteenvironments/ec-default/tokens")
		router.ServeHTTP(rr, req)

		// then
		assert.Equal(t, testHandlerStatus, rr.Code)
		summary, err := codeMiddleware.summaryVec.GetMetricWith(labels)
		require.NoError(t, err)
		assert.NotNil(t, summary)
	})

	t.Run("should ", func(t *testing.T) {
		// given
		url := "/v1/remoteenvironments/{reName}/tokens"

		labels := map[string]string{
			"endpoint":url,
			"method":http.MethodPost,
		}

		req, err := http.NewRequest(http.MethodPost, "", bytes.NewReader(nil))
		require.NoError(t, err)
		req.WithContext(context.Background())
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": "ec-default"})

		codeMiddleware, err := NewCodeMiddleware("code")
		require.NoError(t, err)

		// when
		handler := codeMiddleware.Handle(testHandler{})
		handler.ServeHTTP(rr, req)

		// then
		assert.Equal(t, testHandlerStatus, rr.Code)
		summary, err := codeMiddleware.summaryVec.GetMetricWith(labels)
		require.NoError(t, err)
		assert.NotNil(t, summary)
	})
}


/*
	t.Run("should create certificate", func(t *testing.T) {
		// given
		url := fmt.Sprintf("/v1/remoteenvironments/%s/client-cert?token=%s", reName, token)

		tokenCache := &tokensMock.TokenCache{}
		tokenCache.On("Get", reName).Return(token, true)
		tokenCache.On("Delete", reName).Return()

		secretsRepository := &secrectsMock.Repository{}
		secretsRepository.On("Get", authSecretName).Return(caCrtEncoded, caKeyEncoded, nil)

		certUtils := &certMock.CertificateUtility{}
		certUtils.On("LoadCert", caCrtEncoded).Return(caCrt, nil)
		certUtils.On("LoadKey", caKeyEncoded).Return(caKey, nil)
		certUtils.On("LoadCSR", tokenRequest.CSR).Return(csr, nil)

		subjectValues := certificates.CSRSubject{
			CName:              reName,
			Country:            country,
			Organization:       organization,
			OrganizationalUnit: organizationalUnit,
			Locality:           locality,
			Province:           province,
		}
		certUtils.On("CheckCSRValues", csr, subjectValues).Return(nil)
		certUtils.On("SignWithCA", caCrt, csr, caKey).Return(crtBase64, nil)

		registrationHandler := createSignatureHandler(tokenCache, certUtils, secretsRepository)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		req = mux.SetURLVars(req, map[string]string{"reName": reName})

		// when
		registrationHandler.SignCSR(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var certResponse certResponse
		err = json.Unmarshal(responseBody, &certResponse)
		require.NoError(t, err)

		assert.Equal(t, crtBase64, certResponse.CRT)
		assert.Equal(t, http.StatusCreated, rr.Code)
	})

 */