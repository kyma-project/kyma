package tracepipeline

import (
	"crypto/sha256"
	"fmt"
	"hash"
)

type ConfigHash struct {
	hash hash.Hash
}

func NewConfigHash() *ConfigHash {
	return &ConfigHash{
		hash: sha256.New(),
	}
}

func (c *ConfigHash) AddStringMap(m map[string]string) *ConfigHash {
	for k, v := range m {
		c.hash.Write([]byte(k))
		c.hash.Write([]byte(v))
	}
	return c
}

func (c *ConfigHash) AddByteMap(m map[string][]byte) *ConfigHash {
	for k, v := range m {
		c.hash.Write([]byte(k))
		c.hash.Write(v)
	}
	return c
}

func (c *ConfigHash) Build() string {
	return fmt.Sprintf("%x", c.hash.Sum(nil))
}
