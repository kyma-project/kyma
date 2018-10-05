package gqlschema

type TopicEntry struct {
	ContentType string
	ID          string
	Sections    []Section
}

type Section struct {
	Name      string
	Anchor    string
	TopicType string
	Titles    []Title
}

type Title struct {
	Name   string
	Anchor string
	Titles []Title
}
