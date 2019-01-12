package handler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEchoHandler(t *testing.T) {
	tests := []struct {
		body string
	}{
		{"Hello, World!"},
		{"My name is Gohper!"},
	}

	for _, tc := range tests {

		r := httptest.NewRequest("GET", "http://example.com/foo/bar", strings.NewReader(tc.body))
		w := httptest.NewRecorder()
		EchoHandler(w, r)

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			t.Fail()
		}

		if tc.body != string(body) {
			t.Fail()
		}
	}
}
