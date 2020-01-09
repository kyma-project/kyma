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

//TLSCertReloader enables to create and re-create an instance of tls.Certificate in a thread-safe way.
//It's GetCertificateFunc conforms to tls.Config.GetCertificate function type.
type TLSCertReloader struct {
	constructor TLSCertConstructor
	holder      *TLSCertHolder
}

//NewTLSCertReloader creates a new instance of TLSCertReloader.
func NewTLSCertReloader(constructor TLSCertConstructor) (*TLSCertReloader, error) {

	result := &TLSCertReloader{
		constructor: constructor,
		holder:      NewTLSCertHolder(),
	}

	//Initial read
	err := result.reload()
	if err != nil {
		return nil, err
	}

	return result, nil
}

//reloads the internal instance using provided constructor function
//Note: It must NOT modify the existing value in case of an error!
func (ckpr *TLSCertReloader) reload() error {
	newCert, err := ckpr.constructor()
	if err != nil {
		return err
	}
	ckpr.holder.Set(newCert)
	return nil
}

//Reload reloads the internal instance.
//It's safe to call it from other goroutines.
func (ckpr *TLSCertReloader) Reload() {
	err := ckpr.reload()
	if err != nil {
		glog.Errorf("Failed to reload certificate: %v", err)
	}
}

//GetCertificateFunc conforms to tls.Config.GetCertificate function type
func (ckpr *TLSCertReloader) GetCertificateFunc(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return ckpr.holder.Get(), nil
}

//TLSCertHolder keeps a tls.Certificate instance and allows for Get/Set operations in a thread-safe way
type TLSCertHolder struct {
	rwmu  sync.RWMutex
	value *tls.Certificate
}

//NewTLSCertHolder returns new TLSCertHolder instance
func NewTLSCertHolder() *TLSCertHolder {
	return &TLSCertHolder{}
}

//Get returns the tls.Certificate instance stored in the TLSCertHolder
func (tlsh *TLSCertHolder) Get() *tls.Certificate {
	tlsh.rwmu.RLock()
	defer tlsh.rwmu.RUnlock()
	return tlsh.value
}

//Set stores given tls.Certificate in the TLSCertHolder
func (tlsh *TLSCertHolder) Set(v *tls.Certificate) {
	tlsh.rwmu.Lock()
	defer tlsh.rwmu.Unlock()
	tlsh.value = v
}

//AuthReqConstructor knows how to construct an authn.CancelableAuthRequest instance
type CancelableAuthReqestConstructor func() (authn.CancelableAuthRequest, error)

//CancelableAuthReqestReloader enables to create and re-create an instance of authn.CancelableAuthRequest in a thread-safe way.
//It implements authenticator.Request interface so it can be easily plugged in instead of a "real" instance.
type CancelableAuthReqestReloader struct {
	constructor CancelableAuthReqestConstructor
	holder      *CancelableAuthReqestHolder
}

//NewCancelableAuthReqestReloader creates a new instance of CancelableAuthReqestReloader.
//It requires a constructor to re-create the internal instance once Reload() is invoked.
func NewCancelableAuthReqestReloader(constructor CancelableAuthReqestConstructor) (*CancelableAuthReqestReloader, error) {
	result := CancelableAuthReqestReloader{
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
func (rar *CancelableAuthReqestReloader) Reload() {
	err := rar.reload()
	if err != nil {
		glog.Errorf("Failed to reload OIDC Authenticator instance: %v", err)
	}
}

//reloads the internal instance using provided constructor function
//Because OIDC Authenticators spawn their own goroutines, it also cancels the old object upon creating a new one.
//Note: It must NOT modify the existing value in case of an error!
func (rar *CancelableAuthReqestReloader) reload() error {
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

func ReloadAll(reloaders []CancelableAuthReqestReloader) (func()) {
	return func(){
		for _, r := range  reloaders{
			r.Reload()
		}
	}
}

//AuthenticateRequest implements authenticator.Request interface
func (rar *CancelableAuthReqestReloader) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	//Delegate to internally-stored instance (thread-safe)
	return rar.holder.Get().AuthenticateRequest(req)
}

//CancelableAuthReqestHolder keeps an authenticator.Request instance.
//It allows for Get/Set operations in a thread-safe way
type CancelableAuthReqestHolder struct {
	rwmu  sync.RWMutex
	value authn.CancelableAuthRequest
}

//NewAuthReqHolder returns new CancelableAuthReqestHolder instance
func NewAuthReqHolder() *CancelableAuthReqestHolder {
	return &CancelableAuthReqestHolder{}
}

//Get returns the instances stored in the CancelableAuthReqestHolder
func (arh *CancelableAuthReqestHolder) Get() authn.CancelableAuthRequest {
	arh.rwmu.RLock()
	defer arh.rwmu.RUnlock()
	return arh.value
}

//Set stores given instances in the CancelableAuthReqestHolder
func (arh *CancelableAuthReqestHolder) Set(v authn.CancelableAuthRequest) {
	arh.rwmu.Lock()
	defer arh.rwmu.Unlock()
	arh.value = v
}
