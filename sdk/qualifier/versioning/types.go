package versioning

type GetVersionResponse struct {
	SystemIdentity string `json:"system_identity"`
	SystemVersion  string `json:"system_version"`
}
