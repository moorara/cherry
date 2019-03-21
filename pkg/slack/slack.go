/*
 * Ref:
 *   https://api.slack.com/messaging
 *     https://api.slack.com/messaging/managing
 *     https://api.slack.com/messaging/composing
 *     https://api.slack.com/messaging/interactivity
 *     https://api.slack.com/reference/messaging/payload
 *     https://api.slack.com/reference/messaging/blocks
 *     https://api.slack.com/reference/messaging/block-elements
 *     https://api.slack.com/reference/messaging/interactive-components
 *     https://api.slack.com/reference/messaging/composition-objects
 */

package slack

import (
	"net/http"
	"time"
)

type (
	// Slack client
	Slack struct {
		client *http.Client
	}

	// Message is the model for a Slack message
	Message struct {
		Text     string  `json:"text"`
		Markdown bool    `json:"mrkdwn"`
		Thread   string  `json:"thread_ts"`
		Blocks   []Block `json:"blocks"`
	}
)

// New creates a new instance of Slack
func New(timeout time.Duration) *Slack {
	transport := &http.Transport{}
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	return &slack{
		client: client,
	}
}

// SendMessage sends a message to Slack
func (s *Slack) SendMessage(webhook string, message Message) error {
	return nil
}
