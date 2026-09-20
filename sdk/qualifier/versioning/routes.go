package versioning

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /versioning/versions/astm.f3548.v21", h.GetVersion)
}
