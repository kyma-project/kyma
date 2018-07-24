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
	AsyncApiSpec(id string) (*AsyncApiSpec, bool, error)
	Content(id string) (*Content, bool, error)
	NotificationChannel(stop <-chan struct{}) <-chan notification
}

type handler func(filename string) (interface{}, bool, error)

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

	swc.registerHandler("apiSpec.json", swc.apiSpecHandler)
	swc.registerHandler("asyncApiSpec.json", swc.asyncApiSpecHandler)
	swc.registerHandler("content.json", swc.contentHandler)

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
	exists, err := swc.object(id, "apiSpec.json", apiSpec)

	return apiSpec, exists, err
}

func (swc *cache) AsyncApiSpec(id string) (*AsyncApiSpec, bool, error) {
	asyncApiSpec := new(AsyncApiSpec)
	exists, err := swc.object(id, "asyncApiSpec.json", asyncApiSpec)

	return asyncApiSpec, exists, err
}

func (swc *cache) Content(id string) (*Content, bool, error) {
	content := new(Content)
	exists, err := swc.object(id, "content.json", content)

	return content, exists, err
}

func (swc *cache) object(parent, filename string, value interface{}) (bool, error) {
	data, isCached, err := swc.fromCache(parent, filename)
	if err != nil {
		return false, err
	}

	if !isCached || !swc.isCacheEnabled {
		err = swc.updateCache(parent, filename)
		if err != nil {
			return false, errors.Wrapf(err, "while updating cache for `%s/%s`", parent, filename)
		}

		data, isCached, err = swc.fromCache(parent, filename)
		if err != nil || !isCached {
			return false, err
		}
	}

	err = swc.convertFromCache(data, value)
	if err != nil {
		return false, errors.Wrapf(err, "while decoding `%s/%s` from cache", parent, filename)
	}

	return true, nil
}

func (swc *cache) updateCache(parent, filename string) error {
	handle, ok := swc.handlers[filename]
	if !ok {
		return fmt.Errorf("unknown handler for `%s/%s`", parent, filename)
	}

	object, exists, err := handle(parent)
	if err != nil {
		return errors.Wrapf(err, "while handling `%s/%s`", parent, filename)
	}

	if exists {
		return swc.storeInCache(parent, filename, object)
	}

	return swc.removeFromCache(parent, filename)
}

func (swc *cache) storeInCache(parent, filename string, object interface{}) error {
	data, err := swc.convertToCache(object)
	if err != nil {
		return errors.Wrapf(err, "while converting `%s/%s` to cache format", parent, filename)
	}

	err = swc.cache.Set(swc.cacheId(parent, filename), data)
	if err != nil {
		return errors.Wrapf(err, "while storing `%s/%s` in cache", parent, filename)
	}

	return nil
}

func (swc *cache) removeFromCache(parent, filename string) error {
	err := swc.cache.Delete(swc.cacheId(parent, filename))
	if err != nil && !swc.isEntryNotFound(err) {
		return errors.Wrapf(err, "while removing `%s/%s` from cache", parent, filename)
	}

	return nil
}

func (swc *cache) fromCache(parent, filename string) ([]byte, bool, error) {
	inCache := true
	data, err := swc.cache.Get(swc.cacheId(parent, filename))
	if err != nil {
		if !swc.isEntryNotFound(err) {
			return nil, false, errors.Wrapf(err, "while gathering from cache `%s/%s`", parent, filename)
		}

		inCache = false
	}

	return data, inCache, nil
}

func (swc *cache) apiSpecHandler(id string) (interface{}, bool, error) {
	return swc.store.ApiSpec(id)
}

func (swc *cache) asyncApiSpecHandler(id string) (interface{}, bool, error) {
	return swc.store.AsyncApiSpec(id)
}

func (swc *cache) contentHandler(id string) (interface{}, bool, error) {
	return swc.store.Content(id)
}

func (swc *cache) registerHandler(filename string, handler func(string) (interface{}, bool, error)) {
	_, registered := swc.handlers[filename]
	if registered {
		glog.Warning("handler: `%s` already registered", filename)
	}
	swc.handlers[filename] = handler
}

func (swc *cache) cacheId(parent, filename string) string {
	return fmt.Sprintf("%s/%s", parent, filename)
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
