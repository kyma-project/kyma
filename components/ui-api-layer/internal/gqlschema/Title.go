package gqlschema

type Title struct {
	Name   string
	Anchor string
	Titles []Title
}
