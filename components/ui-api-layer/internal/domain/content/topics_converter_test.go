package content

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestTopicsConverter_ToQGL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {

		topics := []gqlschema.Section{
			{
				Titles: []gqlschema.Title{
					{
						Name:   "Title1",
						Anchor: "title1",
						Titles: []gqlschema.Title{
							{
								Name:   "Title2",
								Anchor: "title2",
							},
						},
					},
				},
				TopicType: "test1",
			},
		}

		topOutput := []gqlschema.TopicEntry{
			{
				ContentType: "test1",
				ID:          "id1",
				Sections:    topics,
			},
		}

		expectedResult := &gqlschema.JSON{
			"topics": []gqlschema.TopicEntry{
				{
					ContentType: "test1",
					ID:          "id1",
					Sections:    topics,
				},
			},
		}

		converter := &topicsConverter{}

		result := converter.ToGQL(topOutput)
		assert.Equal(t, expectedResult, result)
	})
}

func TestTopicsConverter_ExtractSection(t *testing.T) {
	t.Run("Topics with internal false", func(t *testing.T) {

		expectedResult := []gqlschema.Section{
			{TopicType: "test1", Titles: []gqlschema.Title{
				{Name: "test1", Titles: []gqlschema.Title{
					{Name: "Test1", Anchor: "test1"},
					{Name: "Test2", Anchor: "test2"},
					{Name: "Test3", Anchor: "test3"},
				}, Anchor: "test1"},
			}},
			{TopicType: "test2", Titles: []gqlschema.Title{
				{Name: "Alone1", Anchor: "alone1"}},
			},
		}

		docs := []storage.Document{{Title: "Test1", Type: "test1"}, {Title: "Test2", Type: "test1"}, {Title: "Test3", Type: "test1"}, {Title: "Alone1", Type: "test2"}}

		converter := &topicsConverter{}

		result, err := converter.ExtractSection(docs, false)
		assert.Nil(t, err)

		assert.Equal(t, expectedResult, result)
	})

	t.Run("Topics with external true", func(t *testing.T) {
		expectedResultInternalFalse := []gqlschema.Section{
			{
				TopicType: "internalTest1",
				Titles: []gqlschema.Title{
					{
						Name: "internalTest1",
						Titles: []gqlschema.Title{
							{Name: "Test1", Anchor: "test1"},
							{Name: "Test2", Anchor: "test2"},
							{Name: "Test3", Anchor: "test3"},
							{Name: "Test4", Anchor: "test4"},
						},
						Anchor: "internaltest1",
					},
				},
			},
		}

		expectedResultInternalTrue := []gqlschema.Section{
			{
				TopicType: "internalTest1",
				Titles: []gqlschema.Title{
					{
						Name:   "internalTest1",
						Anchor: "internaltest1",
						Titles: []gqlschema.Title{
							{
								Name: "Test3", Anchor: "test3"},
							{
								Name:   "Test4",
								Anchor: "test4",
							},
						},
					},
				},
			},
		}

		d1 := storage.Document{Title: "Test1", Type: "internalTest1", Internal: true}
		d2 := storage.Document{Title: "Test2", Type: "internalTest1", Internal: true}
		d3 := storage.Document{Title: "Test3", Type: "internalTest1"}
		d4 := storage.Document{Title: "Test4", Type: "internalTest1"}

		docs := []storage.Document{d1, d2, d3, d4}

		converter := &topicsConverter{}

		result, err := converter.ExtractSection(docs, true)

		assert.Nil(t, err)
		assert.Equal(t, expectedResultInternalFalse, result)

		result, err = converter.ExtractSection(docs, false)

		assert.Nil(t, err)
		assert.Equal(t, expectedResultInternalTrue, result)

	})

	t.Run("Empty docs input", func(t *testing.T) {

		expectedResult := []gqlschema.Section(nil)
		var docs []storage.Document

		converter := &topicsConverter{}
		result, err := converter.ExtractSection(docs, true)

		assert.Nil(t, err)
		assert.Equal(t, expectedResult, result)

	})

}
