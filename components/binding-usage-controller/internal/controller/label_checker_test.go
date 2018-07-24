package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testedLabeler struct {
	labels map[string]string
}

func (l *testedLabeler) Labels() map[string]string {
	return l.labels
}

func TestDetectLabelsConflictsNotFound(t *testing.T) {
	// then
	source := &testedLabeler{
		labels: map[string]string{
			"key": "val",
		},
	}

	// when
	conflicts, found := detectLabelsConflicts(source, map[string]string{
		"key1": "val",
	})

	// then
	assert.False(t, found)
	assert.Empty(t, conflicts)
}

func TestDetectLabelsConflictsFound(t *testing.T) {
	// given
	fixLabels := map[string]string{
		"key": "val",
	}

	source := &testedLabeler{
		labels: fixLabels,
	}

	// when
	conflicts, found := detectLabelsConflicts(source, fixLabels)

	// then
	assert.True(t, found)
	assert.Len(t, conflicts, 1)
	assert.Equal(t, conflicts[0], "key")
}
