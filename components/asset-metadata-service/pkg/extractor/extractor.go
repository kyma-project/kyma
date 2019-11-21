package extractor

import (
	"time"

	"github.com/gernest/front"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-metadata-service/pkg/fileheader"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

//go:generate mockery -name=Extractor -output=automock -outpkg=automock -case=underscore
// Extractor is a metadata extractor
type Extractor interface {
	ReadMetadata(fileHeader fileheader.FileHeader) (map[string]interface{}, error)
}

type extractor struct {
	frontMatter *front.Matter
}

var (
	readingMetadataHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "assetstore_metadata_service_read_file_metadata_duration_seconds",
		Help:    "Reading Metadata from file duration distribution",
		Buckets: prometheus.ExponentialBuckets(0.00001, 2, 16),
	})
)

// New constructs a new Extractor instance
func New() Extractor {
	f := front.NewMatter()
	f.Handle("---", front.YAMLHandler)

	return &extractor{
		frontMatter: f,
	}
}

// ReadMetadata opens file and reads its metadata
func (e *extractor) ReadMetadata(fileHeader fileheader.FileHeader) (map[string]interface{}, error) {
	start := time.Now()
	f, err := fileHeader.Open()
	if err != nil {
		return nil, errors.Wrapf(err, "while opening file %s", fileHeader.Filename())
	}
	defer func() {
		err := f.Close()
		if err != nil {
			glog.Error(err)
		}
	}()

	metadata, _, err := e.frontMatter.Parse(f)
	if err != nil && front.ErrIsEmpty != err && front.ErrUnknownDelim != err {
		return nil, errors.Wrapf(err, "while reading metadata from file %s", fileHeader.Filename())
	}
	
	readingMetadataHistogram.Observe(time.Since(start).Seconds())

	return metadata, nil
}
