package handlers

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

const (
	JetStreamMemoryStorageType = "memory"
	JetStreamMemoryFileType    = "file"
)

func toJetStreamStorageType(s string) (nats.StorageType, error) {
	switch s {
	case JetStreamMemoryStorageType:
		return nats.MemoryStorage, nil
	case JetStreamMemoryFileType:
		return nats.FileStorage, nil
	}
	return nats.MemoryStorage, fmt.Errorf("invalid stream storage type %q", s)
}
