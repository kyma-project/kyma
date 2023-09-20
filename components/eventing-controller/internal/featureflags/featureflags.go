package featureflags

const FalseValue = false

//nolint:gochecknoglobals // This is global only inside the package.
var f = &flags{
	eventingWebhookAuthEnabled: FalseValue,
}

type flags struct {
	eventingWebhookAuthEnabled bool
	natsProvisioningEnabled    bool
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

// SetNATSProvisioningEnabled enable/disable the NATS resources provisioning feature flag.
func SetNATSProvisioningEnabled(enabled bool) {
	f.natsProvisioningEnabled = enabled
}

// IsNATSProvisioningEnabled returns true if the NATS resources provisioning feature flag is enabled,
// otherwise returns false.
func IsNATSProvisioningEnabled() bool {
	return f.natsProvisioningEnabled
}
