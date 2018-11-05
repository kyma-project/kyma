package content

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	funk "github.com/thoas/go-funk"
)

type topicsConverter struct{}

func (c *topicsConverter) ToGQL(in []gqlschema.TopicEntry) *gqlschema.JSON {
	if in == nil {
		return nil
	}

	result := make(gqlschema.JSON)

	result["topics"] = in

	return &result
}

func (c *topicsConverter) getUniqueTypes(documents []storage.Document) []string {
	r := funk.Map(documents, func(x storage.Document) string {
		return x.Type
	})

	return funk.UniqString(r.([]string))
}
func (c *topicsConverter) adjustTypes(documents []storage.Document) []storage.Document {
	docs := []storage.Document{}

	for _, d := range documents {
		if d.Type == "" {
			d.Type = d.Title
		}
		docs = append(docs, d)
	}

	return docs
}

func (r *topicsConverter) getAnchor(input string) string {
	anchor := strings.TrimSpace(input)
	anchor = strings.Replace(input, " ", "-", -1)

	return strings.ToLower(anchor)
}

func (c *topicsConverter) ExtractSection(documents []storage.Document, internal bool) ([]gqlschema.Section, error) {
	var topics []gqlschema.Section

	documents = c.adjustTypes(documents)

	types := c.getUniqueTypes(documents)

	for _, t := range types {
		entry := gqlschema.Section{}

		doc, ok := funk.Filter(documents, func(x storage.Document) bool {
			if internal == true {
				return x.Type == t
			} else {
				return x.Type == t && x.Internal == false
			}

		}).([]storage.Document)

		if !ok {
			return []gqlschema.Section{}, fmt.Errorf("while converting object from interface to []Docs")
		}

		//if there is only one document of certain type, title equals to the title of this document, otherwise - we go deeper
		// TODO Find better way to do this

		if len(doc) == 1 {
			if doc[0].Type == doc[0].Title {
				entry = gqlschema.Section{Name: doc[0].Title, Anchor: c.getAnchor(doc[0].Title)}
			} else {
				entry = gqlschema.Section{Name: doc[0].Type, Anchor: c.getAnchor(doc[0].Type), TopicType: doc[0].Type}
				tit := gqlschema.Title{Name: doc[0].Title, Anchor: c.getAnchor(doc[0].Title)}
				entry.Titles = append(entry.Titles, tit)
			}
			topics = append(topics, entry)
		} else {
			entry = gqlschema.Section{Name: t, Anchor: c.getAnchor(t), TopicType: t}

			for _, d := range doc {
				titChild := gqlschema.Title{Name: d.Title, Anchor: c.getAnchor(d.Title)}
				entry.Titles = append(entry.Titles, titChild)
			}
			topics = append(topics, entry)
		}

	}

	return topics, nil
}
