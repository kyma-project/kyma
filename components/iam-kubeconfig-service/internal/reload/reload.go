package reload

import (
	"net/http"
	"sync"

	"github.com/kyma-project/kyma/components/iam-kubeconfig-service/internal/authn"
	log "github.com/sirupsen/logrus"
	"k8s.io/apiserver/pkg/authentication/authenticator"
)

//StringConstructor knows how to construct a string
type StringConstructor func() (string, error)

//ReloadableStringProvider enables to create and re-create a string in a thread-safe way.
type ReloadableStringProvider struct {
	constructor StringConstructor
	holder      *StringHolder
}

//NewReloadableStringProvider creates a new instance of ReloadableStringProvider.
func NewReloadableStringProvider(constructor StringConstructor) (*ReloadableStringProvider, error) {

	result := &ReloadableStringProvider{
		constructor: constructor,
		holder:      NewStringHolder(),
	}

	//Initial read
	err := result.reload()
	if err != nil {
		return nil, err
	}

	return result, nil
}

//Reload reloads the internal instance.
//It's purpose is to trigger reloading from other goroutines
func (rsp *ReloadableStringProvider) Reload() {
	err := rsp.reload()
	if err != nil {
		log.Errorf("Failed to reload value: %v", err)
	}
}

//reloads the internal instance using provided constructor function
//Note: It must NOT modify the existing value in case of an error!
func (rsp *ReloadableStringProvider) reload() error {
	v, err := rsp.constructor()
	if err != nil {
		return err
	}
	rsp.holder.Set(v)
	return nil
}

//GetString returns the string value stored in ReloadableStringProvider
func (rsp *ReloadableStringProvider) GetString() string {
	return rsp.holder.Get()
}

//StringHolder keeps a string and allows for Get/Set operations in a thread-safe way
type StringHolder struct {
	rwmu  sync.RWMutex
	value string
}

//NewStringHolder returns new StringHolder instance
func NewStringHolder() *StringHolder {
	return &StringHolder{}
}

//Get returns the string stored in the StringHolder
func (tlsh *StringHolder) Get() string {
	tlsh.rwmu.RLock()
	defer tlsh.rwmu.RUnlock()
	return tlsh.value
}

//Set stores given string in the StringHolder
func (tlsh *StringHolder) Set(v string) {
	tlsh.rwmu.Lock()
	defer tlsh.rwmu.Unlock()
	tlsh.value = v
}

//AuthReqConstructor knows how to construct an authn.CancellableAuthRequest instance
type AuthReqConstructor func() (authn.CancellableAuthRequest, error)

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
func NewReloadableAuthReq(constructor AuthReqConstructor) (*ReloadableAuthReq, error) {
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
func (rar *ReloadableAuthReq) Reload() {
	err := rar.reload()
	if err != nil {
		log.Errorf("Failed to reload OIDC Authenticator instance: %v", err)
	}
}

//reloads the internal instance using provided constructor function
//Note: It must NOT modify the existing value in case of an error!
func (rar *ReloadableAuthReq) reload() error {
	newAuthReq, err := rar.constructor()
	if err != nil {
		return err
	}

	oldAuthReq := rar.holder.Get()
	if oldAuthReq != nil {
		log.Info("Cancelling previous OIDC Authenticator instance")
		oldAuthReq.Cancel()
	}

	rar.holder.Set(newAuthReq)
	return nil
}

//AuthenticateRequest implements authenticator.Request interface
func (rar *ReloadableAuthReq) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	//Delegate to internally-stored instance (thread-safe)
	return rar.holder.Get().AuthenticateRequest(req)
}

//AuthReqHolder keeps an authenticator.Request and an authn.AuthenticatorCancelFunc instance.
//It allows for Get/Set operations in a thread-safe way
type AuthReqHolder struct {
	rwmu  sync.RWMutex
	value authn.CancellableAuthRequest
}

//NewAuthReqHolder returns new AuthReqHolder instance
func NewAuthReqHolder() *AuthReqHolder {
	return &AuthReqHolder{}
}

//Get returns the instances stored in the AuthReqHolder
func (arh *AuthReqHolder) Get() authn.CancellableAuthRequest {
	arh.rwmu.RLock()
	defer arh.rwmu.RUnlock()
	return arh.value
}

//Set stores given instances in the AuthReqHolder
func (arh *AuthReqHolder) Set(v authn.CancellableAuthRequest) {
	arh.rwmu.Lock()
	defer arh.rwmu.Unlock()
	arh.value = v
}
