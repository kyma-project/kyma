package gqlschema

import "time"

type ServiceClass struct {
	Name                string    `json:"name"`
	Environment         string    `json:"environment"`
	ExternalName        string    `json:"externalName"`
	DisplayName         *string   `json:"displayName"`
	CreationTimestamp   time.Time `json:"creationTimestamp"`
	Description         string    `json:"description"`
	LongDescription     *string   `json:"longDescription"`
	ImageURL            *string   `json:"imageUrl"`
	DocumentationURL    *string   `json:"documentationUrl"`
	SupportURL          *string   `json:"supportUrl"`
	ProviderDisplayName *string   `json:"providerDisplayName"`
	Tags                []string  `json:"tags"`
	Labels              Labels    `json:"labels"`
}
