package reload_test

import (
	"crypto/tls"
	"errors"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/apiserver-proxy/cmd/proxy/reload"
	"github.com/kyma-project/kyma/components/apiserver-proxy/internal/authn"
	"github.com/stretchr/testify/assert"
	"k8s.io/apiserver/pkg/authentication/authenticator"
)

func TestTLSCertReloaderErrorOnStart(t *testing.T) {

	constructor := func() (*tls.Certificate, error) {
		return nil, errors.New("expected error")
	}

	reloader, err := reload.NewTLSCertReloader(constructor)

	assert.Nil(t, reloader)
	assert.Equal(t, "expected error", err.Error())
}

func TestTLSCertReloaderScenario(t *testing.T) {

	var constructorCalls int
	var fakeCert1 *tls.Certificate
	var fakeCert2 *tls.Certificate

	constructor := func() (*tls.Certificate, error) {
		constructorCalls += 1

		val := new(tls.Certificate)

		if constructorCalls == 1 {
			fakeCert1 = val
		} else if constructorCalls == 2 {
			fakeCert2 = val
		} else {
			t.Fatalf("Invalid constructorCalls")
		}

		return val, nil
	}

	reloader, err := reload.NewTLSCertReloader(constructor)

	assert.Nil(t, err)
	assert.NotNil(t, reloader)
	assert.Equal(t, 1, constructorCalls)
	assert.NotNil(t, fakeCert1)
	assert.Nil(t, fakeCert2)
	actualCert, _ := reloader.GetCertificateFunc(nil)
	assert.Equal(t, fakeCert1, actualCert)

	reloader.Reload()

	assert.Equal(t, 2, constructorCalls)
	assert.NotNil(t, fakeCert1)
	assert.NotNil(t, fakeCert2)
	actualCert, _ = reloader.GetCertificateFunc(nil)
	assert.Equal(t, fakeCert2, actualCert)

	//Assure we actually have two different values!
	assert.True(t, fakeCert1 != fakeCert2, "Certificates should be different!")
}

func TestAuthRequestProviderError(t *testing.T) {

	constructor := func() (authn.CancelableAuthRequest, error) {
		return nil, errors.New("expected error")
	}

	reloader, err := reload.NewCancelableAuthReqestReloader(constructor)

	assert.Nil(t, reloader)
	assert.Equal(t, "expected error", err.Error())
}

func TestAuthRequestReloaderErrorOnStart(t *testing.T) {

	constructor := func() (authn.CancelableAuthRequest, error) {
		return nil, errors.New("expected error")
	}

	reloader, err := reload.NewCancelableAuthReqestReloader(constructor)

	assert.Nil(t, reloader)
	assert.Equal(t, "expected error", err.Error())
}

func TestAuthRequestReloaderScenario(t *testing.T) {

	var constructorCalls int
	var fakeAuthorizer1 *fakeAuthorizer
	var fakeAuthorizer2 *fakeAuthorizer
	var fakeAuthorizer3 *fakeAuthorizer

	constructor := func() (authn.CancelableAuthRequest, error) {
		constructorCalls += 1

		val := new(fakeAuthorizer)

		if constructorCalls == 1 {
			fakeAuthorizer1 = val
		} else if constructorCalls == 2 {
			fakeAuthorizer2 = val
		} else if constructorCalls == 3 {
			fakeAuthorizer3 = val
		} else {
			t.Fatalf("Invalid constructorCalls")
		}

		return val, nil
	}

	assert.Nil(t, fakeAuthorizer1)
	assert.Nil(t, fakeAuthorizer2)
	assert.Nil(t, fakeAuthorizer3)

	reloader, err := reload.NewCancelableAuthReqestReloader(constructor)

	assert.Nil(t, err)
	assert.NotNil(t, reloader)

	assert.Equal(t, 1, constructorCalls)
	assert.NotNil(t, fakeAuthorizer1)
	assert.Nil(t, fakeAuthorizer2)
	assert.Nil(t, fakeAuthorizer3)

	assert.Equal(t, 0, fakeAuthorizer1.cancelFuncCalls) //Not canceled yet

	reloader.Reload() //Creates new and cancel previous

	assert.Equal(t, 2, constructorCalls)
	assert.NotNil(t, fakeAuthorizer1)
	assert.NotNil(t, fakeAuthorizer2)
	assert.Nil(t, fakeAuthorizer3)

	assert.Equal(t, 1, fakeAuthorizer1.cancelFuncCalls)
	assert.Equal(t, 0, fakeAuthorizer2.cancelFuncCalls) //Not canceled yet

	reloader.Reload() //Creates new and cancel previous

	assert.Equal(t, 3, constructorCalls)
	assert.NotNil(t, fakeAuthorizer1)
	assert.NotNil(t, fakeAuthorizer2)
	assert.NotNil(t, fakeAuthorizer3)

	assert.Equal(t, 1, fakeAuthorizer1.cancelFuncCalls)
	assert.Equal(t, 1, fakeAuthorizer2.cancelFuncCalls)
	assert.Equal(t, 0, fakeAuthorizer3.cancelFuncCalls) //Not canceled yet

	//Assure we actually have two different values!
	assert.True(t, fakeAuthorizer1 != fakeAuthorizer2, "Instances should be different")
	assert.True(t, fakeAuthorizer2 != fakeAuthorizer3, "Instances should be different")
	assert.True(t, fakeAuthorizer3 != fakeAuthorizer1, "Instances should be different")
}

type fakeAuthorizer struct {
	cancelFuncCalls int
}

func (fa *fakeAuthorizer) Cancel() {
	fa.cancelFuncCalls += 1
}

func (fa *fakeAuthorizer) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	return nil, false, nil
}
