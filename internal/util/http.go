package util

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type (
	// HTTPError is an error for http requests
	HTTPError struct {
		StatusCode int
		Body       string
	}

	// HTTPResult is the result of an http request
	HTTPResult struct {
		Res *http.Response
		Err error
	}
)

// NewHTTPError creates a new instance of HTTPError
func NewHTTPError(res *http.Response) *HTTPError {
	err := &HTTPError{
		StatusCode: res.StatusCode,
	}

	if res.Body != nil {
		if data, e := ioutil.ReadAll(res.Body); e == nil {
			err.Body = string(data)
		}
	}

	return err
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("unexpected status code %d", e.StatusCode)
}
