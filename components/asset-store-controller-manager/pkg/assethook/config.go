package assethook

import "time"

type Config struct {
	MutationTimeout           time.Duration `envconfig:"default=1m"`
	ValidationTimeout         time.Duration `envconfig:"default=1m"`
	MetadataExtractionTimeout time.Duration `envconfig:"default=1m"`
}
