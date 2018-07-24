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

func (r *topicsConverter) getAnchor(input string) string {
	anchor := strings.TrimSpace(input)
	anchor = strings.Replace(input, " ", "-", -1)

	return strings.ToLower(anchor)
}

func (c *topicsConverter) ExtractSection(documents []storage.Document, internal bool) ([]gqlschema.Section, error) {
	var topics []gqlschema.Section
	types := c.getUniqueTypes(documents)

	for _, t := range types {
		entry := gqlschema.Section{TopicType: t}

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

		//if there is only one document of certain type, title equals to the title of this docuemnt, otherwise - we go deeper
		if len(doc) == 1 {
			tit := gqlschema.Title{Name: doc[0].Title, Anchor: c.getAnchor(doc[0].Title)}
			entry.Titles = append(entry.Titles, tit)
		} else {
			tit := gqlschema.Title{Name: t, Anchor: c.getAnchor(t)}

			for _, d := range doc {
				titChild := gqlschema.Title{Name: d.Title, Anchor: c.getAnchor(d.Title)}
				tit.Titles = append(tit.Titles, titChild)
			}
			entry.Titles = append(entry.Titles, tit)
		}
		topics = append(topics, entry)
	}

	return topics, nil
}
