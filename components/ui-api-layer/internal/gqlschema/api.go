package gqlschema

type API struct {
	Name                   string
	Hostname               string
	Service                Service
	AuthenticationPolicies []AuthenticationPolicy
}
