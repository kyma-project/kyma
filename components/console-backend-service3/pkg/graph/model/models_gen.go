// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type BackendModule struct {
	Name string `json:"name"`
}

type MicroFrontend struct {
	Name string `json:"name"`
}

type ResourceAttributes struct {
	Verb            string  `json:"verb"`
	APIGroup        *string `json:"apiGroup"`
	APIVersion      *string `json:"apiVersion"`
	Resource        *string `json:"resource"`
	ResourceArg     *string `json:"resourceArg"`
	Subresource     string  `json:"subresource"`
	NameArg         *string `json:"nameArg"`
	NamespaceArg    *string `json:"namespaceArg"`
	IsChildResolver bool    `json:"isChildResolver"`
}
