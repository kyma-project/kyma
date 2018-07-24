package content

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
)

type topicsResolver struct {
	contentGetter contentGetter
	converter     topicsConverterInterface
}

func newTopicsResolver(contentGetter contentGetter) *topicsResolver {
	return &topicsResolver{
		contentGetter: contentGetter,
		converter:     &topicsConverter{},
	}
}

func (r *topicsResolver) TopicsQuery(ctx context.Context, topics []gqlschema.InputTopic, internal *bool) ([]gqlschema.TopicEntry, error) {

	var tOutput []gqlschema.TopicEntry
	var includeInternal bool

	if internal == nil {
		includeInternal = false
	} else {
		includeInternal = *internal
	}

	for _, t := range topics {
		contentType := t.Type
		id := t.ID

		topic := gqlschema.TopicEntry{ContentType: contentType, ID: id}
		item, err := r.contentGetter.Find(contentType, id)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while gathering content for type `%s` with id `%s`", contentType, id))
			return nil, r.genericError()
		}
		if item != nil {
			topic.Sections, err = r.converter.ExtractSection(item.Data.Docs, includeInternal)
			if err != nil {
				glog.Error(errors.Wrapf(err, "while extracting topics for type `%s` with id `%s`", contentType, id))
				return nil, r.genericError()
			}

			tOutput = append(tOutput, topic)
		}
	}

	return tOutput, nil
}

func (r *topicsResolver) genericError() error {
	return errors.New("Cannot get Topics")
}
