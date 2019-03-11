package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/allegro/bigcache"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

//go:generate mockery -name=storeGetter -inpkg -case=underscore
type storeGetter interface {
	ApiSpec(id string) (*ApiSpec, bool, error)
	OpenApiSpec(id string) (*OpenApiSpec, bool, error)
	ODataSpec(id string) (*ODataSpec, bool, error)
	AsyncApiSpec(id string) (*AsyncApiSpec, bool, error)
	Content(id string) (*Content, bool, error)
	NotificationChannel(stop <-chan struct{}) <-chan notification
}

const (
	ApiSpecField      = "apiSpec"
	OpenApiSpecField  = "openApiSpec"
	ODataSpecField    = "odataSpec"
	AsyncApiSpecField = "asyncApiSpec"
	ContentField      = "content"
)

type handler func(name string) (interface{}, bool, error)

type cache struct {
	store          storeGetter
	cache          Cache
	isInitialized  bool
	isCacheEnabled bool
	handlers       map[string]handler
}

func newCache(store storeGetter, cacheClient Cache) *cache {
	swc := &cache{
		store:          store,
		cache:          cacheClient,
		isInitialized:  false,
		isCacheEnabled: false,
		handlers:       make(map[string]handler),
	}

	swc.registerHandler("apiSpec.json", ApiSpecField, swc.apiSpecHandler)
	swc.registerHandler("apiSpec.json", OpenApiSpecField, swc.openApiSpecHandler)
	swc.registerHandler("apiSpec.json", ODataSpecField, swc.odataSpecHandler)
	swc.registerHandler("asyncApiSpec.json", AsyncApiSpecField, swc.asyncApiSpecHandler)
	swc.registerHandler("content.json", ContentField, swc.contentHandler)

	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})

	return swc
}

func (swc *cache) Initialize(stop <-chan struct{}) {
	if swc.isInitialized {
		return
	}
	swc.isInitialized = true

	go func() {
		for {
			if err := swc.cache.Reset(); err != nil {
				glog.Error(errors.Wrap(err, "while resetting cache"))
				// if cache cannot bb restarted then caching should be disabled
				return
			}

			notifications := swc.store.NotificationChannel(stop)
			swc.isCacheEnabled = true
			for notification := range notifications {
				_, ok := swc.handlers[notification.filename]
				if !ok {
					// unknown file type
					continue
				}

				err := swc.updateCache(notification.parent, notification.filename)
				if err != nil {
					glog.Error(errors.Wrapf(err, "while handling %v notification", notification))
				}
			}
			swc.isCacheEnabled = false

			select {
			case <-stop:
				return
			default:
				time.Sleep(15 * time.Second)
			}
		}
	}()
}

func (swc *cache) IsSynced() bool {
	return swc.isCacheEnabled
}

func (swc *cache) ApiSpec(id string) (*ApiSpec, bool, error) {
	apiSpec := new(ApiSpec)
	exists, err := swc.object(id, "apiSpec.json", ApiSpecField, apiSpec)

	return apiSpec, exists, err
}

func (swc *cache) OpenApiSpec(id string) (*OpenApiSpec, bool, error) {
	openApiSpec := new(OpenApiSpec)
	exists, err := swc.object(id, "apiSpec.json", OpenApiSpecField, openApiSpec)

	return openApiSpec, exists, err
}

func (swc *cache) ODataSpec(id string) (*ODataSpec, bool, error) {
	odataSpec := new(ODataSpec)
	exists, err := swc.object(id, "apiSpec.json", ODataSpecField, odataSpec)

	return odataSpec, exists, err
}

func (swc *cache) AsyncApiSpec(id string) (*AsyncApiSpec, bool, error) {
	asyncApiSpec := new(AsyncApiSpec)
	exists, err := swc.object(id, "asyncApiSpec.json", AsyncApiSpecField, asyncApiSpec)

	return asyncApiSpec, exists, err
}

func (swc *cache) Content(id string) (*Content, bool, error) {
	content := new(Content)
	exists, err := swc.object(id, "content.json", ContentField, content)

	return content, exists, err
}

func (swc *cache) object(parent, filename, fieldName string, value interface{}) (bool, error) {
	name := swc.registerHandlerName(filename, fieldName)

	data, isCached, err := swc.fromCache(parent, name)
	if err != nil {
		return false, err
	}

	if !isCached || !swc.isCacheEnabled {
		err = swc.updateCache(parent, name)
		if err != nil {
			return false, errors.Wrapf(err, "while updating cache for `%s/%s`", parent, name)
		}

		data, isCached, err = swc.fromCache(parent, name)
		if err != nil || !isCached {
			return false, err
		}
	}

	err = swc.convertFromCache(data, value)
	if err != nil {
		return false, errors.Wrapf(err, "while decoding `%s/%s` from cache", parent, name)
	}

	return true, nil
}

func (swc *cache) updateCache(parent, name string) error {
	handle, ok := swc.handlers[name]
	if !ok {
		return fmt.Errorf("unknown handler for `%s/%s`", parent, name)
	}

	object, exists, err := handle(parent)
	if err != nil {
		return errors.Wrapf(err, "while handling `%s/%s`", parent, name)
	}

	if exists {
		return swc.storeInCache(parent, name, object)
	}

	return swc.removeFromCache(parent, name)
}

func (swc *cache) storeInCache(parent, name string, object interface{}) error {
	data, err := swc.convertToCache(object)
	if err != nil {
		return errors.Wrapf(err, "while converting `%s/%s` to cache format", parent, name)
	}

	err = swc.cache.Set(swc.cacheId(parent, name), data)
	if err != nil {
		return errors.Wrapf(err, "while storing `%s/%s` in cache", parent, name)
	}

	return nil
}

func (swc *cache) removeFromCache(parent, name string) error {
	err := swc.cache.Delete(swc.cacheId(parent, name))
	if err != nil && !swc.isEntryNotFound(err) {
		return errors.Wrapf(err, "while removing `%s/%s` from cache", parent, name)
	}

	return nil
}

func (swc *cache) fromCache(parent, name string) ([]byte, bool, error) {
	inCache := true
	data, err := swc.cache.Get(swc.cacheId(parent, name))
	if err != nil {
		if !swc.isEntryNotFound(err) {
			return nil, false, errors.Wrapf(err, "while gathering from cache `%s/%s`", parent, name)
		}

		inCache = false
	}

	return data, inCache, nil
}

func (swc *cache) apiSpecHandler(id string) (interface{}, bool, error) {
	return swc.store.ApiSpec(id)
}

func (swc *cache) openApiSpecHandler(id string) (interface{}, bool, error) {
	return swc.store.OpenApiSpec(id)
}

func (swc *cache) odataSpecHandler(id string) (interface{}, bool, error) {
	return swc.store.ODataSpec(id)
}

func (swc *cache) asyncApiSpecHandler(id string) (interface{}, bool, error) {
	return swc.store.AsyncApiSpec(id)
}

func (swc *cache) contentHandler(id string) (interface{}, bool, error) {
	return swc.store.Content(id)
}

func (swc *cache) registerHandler(filename, fieldName string, handler func(string) (interface{}, bool, error)) {
	name := swc.registerHandlerName(filename, fieldName)

	_, registered := swc.handlers[name]
	if registered {
		glog.Warningf("handler: `%s` already registered", name)
	}
	swc.handlers[name] = handler
}

func (swc *cache) registerHandlerName(filename, fieldName string) string {
	return fmt.Sprintf("%s/%s", filename, fieldName)
}

func (swc *cache) cacheId(parent, name string) string {
	return fmt.Sprintf("%s/%s", parent, name)
}

func (swc *cache) isEntryNotFound(err error) bool {
	_, ok := err.(*bigcache.EntryNotFoundError)
	return ok
}

func (swc *cache) convertToCache(object interface{}) ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := gob.NewEncoder(buffer).Encode(object)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (swc *cache) convertFromCache(data []byte, value interface{}) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(value)
}
