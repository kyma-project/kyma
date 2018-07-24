package ybundle

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDTO(t *testing.T) {
	// GIVEN
	data := `
apiVersion: v1
entries:
  redis:
    - name: redis
      description: Redis service
      version: 0.0.1
`
	dto := indexDTO{}
	// WHEN
	yaml.Unmarshal([]byte(data), &dto)
	// THEN
	require.Len(t, dto.Entries, 1)
	redis, ex := dto.Entries["redis"]
	assert.True(t, ex)
	assert.Len(t, redis, 1)
	v001 := redis[0]
	assert.Equal(t, BundleName("redis"), v001.Name)
	assert.Equal(t, BundleVersion("0.0.1"), v001.Version)
	assert.Equal(t, "Redis service", v001.Description)

}
