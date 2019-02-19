package main

import (
	"fmt"

	"github.com/go-logr/logr"
	apiSlack "github.com/nlopes/slack"
	"github.com/pkg/errors"
)

// Reporter is a reporter client for slack
type Reporter struct {
	channel  string
	icon     string
	username string
	slackCli apiSlack.Client
	log      logr.Logger
}

// NewReporter creates a new Slack reporter
func NewReporter(slackCli apiSlack.Client, channel, username, icon string) *Reporter {
	return &Reporter{
		channel:  channel,
		icon:     icon,
		username: username,
		slackCli: slackCli,
	}
}

// Report takes a message publish to given channel
func (r *Reporter) Report(msg string, premature bool) error {
	var (
		body apiSlack.Attachment
	)
	header := r.generateHeader()
	if premature {
		body = r.generateBody(msg, true)
	} else {
		body = r.generateBody(msg, false)
	}

	_, _, _, err := r.slackCli.SendMessage(
		r.channel,
		apiSlack.MsgOptionPostMessageParameters(apiSlack.PostMessageParameters{IconEmoji: ":weight_lifter", Markdown: true, Username: r.username}),
		apiSlack.MsgOptionText(header, false),
		apiSlack.MsgOptionAttachments(body),
	)
	if err != nil {
		return errors.Wrap(err, "while sending slack message")
	}

	return nil
}

func (r *Reporter) generateBody(msg string, premature bool) apiSlack.Attachment {
	var (
		blue  = "#007FFF"
		red   = "#FF0022"
		color string
	)

	if premature {
		color = red
	} else {
		color = blue
	}
	body := apiSlack.Attachment{
		Color: color,
		Fields: []apiSlack.AttachmentField{
			{
				Title: "Horizontal Pod Autoscaler Job Details",
				Value: msg,
				Short: true,
			},
		},
	}

	return body
}

func (r *Reporter) generateHeader() string {
	return fmt.Sprintf("*Load test job*")
}
