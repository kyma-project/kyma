package pager

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPager_Limit(t *testing.T) {
	keys := []string{
		"Test1",
		"Test2",
		"Test3",
	}
	values := []interface{}{
		1,
		2,
		3,
	}

	t.Run("Empty list", func(t *testing.T) {
		first := 30
		offset := 0
		params := PagingParams{
			First:  &first,
			Offset: &offset,
		}
		expectedItems := []interface{}{}
		store := new(automock.PageableStore)
		store.On("ListKeys").Return(keys)
		store.On("List").Return([]interface{}{})
		pager := From(store)

		items, err := pager.Limit(params)

		require.NoError(t, err)
		assert.Equal(t, expectedItems, items)
	})

	t.Run("No paging parameters given", func(t *testing.T) {
		store := fixStore(keys, values)
		pager := From(store)
		first := 0
		offset := 0
		params := PagingParams{
			First:  &first,
			Offset: &offset,
		}
		expectedItems := values

		items, err := pager.Limit(params)

		require.NoError(t, err)
		assert.Equal(t, expectedItems, items)
	})

	t.Run("Less items than given 'first' parameter, no offset", func(t *testing.T) {
		store := fixStore(keys, values)
		pager := From(store)
		first := 5
		offset := 0
		params := PagingParams{
			First:  &first,
			Offset: &offset,
		}
		expectedItems := values

		items, err := pager.Limit(params)

		require.NoError(t, err)
		assert.Equal(t, expectedItems, items)
	})

	t.Run("Less items than given 'first' parameter, offset included", func(t *testing.T) {
		store := fixStore(keys, values)
		pager := From(store)
		first := 5
		offset := 1
		params := PagingParams{
			First:  &first,
			Offset: &offset,
		}
		expectedItems := []interface{}{
			2,
			3,
		}

		items, err := pager.Limit(params)

		require.NoError(t, err)
		assert.Equal(t, expectedItems, items)
	})

	t.Run("More items, 'first' and 'offset' parameters given", func(t *testing.T) {
		store := fixStore(keys, values)
		pager := From(store)
		first := 1
		offset := 1
		params := PagingParams{
			First:  &first,
			Offset: &offset,
		}
		expectedItems := []interface{}{
			2,
		}

		items, err := pager.Limit(params)

		require.NoError(t, err)
		assert.Equal(t, expectedItems, items)
	})

}

func fixStore(keys []string, values []interface{}) *automock.PageableStore {
	store := new(automock.PageableStore)
	store.On("ListKeys").Return(keys)
	store.On("List").Return(values)

	for index, key := range keys {
		store.On("GetByKey", key).Return(values[index], true, nil)
	}

	return store
}
