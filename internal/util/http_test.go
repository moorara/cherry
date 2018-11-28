package util

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			"400",
			http.StatusBadRequest,
			"Invalid request",
		},
		{
			"500",
			http.StatusInternalServerError,
			"Internal error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			br := strings.NewReader(tc.body)
			rc := ioutil.NopCloser(br)

			res := &http.Response{
				StatusCode: tc.statusCode,
				Body:       rc,
			}

			err := NewHTTPError(res)

			var e error = err
			assert.NotEmpty(t, e.Error())

			assert.Equal(t, tc.statusCode, err.StatusCode)
			assert.Equal(t, tc.body, err.Body)
		})
	}
}
