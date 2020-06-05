package pager

import (
	"github.com/pkg/errors"
)

type IndexPager struct {
	*Pager
	indexer   PageableIndexer
	indexName string
	indexKey  string
}

//go:generate mockery -name=PageableIndexer -output=automock -outpkg=automock -case=underscore
type PageableIndexer interface {
	ByIndex(indexName, indexKey string) ([]interface{}, error)
	IndexKeys(indexName, indexKey string) ([]string, error)
	GetByKey(key string) (item interface{}, exists bool, err error)
}

func FromIndexer(indexer PageableIndexer, indexName, indexKey string) *IndexPager {
	return &IndexPager{
		indexer:   indexer,
		indexName: indexName,
		indexKey:  indexKey,
	}
}

func (p *IndexPager) Limit(params PagingParams) ([]interface{}, error) {
	items, err := p.indexer.ByIndex(p.indexName, p.indexKey)
	if err != nil {
		return nil, errors.Wrap(err, "while getting items by index from indexer")
	}
	keys, err := p.indexer.IndexKeys(p.indexName, p.indexKey)
	if err != nil {
		return nil, errors.Wrap(err, "while getting index keys for indexer")
	}

	internalParams := p.readParams(params)
	return p.limitList(internalParams, items, keys, p.indexer)
}
