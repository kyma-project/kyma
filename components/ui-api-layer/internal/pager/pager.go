package pager

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type PagingParams struct {
	First  *int
	Offset *int
}

type Pager struct {
	store PageableStore
}

//go:generate mockery -name=PageableStore -output=automock -outpkg=automock -case=underscore
type PageableStore interface {
	GetByKey(key string) (item interface{}, exists bool, err error)
	List() []interface{}
	ListKeys() []string
}

func From(store PageableStore) *Pager {
	return &Pager{
		store: store,
	}
}

func (p *Pager) Limit(params PagingParams) ([]interface{}, error) {
	items := p.store.List()
	keys := p.store.ListKeys()
	internalParams := p.readParams(params)
	return p.limitList(internalParams, items, keys, p.store)
}

type itemGetter interface {
	GetByKey(key string) (item interface{}, exists bool, err error)
}

func (p *Pager) readParams(params PagingParams) PagingParams {
	var f, o int
	if params.First != nil {
		f = *params.First
	}
	if params.Offset != nil {
		o = *params.Offset
	}

	return PagingParams{
		First:  &f,
		Offset: &o,
	}
}

func (p *Pager) limitList(params PagingParams, items []interface{}, keys []string, getter itemGetter) ([]interface{}, error) {
	if len(items) == 0 {
		return []interface{}{}, nil
	}

	keysCount := len(keys)

	first := *params.First
	offset := *params.Offset

	if first < 0 {
		return nil, errors.New("'First' parameter cannot be below 0")
	}

	if offset < 0 {
		return nil, errors.New("'Offset' parameter cannot be below 0")
	}

	sliceStart := offset
	sliceEnd := sliceStart + first

	if sliceStart >= keysCount {
		return nil, fmt.Errorf("Offset %d is out of range; maximum value: %d", sliceStart, keysCount-1)
	}

	if sliceEnd >= keysCount {
		sliceEnd = keysCount
	}

	sortedList, err := p.sortByKey(keys, getter)
	if err != nil {
		return nil, errors.Wrap(err, "while sorting store")
	}

	if offset == 0 && (first == 0 || first >= keysCount) {
		return sortedList, nil
	}

	return sortedList[sliceStart:sliceEnd], nil
}

func (p *Pager) sortByKey(keys []string, store itemGetter) ([]interface{}, error) {
	var sortedKeys []string
	sortedKeys = append(sortedKeys, keys...)

	sort.SliceStable(sortedKeys, func(i, j int) bool {
		result := strings.Compare(sortedKeys[i], sortedKeys[j])
		return result != 1
	})

	var sortedList []interface{}
	for _, key := range sortedKeys {
		item, exists, err := store.GetByKey(key)
		if !exists {
			return nil, fmt.Errorf("Item with key %s doesn't exist", key)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "While getting item with key %s", key)
		}

		sortedList = append(sortedList, item)
	}

	return sortedList, nil
}
