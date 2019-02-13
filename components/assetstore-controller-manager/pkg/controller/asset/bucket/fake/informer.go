package fake

import (
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"strings"
	"time"
)

type Informer struct {
	store cache.Store
}

type Object interface {
	GetObjectMeta() v1.Object
}

func NewInformer(objects ...Object) *Informer {
	store := &store{
		cache: make(map[string]interface{}),
	}
	store.addBuckets(objects...)

	return &Informer{
		store: store,
	}
}

var _ cache.SharedIndexInformer = &Informer{}

func (i *Informer) AddEventHandler(handler cache.ResourceEventHandler) {
	panic("stub")
}

func (i *Informer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) {
	panic("stub")
}

func (i *Informer) GetStore() cache.Store {
	return i.store
}

func (i *Informer) GetController() cache.Controller {
	panic("stub")
}

func (i *Informer) Run(stopCh <-chan struct{}) {
	panic("stub")
}

func (i *Informer) HasSynced() bool {
	panic("stub")
}

func (i *Informer) LastSyncResourceVersion() string {
	panic("stub")
}

func (i *Informer) AddIndexers(indexers cache.Indexers) error {
	panic("stub")
}

func (i *Informer) GetIndexer() cache.Indexer {
	panic("stub")
}

type store struct {
	cache map[string]interface{}
}

var _ cache.Store = &store{}

func (s *store) addBuckets(objects ...Object) {
	for _, obj := range objects {
		s.cache[s.getKey(obj)] = obj
	}
}

func (s *store) getKey(object Object) string {
	meta := object.GetObjectMeta()
	return fmt.Sprintf("%s/%s", meta.GetNamespace(), meta.GetName())
}

func (s *store) Add(obj interface{}) error {
	panic("stub")
}

func (s *store) Update(obj interface{}) error {
	panic("stub")
}

func (s *store) Delete(obj interface{}) error {
	panic("stub")
}

func (s *store) List() []interface{} {
	panic("stub")
}

func (s *store) ListKeys() []string {
	panic("stub")
}

func (s *store) Get(obj interface{}) (item interface{}, exists bool, err error) {
	panic("stub")
}

func (s *store) GetByKey(key string) (item interface{}, exists bool, err error) {
	if strings.HasPrefix(key, "/") {
		return nil, false, errors.New("invalid key")
	}

	obj, ok := s.cache[key]

	return obj, ok, nil
}

func (s *store) Replace([]interface{}, string) error {
	panic("stub")
}

func (s *store) Resync() error {
	panic("stub")
}
