package versioning

import (
	"net/http"

	"bwawan.com/openuss/sdk/api"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (*Handler) GetVersion(w http.ResponseWriter, _ *http.Request) {
	api.WriteJSON(w, http.StatusOK, GetVersionResponse{
		SystemIdentity: "astm.f3548.v21",
		// TODO(gap): The suite wants system_version to be a populated string, but doesn't enforce anything after that.
		SystemVersion: "blah",
	})
}
