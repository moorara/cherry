/*
 * Ref:
 *   https://docs.rollbar.com/reference
 *     https://docs.rollbar.com/reference#post-deploy
 */

package rollbar

import (
	"net/http"
	"time"
)

type (
	// Rollbar client
	Rollbar struct {
		client *http.Client
	}

	// Status is the type for deploy status
	Status string

	// Deploy is the model for a deploy event
	Deploy struct {
		AccessToken     string `json:"access_token"`     // Required, post_server_item access token
		Environment     string `json:"environment"`      // Required
		Revision        string `json:"revision"`         // Required, Git commit SHA
		Status          Status `json:"status"`           // Optional, default: succeeded
		RollbarUsername string `json:"rollbar_username"` // Optional
		LocalUsername   string `json:"local_username"`   // Optional
		Comment         string `json:"comment"`          // Optional
	}
)

const (
	// StatusStarted represents a deployment that started
	StatusStarted = "started"
	// StatusSucceeded represents a deployment that finished successfully
	StatusSucceeded = "succeeded"
	// StatusFailed represents a deployment that failed
	StatusFailed = "failed"
	// StatusTimeout represents a deployment that timed out
	StatusTimeout = "timed_out"
)

// New creates a new instance of Rollbar
func New(timeout time.Duration) *Rollbar {
	transport := &http.Transport{}
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	return &Rollbar{
		client: client,
	}
}

// ReportDeploy reports a deploy to Rollbar
func (r *Rollbar) ReportDeploy(deploy Deploy) error {
	return nil
}

// UpdateDeploy updates a deploy that is started
func (r *Rollbar) UpdateDeploy(id, token string, status Status, comment string) error {
	return nil
}
