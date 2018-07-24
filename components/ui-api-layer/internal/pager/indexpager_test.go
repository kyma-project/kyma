package pager

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLimitList(t *testing.T) {
	indexName := "environment"
	indexKey := "test"
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
		indexer := new(automock.PageableIndexer)
		indexer.On("IndexKeys", indexName, indexKey).Return(nil, nil)
		indexer.On("ByIndex", indexName, indexKey).Return(nil, nil)
		pager := FromIndexer(indexer, indexName, indexKey)

		items, err := pager.Limit(params)

		require.NoError(t, err)
		assert.Equal(t, expectedItems, items)
	})

	t.Run("No paging parameters given", func(t *testing.T) {
		indexer := fixIndexer(keys, values, indexName, indexKey)
		pager := FromIndexer(indexer, indexName, indexKey)
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
		indexer := fixIndexer(keys, values, indexName, indexKey)
		pager := FromIndexer(indexer, indexName, indexKey)
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
		indexer := fixIndexer(keys, values, indexName, indexKey)
		pager := FromIndexer(indexer, indexName, indexKey)
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
		indexer := fixIndexer(keys, values, indexName, indexKey)
		pager := FromIndexer(indexer, indexName, indexKey)
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

	t.Run("Error thrown", func(t *testing.T) {
		err := errors.New("New error")
		first := 30
		offset := 0
		params := PagingParams{
			First:  &first,
			Offset: &offset,
		}

		indexer := new(automock.PageableIndexer)
		indexer.On("IndexKeys", indexName, indexKey).Return(nil, nil)
		indexer.On("ByIndex", indexName, indexKey).Return(nil, err)
		pager := FromIndexer(indexer, indexName, indexKey)

		_, err = pager.Limit(params)

		require.Error(t, err, "while getting items by index from indexer: New error")
	})

}

func fixIndexer(keys []string, values []interface{}, indexName, indexKey string) *automock.PageableIndexer {
	indexer := new(automock.PageableIndexer)
	indexer.On("IndexKeys", indexName, indexKey).Return(keys, nil)
	indexer.On("ByIndex", indexName, indexKey).Return(values, nil)

	for index, key := range keys {
		indexer.On("GetByKey", key).Return(values[index], true, nil)
	}

	return indexer
}
