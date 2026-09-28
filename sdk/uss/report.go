package uss

import (
	"encoding/json/v2"
	"net/http"
	"uuid"

	"bwawan.com/openuss/sdk/api"
	"bwawan.com/openuss/sdk/api/scdussv1"
)

// TODO(gap): Simply echos the report with a random ID
func (*Handler) MakeUSSReport(w http.ResponseWriter, r *http.Request) {
	var report scdussv1.ErrorReport
	// TODO(gap): What happens if malformed JSON is sent?
	json.UnmarshalRead(r.Body, &report)
	report.ReportId = new(uuid.New().String())
	api.WriteJSON(w, http.StatusCreated, report)
}
