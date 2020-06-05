package gqlschema

import (
	"io"
	"strconv"
	"time"

	"github.com/golang/glog"

	"github.com/99designs/gqlgen/graphql"
	"github.com/pkg/errors"
)

func MarshalTimestamp(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		_, err := io.WriteString(w, strconv.FormatInt(t.Unix(), 10))
		if err != nil {
			glog.Error(errors.Wrap(err, "while writing marshalled timestamp"))
			return
		}
	})
}

func UnmarshalTimestamp(v interface{}) (time.Time, error) {
	if tmpStr, ok := v.(int); ok {
		return time.Unix(int64(tmpStr), 0), nil
	}
	return time.Time{}, errors.New("Time should be an unix timestamp")
}
