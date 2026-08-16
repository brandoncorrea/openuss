package flightplanning

import (
	"time"

	"bwawan.com/openuss/internal/api/scdussv1"
)

type FlightPlanBasicInformation struct {
	UsageState  string              `json:"usage_state"`
	UasState    string              `json:"uas_state"`
	Area        []scdussv1.Volume4D `json:"area"`
	UtmId       *string             `json:"utm_id"`
	Description *string             `json:"description"`
}

type AstmF3548v21 struct {
	Priority int `json:"priority"`
}

type FlightPlan struct {
	BasicInformation FlightPlanBasicInformation `json:"basic_information"`
	Astm             AstmF3548v21               `json:"astm_f3548_21"`
}

func (flight *FlightPlan) StartTime() time.Time {
	// TODO(gap): Nothing validates a zero-area or multi-area flight plan
	return rfc3339(flight.BasicInformation.Area[0].TimeStart.Value)
}

func (flight *FlightPlan) EndTime() time.Time {
	// TODO(gap): Nothing validates a zero-area or multi-area flight plan
	return rfc3339(flight.BasicInformation.Area[0].TimeEnd.Value)
}

func rfc3339(s string) time.Time {
	// TODO(gap): Nothing validates a malformed timestamp
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
