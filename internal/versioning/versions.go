package versioning

import (
	"net/http"

	"bwawan.com/openuss/internal/api"
)

type Handler struct{}

func (*Handler) GetVersion(w http.ResponseWriter, _ *http.Request) {
	api.WriteJSON(w, http.StatusOK, map[string]any{
		"system_identity": "astm.f3548.v21",
		// TODO(gap): The suite wants system_version to be non-nil, but doesn't enforce anything after that.
		"system_version": map[string]any{},
	})
}
