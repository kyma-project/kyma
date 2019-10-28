package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCounterCollector(t *testing.T) {
	t.Run("should create counter collector", func(t *testing.T) {
		// given
		opts := prometheus.CounterOpts{
			Name: "counter",
			Help: "help",
		}
		// when
		collector, err := NewCounterCollector(opts, []string{"label"})
		//then
		require.NoError(t, err)
		assert.NotNil(t, collector)
	})

	t.Run("should return error if name not specified", func(t *testing.T) {
		// given
		opts := prometheus.CounterOpts{
			Help: "help",
		}
		// when
		collector, err := NewCounterCollector(opts, []string{"label"})
		//then
		require.Error(t, err)
		assert.Nil(t, collector)
	})

	t.Run("should return error if help not specified", func(t *testing.T) {
		// given
		opts := prometheus.CounterOpts{
			Name: "name",
		}
		// when
		collector, err := NewCounterCollector(opts, []string{"label"})
		//then
		require.Error(t, err)
		assert.Nil(t, collector)
	})
}
