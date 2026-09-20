package flightplanning

import "bwawan.com/openuss/internal/api/scdussv1"

type FlightPlanBasicInformation struct {
	UsageState    string              `json:"usage_state"`
	UASState      string              `json:"uas_state"`
	Area          []scdussv1.Volume4D `json:"area"`
	UTMIdentifier *string             `json:"utm_id"`
	Description   *string             `json:"description"`
}

type F3548 struct {
	Priority int `json:"priority"`
}

type FlightPlan struct {
	BasicInformation FlightPlanBasicInformation `json:"basic_information"`
	F3548            F3548                      `json:"astm_f3548_21"`
}
