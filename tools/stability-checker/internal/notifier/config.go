package notifier

// SlackClientConfig holds configuration for slack client
type SlackClientConfig struct {
	ChannelID  string
	WebhookURL string
	Token      string
}
