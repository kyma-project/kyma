package reload

import (
	"net/http"
	"sync"

	"github.com/kyma-project/kyma/components/iam-kubeconfig-service/internal/authn"
	log "github.com/sirupsen/logrus"
	"k8s.io/apiserver/pkg/authentication/authenticator"
)

//StringValueConstructor knows how to construct a string
type StringValueConstructor func() (string, error)

//StringValueReloader enables to create and re-create a string in a thread-safe way.
type StringValueReloader struct {
	constructor StringValueConstructor
	holder      *StringValueHolder
}

//NewStringValueReloader returns a new instance of StringValueReloader.
func NewStringValueReloader(constructor StringValueConstructor) (*StringValueReloader, error) {

	result := &StringValueReloader{
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
//It's safe to call it from other goroutines
func (rsp *StringValueReloader) Reload() {
	err := rsp.reload()
	if err != nil {
		log.Errorf("Failed to reload value: %v", err)
	}
}

//reloads the internal instance using provided constructor function
//Note: It must NOT modify the existing value in case of an error!
func (rsp *StringValueReloader) reload() error {
	v, err := rsp.constructor()
	if err != nil {
		return err
	}
	rsp.holder.Set(v)
	return nil
}

//GetString returns the string value stored in StringValueReloader
func (rsp *StringValueReloader) GetString() string {
	return rsp.holder.Get()
}

//StringValueHolder keeps a string and allows for Get/Set operations in a thread-safe way
type StringValueHolder struct {
	rwmu  sync.RWMutex
	value string
}

//NewStringHolder returns new StringValueHolder instance
func NewStringHolder() *StringValueHolder {
	return &StringValueHolder{}
}

//Get returns the string stored in the StringValueHolder
func (tlsh *StringValueHolder) Get() string {
	tlsh.rwmu.RLock()
	defer tlsh.rwmu.RUnlock()
	return tlsh.value
}

//Set stores given string in the StringValueHolder
func (tlsh *StringValueHolder) Set(v string) {
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
		log.Errorf("Failed to reload OIDC Authenticator instance: %v", err)
	}
}

//Reloads the internal instance using provided constructor function
//Because OIDC Authenticators spawn their own goroutines, it also cancels the old object upon creating a new one.
//Note: It must NOT modify the existing value in case of an error!
func (rar *CancelableAuthReqestReloader) reload() error {
	newObject, err := rar.constructor()
	if err != nil {
		return err
	}

	oldObject := rar.holder.Get()
	if oldObject != nil {
		log.Info("Canceling previous OIDC Authenticator instance")
		oldObject.Cancel()
	}

	rar.holder.Set(newObject)
	return nil
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
