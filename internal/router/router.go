package router

import (
	"net/http"

	"bwawan.com/openuss/internal/scd"
)

// New returns a handler covering every route OpenUSS serves.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /versioning/versions/astm.f3548.v21", scd.GetVersion)
	return mux
}
