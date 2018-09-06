package gqlschema

import "time"

type ServiceClass struct {
	Name                string
	ExternalName        string
	DisplayName         *string
	CreationTimestamp   time.Time
	Description         string
	LongDescription     *string
	ImageUrl            *string
	DocumentationUrl    *string
	SupportUrl          *string
	ProviderDisplayName *string
	Tags                []string
	activated           bool
	apiSpec             *JSON
	asyncApiSpec        *JSON
	content             *JSON
}
