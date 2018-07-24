package content_test

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopicsResolver_TopicsQuery(t *testing.T) {

	t.Run("Success with internal true", func(t *testing.T) {
		cnt := &storage.Content{
			Raw: map[string]interface{}{
				"Description": "data",
				"DisplayName": "data",
				"Docs": []map[string]interface{}{
					{"Order": "",
						"Source":   "",
						"Title":    "",
						"Type":     "",
						"Internal": true,
					},
				},
				"ID":       "data",
				"Internal": true,
			},
			Data: storage.ContentData{
				Description: "data",
				DisplayName: "data",
				Docs: []storage.Document{
					{
						Order:    "",
						Source:   "",
						Title:    "",
						Type:     "",
						Internal: true,
					},
				},
				ID: "data",
			},
		}

		expectedResult := []gqlschema.TopicEntry{
			{ContentType: "test1", ID: "test1", Sections: []gqlschema.Section{
				{Titles: nil, TopicType: "Test"},
			}},
			{ContentType: "test2", ID: "test2", Sections: []gqlschema.Section{
				{Titles: nil, TopicType: "Test"},
			}},
			{ContentType: "test3", ID: "test3", Sections: []gqlschema.Section{
				{Titles: nil, TopicType: "Test"},
			}},
		}

		getter := automock.NewContentGetter()
		getter.On("Find", "test1", "test1").Return(cnt, nil)
		getter.On("Find", "test2", "test2").Return(cnt, nil)
		getter.On("Find", "test3", "test3").Return(cnt, nil)

		converter := automock.NewMockTopicsConverter()
		converter.On("ExtractSection", []storage.Document{{Order: "", Source: "", Title: "", Type: "", Internal: true}}, true).Return([]gqlschema.Section{
			{
				Titles:    nil,
				TopicType: "Test",
			},
		}, nil).Times(3)

		defer converter.AssertExpectations(t)

		resolver := content.NewTopicsResolver(getter)
		resolver.SetTopicsConverter(converter)

		internal := true

		inputTopic := []gqlschema.InputTopic{{ID: "test1", Type: "test1"}, {ID: "test2", Type: "test2"}, {ID: "test3", Type: "test3"}}

		result, err := resolver.TopicsQuery(nil, inputTopic, &internal)

		require.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("Success with internal false", func(t *testing.T) {

		cnt := &storage.Content{
			Raw: map[string]interface{}{
				"Description": "data",
				"DisplayName": "data",
				"Docs": []map[string]interface{}{
					{"Order": "",
						"Source":   "",
						"Title":    "",
						"Type":     "",
						"Internal": true,
					},
				},
				"ID":       "data",
				"Internal": true,
			},
			Data: storage.ContentData{
				Description: "data",
				DisplayName: "data",
				Docs: []storage.Document{
					{
						Order:    "",
						Source:   "",
						Title:    "",
						Type:     "",
						Internal: true,
					},
				},
				ID: "data",
			},
		}

		getter := automock.NewContentGetter()
		getter.On("Find", "test1", "test1").Return(cnt, nil)
		getter.On("Find", "test2", "test2").Return(cnt, nil)
		getter.On("Find", "test3", "test3").Return(cnt, nil)

		converter := automock.NewMockTopicsConverter()
		converter.On("ExtractSection", []storage.Document{{Order: "", Source: "", Title: "", Type: "", Internal: true}}, false).Return([]gqlschema.Section{
			{
				Titles:    nil,
				TopicType: "Test",
			},
		}, nil).Times(3)

		defer converter.AssertExpectations(t)

		resolver := content.NewTopicsResolver(getter)
		resolver.SetTopicsConverter(converter)

		inputTopic := []gqlschema.InputTopic{{ID: "test1", Type: "test1"}, {ID: "test2", Type: "test2"}, {ID: "test3", Type: "test3"}}
		result, err := resolver.TopicsQuery(nil, inputTopic, nil)

		expectedResult := []gqlschema.TopicEntry{
			{ContentType: "test1", ID: "test1", Sections: []gqlschema.Section{
				{Titles: nil, TopicType: "Test"},
			}},
			{ContentType: "test2", ID: "test2", Sections: []gqlschema.Section{
				{Titles: nil, TopicType: "Test"},
			}},
			{ContentType: "test3", ID: "test3", Sections: []gqlschema.Section{
				{Titles: nil, TopicType: "Test"},
			}},
		}

		require.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("Error when content is not found", func(t *testing.T) {

		getter := automock.NewContentGetter()
		getter.On("Find", "test1", "test1").Return(nil, errors.New("Test"))

		converter := automock.NewMockTopicsConverter()

		defer converter.AssertExpectations(t)

		resolver := content.NewTopicsResolver(getter)
		resolver.SetTopicsConverter(converter)

		inputTopic := []gqlschema.InputTopic{{ID: "test1", Type: "test1"}}
		result, err := resolver.TopicsQuery(nil, inputTopic, nil)

		require.Error(t, err)
		assert.Nil(t, result)
	})

}
