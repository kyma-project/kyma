package reload

import (
	"crypto/tls"
	"net/http"
	"sync"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/apiserver-proxy/internal/authn"
	"k8s.io/apiserver/pkg/authentication/authenticator"
)

//TLSCertConstructor knows how to construct a tls.Certificate instance
type TLSCertConstructor func() (*tls.Certificate, error)

//ReloadableTLSCertProvider enables to create and re-create an instance of tls.Certificate in a thread-safe way.
//It's GetCertificateFunc conforms to tls.Config.GetCertificate function type.
type ReloadableTLSCertProvider struct {
	constructor TLSCertConstructor
	holder      *TLSCrtKeyPairHolder
}

//NewReloadableTLSCertProvider creates a new instance of ReloadableTLSCertProvider.
func NewReloadableTLSCertProvider(constructor TLSCertConstructor) (*ReloadableTLSCertProvider, error) {

	result := &ReloadableTLSCertProvider{
		constructor: constructor,
		holder:      NewTLSCrtKeyPairHolder(),
	}

	//Initial read
	err := result.reload()
	if err != nil {
		return nil, err
	}

	return result, nil
}

//Reload reloads the internal instance.
//It's safe to call it from other goroutines.
func (ckpr *ReloadableTLSCertProvider) Reload() {
	err := ckpr.reload()
	if err != nil {
		glog.Errorf("Failed to reload certificate: %v", err)
	}
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

//AuthReqConstructor knows how to construct an authn.CancellableAuthRequest instance
type CancellableAuthReqestConstructor func() (authn.CancellableAuthRequest, error)

//ReloadableAuthReq enables to create and re-create an instance of authenticator.Request in a thread-safe way.
//It's used to re-create authenticator.Request instance every time a change in oidc-ca-file is detected.
//It implements authenticator.Request interface so it can be easily plugged in instead of a "real" instance.
type ReloadableAuthReq struct {
	constructor CancellableAuthReqestConstructor
	holder      *AuthenticatorRequestHolder
}

//NewReloadableAuthReq creates a new instance of ReloadableAuthReq.
//It requires a constructor to re-create the internal instance once Reload() is invoked.
func NewReloadableAuthReq(constructor CancellableAuthReqestConstructor) (*ReloadableAuthReq, error) {
	result := ReloadableAuthReq{
		constructor: constructor,
		holder:      NewAuthReqHolder(),
	}

	//Initial read
	err := result.reload()
	if err != nil {
		return nil, err
	}

	return &result, nil
}

//Reload reloads internal instance.
//It's safe to call it from other goroutines.
func (rar *ReloadableAuthReq) Reload() {
	err := rar.reload()
	if err != nil {
		glog.Errorf("Failed to reload OIDC Authenticator instance: %v", err)
	}
}

//reloads the internal instance using provided constructor function
//Because OIDC Authenticators spawn their own goroutines, it also cancels the old object upon creating a new one.
//Note: It must NOT modify the existing value in case of an error!
func (rar *ReloadableAuthReq) reload() error {
	newObject, err := rar.constructor()
	if err != nil {
		return err
	}

	oldObject := rar.holder.Get()
	if oldObject != nil {
		glog.Info("Cancelling previous OIDC Authenticator instance")
		oldObject.Cancel()
	}

	rar.holder.Set(newObject)
	return nil
}

//AuthenticateRequest implements authenticator.Request interface
func (rar *ReloadableAuthReq) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	//Delegate to internally-stored instance (thread-safe)
	return rar.holder.Get().AuthenticateRequest(req)
}

//AuthenticatorRequestHolder keeps an authenticator.Request instance.
//It allows for Get/Set operations in a thread-safe way
type AuthenticatorRequestHolder struct {
	rwmu  sync.RWMutex
	value authn.CancellableAuthRequest
}

//NewAuthReqHolder returns new AuthenticatorRequestHolder instance
func NewAuthReqHolder() *AuthenticatorRequestHolder {
	return &AuthenticatorRequestHolder{}
}

//Get returns the instances stored in the AuthenticatorRequestHolder
func (arh *AuthenticatorRequestHolder) Get() authn.CancellableAuthRequest {
	arh.rwmu.RLock()
	defer arh.rwmu.RUnlock()
	return arh.value
}

//Set stores given instances in the AuthenticatorRequestHolder
func (arh *AuthenticatorRequestHolder) Set(v authn.CancellableAuthRequest) {
	arh.rwmu.Lock()
	defer arh.rwmu.Unlock()
	arh.value = v
}
