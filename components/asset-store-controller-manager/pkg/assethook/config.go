package assethook

import "time"

type Config struct {
	MutationWorkers           int           `envconfig:"default=10"`
	MutationTimeout           time.Duration `envconfig:"default=1m"`
	ValidationWorkers         int           `envconfig:"default=10"`
	ValidationTimeout         time.Duration `envconfig:"default=1m"`
	MetadataExtractionTimeout time.Duration `envconfig:"default=1m"`
}
