package handler

import (
	"io/ioutil"
	"net/http"
)

// EchoHandler is an http handler for echoing request body
func EchoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
