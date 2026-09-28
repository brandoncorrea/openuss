package versioning

import "net/http"

// TODO(gap): No inbound authentication or scope checks
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /versioning/versions/astm.f3548.v21", h.GetVersion)
}
