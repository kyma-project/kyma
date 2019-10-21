package reload

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/apiserver-proxy/internal/authn"
	"k8s.io/apiserver/pkg/authentication/authenticator"
)

//ReloadNotifier is used to sent notifications about events requiring refreshing data
type ReloadNotifier interface {
	//Registers callback handler that is called when an event requiring refreshing data occurs
	RegisterCallback(handler func())
}

//TLSCertConstructor knows how to construct a tls.Certificate instance
type TLSCertConstructor func() (*tls.Certificate, error)

//ReloadableTLSCertProvider enables to create and re-create an instance of tls.Certificate in a thread-safe way.
//It's GetCertificateFunc conforms to tls.Config.GetCertificate function type.
type ReloadableTLSCertProvider struct {
	constructor TLSCertConstructor
	holder      *TLSCrtKeyPairHolder
}

//NewReloadableTLSCertProvider creates a new instance of ReloadableTLSCertProvider.
//notifier parameter is used register a data reloading callback.
//External code can make use of this callback to trigger data reloading from outside.
//It's safe to trigger reloading from other goroutines.
func NewReloadableTLSCertProvider(constructor TLSCertConstructor, notifier ReloadNotifier) (*ReloadableTLSCertProvider, error) {

	result := &ReloadableTLSCertProvider{
		constructor: constructor,
		holder:      NewTLSCrtKeyPairHolder(),
	}

	//Initial read
	err := result.reload()
	if err != nil {
		return nil, err
	}

	//Used by external notifier to trigger certificate reloads
	onDataChangeFunc := func() {
		err := result.reload()
		if err != nil {
			glog.Errorf("Failed to reload certificate: %v", err)
		}
	}
	notifier.RegisterCallback(onDataChangeFunc)

	return result, nil
}

//reloads the internal instance using provided constructor function
//Note: It must NOT modify the existing value in case of an error!
func (ckpr *ReloadableTLSCertProvider) reload() error {
	newCert, err := ckpr.constructor()
	if err != nil {
		return err
	}
	ckpr.holder.Set(newCert)
	return nil
}

//GetCertificateFunc conforms to tls.Config.GetCertificate function type
func (ckpr *ReloadableTLSCertProvider) GetCertificateFunc(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	fmt.Println("getCertificateFunc")
	return ckpr.holder.Get(), nil
}

//TLSCrtKeyPairHolder keeps a tls.Certificate instance and allows for Get/Set operations in a thread-safe way
type TLSCrtKeyPairHolder struct {
	rwmu  sync.RWMutex
	value *tls.Certificate
}

//NewTLSCrtKeyPairHolder returns new TLSCrtKeyPairHolder instance
func NewTLSCrtKeyPairHolder() *TLSCrtKeyPairHolder {
	return &TLSCrtKeyPairHolder{}
}

//Get returns the tls.Certificate instance stored in the TLSCrtKeyPairHolder
func (tlsh *TLSCrtKeyPairHolder) Get() *tls.Certificate {
	tlsh.rwmu.RLock()
	defer tlsh.rwmu.RUnlock()
	return tlsh.value
}

//Set stores given tls.Certificate in the TLSCrtKeyPairHolder
func (tlsh *TLSCrtKeyPairHolder) Set(v *tls.Certificate) {
	tlsh.rwmu.Lock()
	defer tlsh.rwmu.Unlock()
	tlsh.value = v
}

//AuthReqConstructor knows how to construct an authenticator.Request instance
//It also returns an instance of authn.AuthenticatorCancelFunc that allows to cancel the authenticator.
type AuthReqConstructor func() (authenticator.Request, authn.AuthenticatorCancelFunc, error)

//ReloadableAuthReq enables to create and re-create an instance of authenticator.Request in a thread-safe way.
//It's used to re-create authenticator.Request instance every time a change in oidc-ca-file is detected.
//It implements authenticator.Request interface so it can be easily plugged in instead of a "real" instance.
type ReloadableAuthReq struct {
	constructor AuthReqConstructor
	holder      *AuthReqHolder
}

//NewReloadableAuthReq creates a new instance of ReloadableAuthReq.
//notifier parameter allows to control when the instance is re-created from outside.
//It's safe to trigger re-creation from other goroutines.
func NewReloadableAuthReq(constructor AuthReqConstructor, notifier ReloadNotifier) (*ReloadableAuthReq, error) {
	result := ReloadableAuthReq{
		constructor: constructor,
		holder:      NewAuthReqHolder(),
	}

	//Initial read
	err := result.reload()
	if err != nil {
		return nil, err
	}

	onDataChangeFunc := func() {
		err := result.reload()
		if err != nil {
			glog.Errorf("Failed to reload authenticator.Request: %v", err)
		}
	}
	notifier.RegisterCallback(onDataChangeFunc)

	return &result, nil
}

//reloads the internal instance using provided constructor function
//Note: It must NOT modify the existing value in case of an error!
func (rar *ReloadableAuthReq) reload() error {
	newAuthReq, newAuthCancelFunc, err := rar.constructor()
	if err != nil {
		return err
	}

	_, oldAuthCancelFunc := rar.holder.Get()
	if oldAuthCancelFunc != nil {
		glog.Info("Cancelling previous OIDC Authenticator instance")
		oldAuthCancelFunc()
	}

	rar.holder.Set(newAuthReq, newAuthCancelFunc)
	return nil
}

//AuthenticateRequest implements authenticator.Request interface
func (rar *ReloadableAuthReq) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	//Delegate to internally-stored instance (thread-safe)
	actual, _ := rar.holder.Get()
	return actual.AuthenticateRequest(req)
}

//AuthReqHolder keeps an authenticator.Request and an authn.AuthenticatorCancelFunc instance.
//It allows for Get/Set operations in a thread-safe way
type AuthReqHolder struct {
	rwmu       sync.RWMutex
	value      authenticator.Request
	cancelFunc authn.AuthenticatorCancelFunc
}

//NewAuthReqHolder returns new AuthReqHolder instance
func NewAuthReqHolder() *AuthReqHolder {
	return &AuthReqHolder{}
}

//Get returns the instances stored in the AuthReqHolder
func (arh *AuthReqHolder) Get() (authenticator.Request, authn.AuthenticatorCancelFunc) {
	arh.rwmu.RLock()
	defer arh.rwmu.RUnlock()
	return arh.value, arh.cancelFunc
}

//Set stores given instances in the AuthReqHolder
func (arh *AuthReqHolder) Set(v authenticator.Request, c authn.AuthenticatorCancelFunc) {
	arh.rwmu.Lock()
	defer arh.rwmu.Unlock()
	arh.value = v
	arh.cancelFunc = c
}
