package handlers

import (
	"net/http"
)

// healthcheck endpoint
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
