package tracepipeline

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	empty1  = map[string]string{}
	empty2  = map[string][]byte{}
	config1 = map[string]string{
		"a": "b",
		"c": "d",
	}
	config2 = map[string][]byte{
		"1": []byte("2"),
		"3": []byte("4"),
	}
)

func TestEqualConfig(t *testing.T) {
	hash1 := NewConfigHash().AddStringMap(config1).AddByteMap(config2).Build()
	hash2 := NewConfigHash().AddStringMap(config1).AddByteMap(config2).Build()
	require.Equal(t, hash1, hash2)
}

func TestUnequalConfig(t *testing.T) {
	hash1 := NewConfigHash().AddStringMap(config1).AddByteMap(config2).Build()
	hash2 := NewConfigHash().AddByteMap(config2).AddByteMap(config2).Build()
	require.NotEqual(t, hash1, hash2)
}

func TestEmptyConfig(t *testing.T) {
	hash := NewConfigHash().AddStringMap(empty1).AddByteMap(empty2).Build()
	require.NotEmpty(t, hash)
}
