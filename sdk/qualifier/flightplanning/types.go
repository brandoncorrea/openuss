package flightplanning

import "bwawan.com/openuss/sdk/api/scdussv1"

type UsageState string

const (
	UsageStatePlanned UsageState = "Planned"
	UsageStateInUse   UsageState = "InUse"
	UsageStateClosed  UsageState = "Closed"
)

type UASState string

const (
	UASStateNominal      UASState = "Nominal"
	UASStateOffNominal   UASState = "OffNominal"
	UASStateContingent   UASState = "Contingent"
	UASStateNotSpecified UASState = "NotSpecified"
)

type FlightPlanStatus string

const (
	FlightPlanStatusNotPlanned FlightPlanStatus = "NotPlanned"
	FlightPlanStatusPlanned    FlightPlanStatus = "Planned"
	FlightPlanStatusOkToFly    FlightPlanStatus = "OkToFly"
	FlightPlanStatusOffNominal FlightPlanStatus = "OffNominal"
	FlightPlanStatusClosed     FlightPlanStatus = "Closed"
)

// PlanningActivityResult
type PlanningActivityResult string

const (
	PlanningActivityResultCompleted    PlanningActivityResult = "Completed"
	PlanningActivityResultRejected     PlanningActivityResult = "Rejected"
	PlanningActivityResultFailed       PlanningActivityResult = "Failed"
	PlanningActivityResultNotSupported PlanningActivityResult = "NotSupported"
)

type ExecutionStyle string

const (
	ExecutionStyleHypothetical ExecutionStyle = "Hypothetical"
	ExecutionStyleIfAllowed    ExecutionStyle = "IfAllowed"
	ExecutionStyleInReality    ExecutionStyle = "InReality"
)

type ServiceStatus string

const (
	ServiceStatusReady    ServiceStatus = "Ready"
	ServiceStatusStarting ServiceStatus = "Starting"
)

type ClearAreaResponse struct {
	Outcome ClearAreaOutcome `json:"outcome"`
}

type ClearAreaOutcome struct {
	Success bool `json:"success"`
}

// TODO(gap): Missing api_name and api_version
type StatusResponse struct {
	Status ServiceStatus `json:"status"`
}

type UpsertFlightPlanRequest struct {
	RequestID      scdussv1.UUIDv4Format `json:"request_id"`
	ExecutionStyle ExecutionStyle        `json:"execution_style"`
	FlightPlan     FlightPlan            `json:"flight_plan"`
}

// TODO(gap): Missing Fields: notes, as_planned, flight_id, includes_advisories, queries(?), log_messages(?)
type FlightPlanResponse struct {
	FlightPlanStatus FlightPlanStatus       `json:"flight_plan_status"`
	PlanningResult   PlanningActivityResult `json:"planning_result"`
}

type BasicFlightPlanInformation struct {
	UsageState    UsageState          `json:"usage_state"`
	UASState      UASState            `json:"uas_state"`
	Area          []scdussv1.Volume4D `json:"area"`
	UTMIdentifier *string             `json:"utm_id"`
	Description   *string             `json:"description"`
}

type ASTMF354821OpIntentInformation struct {
	Priority int `json:"priority"`
}

type FlightPlan struct {
	BasicInformation BasicFlightPlanInformation     `json:"basic_information"`
	F3548            ASTMF354821OpIntentInformation `json:"astm_f3548_21"`
}
