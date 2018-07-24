package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// SlackClient sends message to slack channel
type SlackClient struct {
	channelID  string
	webhookURL string
	token      string
}

// NewSlackClient returns new instance of SlackClient
func NewSlackClient(cfg SlackClientConfig) *SlackClient {
	return &SlackClient{
		channelID:  cfg.ChannelID,
		webhookURL: cfg.WebhookURL,
		token:      cfg.Token,
	}
}

// Send sends message with given content to slack channel
func (c *SlackClient) Send(header, body, footer, color string) error {
	url := fmt.Sprintf("%s?token=%s", c.webhookURL, c.token)

	payload := payload{
		Channel: c.channelID,
		Text:    header,
		Attachments: []*attachment{
			{
				Color: color,
				Text:  body,
			},
			{
				Text: footer, // if not provided (empty) then this attachment will not be showed in slack channel
			},
		},
	}

	dto, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "while marshaling body to send")
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(dto))
	if err != nil {
		return errors.Wrap(err, "while sending notification to web-hook")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected to get %d status code but got %d status code", http.StatusOK, resp.StatusCode)
	}

	return nil
}

type attachment struct {
	Color  string `json:"color,omitempty"`
	Text   string `json:"text,omitempty"`
	Footer string `json:"footer,omitempty"`
}

type payload struct {
	Channel     string        `json:"channel"`
	Text        string        `json:"text"`
	Attachments []*attachment `json:"attachments,omitempty"`
}
