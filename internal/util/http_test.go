package util

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPError(t *testing.T) {
	tests := []struct {
		name          string
		request       *http.Request
		statusCode    int
		body          string
		expectedError string
	}{
		{
			"400",
			&http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "/",
				},
			},
			http.StatusBadRequest,
			"Invalid request",
			"GET / 400: Invalid request",
		},
		{
			"500",
			&http.Request{
				Method: "POST",
				URL: &url.URL{
					Path: "/",
				},
			},
			http.StatusInternalServerError,
			"Internal error",
			"POST / 500: Internal error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			br := strings.NewReader(tc.body)
			rc := ioutil.NopCloser(br)

			res := &http.Response{
				Request:    tc.request,
				StatusCode: tc.statusCode,
				Body:       rc,
			}

			err := NewHTTPError(res)
			assert.Equal(t, tc.request, err.Request)
			assert.Equal(t, tc.statusCode, err.StatusCode)
			assert.Equal(t, tc.body, err.Body)

			var e error = err
			assert.Equal(t, tc.expectedError, e.Error())
		})
	}
}
