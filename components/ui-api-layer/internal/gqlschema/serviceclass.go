package gqlschema

import "time"

type ServiceClass struct {
	Name                string
	ExternalName        string
	DisplayName         *string
	CreationTimestamp   time.Time
	Description         string
	ImageUrl            *string
	DocumentationUrl    *string
	ProviderDisplayName *string
	Tags                []string
	activated           bool
	apiSpec             *JSON
	asyncApiSpec        *JSON
	content             *JSON
}
