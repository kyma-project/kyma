package content_test

import (
	"testing"

	"github.com/kyma-project/kyma/tools/docsbuilder/internal/content"
	"github.com/stretchr/testify/assert"
)

func TestConstructPath(t *testing.T) {

	testCases := []struct {
		content        content.Content
		contentDirPath string
		expectedPath   string
	}{
		{
			content: content.Content{
				Name: "test",
			},
			contentDirPath: "./example",
			expectedPath:   "./example/test",
		},
		{
			content: content.Content{
				Name:      "test",
				Directory: "foo/bar",
			},
			contentDirPath: "./example",
			expectedPath:   "./example/foo/bar",
		},
	}

	for _, tC := range testCases {
		result := content.ConstructPath(tC.content, tC.contentDirPath)
		assert.Equal(t, tC.expectedPath, result)
	}
}
