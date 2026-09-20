package flightplanning_test

import (
	"testing"
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/internal/wiretest"
	"bwawan.com/openuss/sdk/qualifier/flightplanning"
	"github.com/stretchr/testify/require"
)

func TestUsageState(t *testing.T) {
	require.Equal(t, flightplanning.UsageState("Planned"), flightplanning.UsageStatePlanned)
	require.Equal(t, flightplanning.UsageState("InUse"), flightplanning.UsageStateInUse)
	require.Equal(t, flightplanning.UsageState("Closed"), flightplanning.UsageStateClosed)
}

func TestUASState(t *testing.T) {
	require.Equal(t, flightplanning.UASState("Nominal"), flightplanning.UASStateNominal)
	require.Equal(t, flightplanning.UASState("OffNominal"), flightplanning.UASStateOffNominal)
	require.Equal(t, flightplanning.UASState("Contingent"), flightplanning.UASStateContingent)
	require.Equal(t, flightplanning.UASState("NotSpecified"), flightplanning.UASStateNotSpecified)
}

func TestFlightPlanStatus(t *testing.T) {
	require.Equal(t, flightplanning.FlightPlanStatus("NotPlanned"), flightplanning.FlightPlanStatusNotPlanned)
	require.Equal(t, flightplanning.FlightPlanStatus("Planned"), flightplanning.FlightPlanStatusPlanned)
	require.Equal(t, flightplanning.FlightPlanStatus("OkToFly"), flightplanning.FlightPlanStatusOkToFly)
	require.Equal(t, flightplanning.FlightPlanStatus("OffNominal"), flightplanning.FlightPlanStatusOffNominal)
	require.Equal(t, flightplanning.FlightPlanStatus("Closed"), flightplanning.FlightPlanStatusClosed)
}

func TestPlanningActivityResult(t *testing.T) {
	require.Equal(t, flightplanning.PlanningActivityResult("Completed"), flightplanning.PlanningActivityResultCompleted)
	require.Equal(t, flightplanning.PlanningActivityResult("Rejected"), flightplanning.PlanningActivityResultRejected)
	require.Equal(t, flightplanning.PlanningActivityResult("Failed"), flightplanning.PlanningActivityResultFailed)
	require.Equal(t, flightplanning.PlanningActivityResult("NotSupported"), flightplanning.PlanningActivityResultNotSupported)
}

func TestExecutionStyle(t *testing.T) {
	require.Equal(t, flightplanning.ExecutionStyle("Hypothetical"), flightplanning.ExecutionStyleHypothetical)
	require.Equal(t, flightplanning.ExecutionStyle("IfAllowed"), flightplanning.ExecutionStyleIfAllowed)
	require.Equal(t, flightplanning.ExecutionStyle("InReality"), flightplanning.ExecutionStyleInReality)
}

func TestServiceStatus(t *testing.T) {
	require.Equal(t, flightplanning.ServiceStatus("Ready"), flightplanning.ServiceStatusReady)
	require.Equal(t, flightplanning.ServiceStatus("Starting"), flightplanning.ServiceStatusStarting)
}

func TestClearAreaResponse(t *testing.T) {
	response := flightplanning.ClearAreaResponse{
		Outcome: flightplanning.ClearAreaOutcome{
			Success: true,
		},
	}
	json := map[string]any{
		"outcome": map[string]any{
			"success": true,
		},
	}
	wiretest.RequireJSONEncoding(t, json, response)
}

func TestStatusResponse(t *testing.T) {
	response := flightplanning.StatusResponse{
		Status: flightplanning.ServiceStatusReady,
	}
	json := map[string]any{"status": string(flightplanning.ServiceStatusReady)}
	wiretest.RequireJSONEncoding(t, json, response)
}

func TestUpsertFlightPlanRequest(t *testing.T) {
	requestID := uuid.New().String()
	body := flightplanning.UpsertFlightPlanRequest{
		RequestID:      scdussv1.UUIDv4Format(requestID),
		ExecutionStyle: "the-execution-style",
		FlightPlan: flightplanning.FlightPlan{
			BasicInformation: flightplanning.BasicFlightPlanInformation{
				UsageState:    flightplanning.UsageStatePlanned,
				UASState:      flightplanning.UASStateNominal,
				Area:          []scdussv1.Volume4D{},
				UTMIdentifier: new("the-utm-id"),
				Description:   new("the-description"),
			},
			F3548: flightplanning.ASTMF354821OpIntentInformation{
				Priority: 2,
			},
		},
	}
	json := map[string]any{
		"request_id":      requestID,
		"execution_style": "the-execution-style",
		"flight_plan": map[string]any{
			"basic_information": map[string]any{
				"usage_state": string(flightplanning.UsageStatePlanned),
				"uas_state":   string(flightplanning.UASStateNominal),
				"area":        []any{},
				"utm_id":      "the-utm-id",
				"description": "the-description",
			},
			"astm_f3548_21": map[string]any{
				"priority": float64(2),
			},
		},
	}
	wiretest.RequireJSONEncoding(t, json, body)
}

func TestFlightPlanResponse(t *testing.T) {
	response := flightplanning.FlightPlanResponse{
		PlanningResult:   flightplanning.PlanningActivityResultCompleted,
		FlightPlanStatus: flightplanning.FlightPlanStatusPlanned,
	}
	json := map[string]any{
		"planning_result":    string(flightplanning.PlanningActivityResultCompleted),
		"flight_plan_status": string(flightplanning.FlightPlanStatusPlanned),
	}
	wiretest.RequireJSONEncoding(t, json, response)
}
