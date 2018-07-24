package content

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
)

type contentResolver struct {
	contentGetter contentGetter
	converter     *contentConverter
}

func newContentResolver(contentGetter contentGetter) *contentResolver {
	return &contentResolver{
		contentGetter: contentGetter,
		converter:     &contentConverter{},
	}
}

func (r *contentResolver) ContentQuery(ctx context.Context, contentType, id string) (*gqlschema.JSON, error) {
	item, err := r.contentGetter.Find(contentType, id)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering content for type `%s` with id `%s`", contentType, id))
		return nil, r.genericError()
	}

	return r.converter.ToGQL(item), nil
}

func (r *contentResolver) genericError() error {
	return errors.New("Cannot get Content")
}
