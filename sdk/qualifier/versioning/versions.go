package versioning

import (
	"net/http"

	"bwawan.com/openuss/sdk/api"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (*Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	api.WriteJSON(w, http.StatusOK, GetVersionResponse{
		// TODO(gap): Identity is simply echoed back to the caller
		SystemIdentity: r.PathValue("system_identity"),
		// TODO(gap): The suite wants system_version to be a populated string, but doesn't enforce anything after that.
		SystemVersion: "blah",
	})
}
