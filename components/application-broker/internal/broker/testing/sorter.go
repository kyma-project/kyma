package testing

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	util_runtime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

func NewObjectSorter(scheme *runtime.Scheme) ObjectSorter {
	cache := make(map[reflect.Type]cache.Indexer)

	for _, v := range scheme.AllKnownTypes() {
		cache[v] = emptyIndexer()
	}

	ls := ObjectSorter{
		cache: cache,
	}

	return ls
}

type ObjectSorter struct {
	cache map[reflect.Type]cache.Indexer
}

func (o *ObjectSorter) AddObjects(objs ...runtime.Object) {
	for _, obj := range objs {
		t := reflect.TypeOf(obj).Elem()
		indexer, ok := o.cache[t]
		if !ok {
			panic(fmt.Sprintf("Unrecognized type %T", obj))
		}
		indexer.Add(obj)
	}
}

func (o *ObjectSorter) ObjectsForScheme(scheme *runtime.Scheme) []runtime.Object {
	var objs []runtime.Object

	for _, t := range scheme.AllKnownTypes() {
		indexer := o.cache[t]
		for _, item := range indexer.List() {
			objs = append(objs, item.(runtime.Object))
		}
	}

	return objs
}

func (o *ObjectSorter) ObjectsForSchemeFunc(funcs ...func(scheme *runtime.Scheme) error) []runtime.Object {
	scheme := runtime.NewScheme()

	for _, addToScheme := range funcs {
		util_runtime.Must(addToScheme(scheme))
	}

	return o.ObjectsForScheme(scheme)
}

func (o *ObjectSorter) IndexerForObjectType(obj runtime.Object) cache.Indexer {
	objType := reflect.TypeOf(obj).Elem()

	indexer, ok := o.cache[objType]

	if !ok {
		panic(fmt.Sprintf("indexer for type %v doesn't exist", objType.Name()))
	}

	return indexer
}

func emptyIndexer() cache.Indexer {
	return cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}
