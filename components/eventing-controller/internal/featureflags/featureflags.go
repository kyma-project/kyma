package featureflags

//nolint:gochecknoglobals // This is global only inside the package.
var f = &flags{
	eventingWebhookAuthEnabled: false,
}

type flags struct {
	eventingWebhookAuthEnabled bool
}

// SetEventingWebhookAuthEnabled enable/disable the Eventing webhook auth feature flag.
func SetEventingWebhookAuthEnabled(enabled bool) {
	f.eventingWebhookAuthEnabled = enabled
}

// IsEventingWebhookAuthEnabled returns true if the Eventing webhook auth feature flag is enabled,
// otherwise returns false.
func IsEventingWebhookAuthEnabled() bool {
	return f.eventingWebhookAuthEnabled
}
