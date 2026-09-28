package scd

import (
	"time"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

type IntentParams struct {
	Volumes  []scdussv1.Volume4D
	State    scdussv1.OperationalIntentState
	Priority scdussv1.Priority
}

func (p IntentParams) validate() error {
	if p.State == scdussv1.OperationalIntentState_Nonconforming {
		return ErrNotSupported
	}
	if p.isTooEager() || p.hasEnded() {
		return ErrRejected
	}
	return nil
}

const planningHorizon = 30 * 24 * time.Hour

func (p IntentParams) isTooEager() bool {
	return time.Now().Add(planningHorizon).Before(p.startTime())
}

func (p IntentParams) hasEnded() bool {
	return p.endTime().Before(time.Now())
}

func (p IntentParams) startTime() time.Time {
	// TODO(gap): Nothing validates a zero-area or multi-area flight plan
	return rfc3339(p.Volumes[0].TimeStart.Value)
}

func (p IntentParams) endTime() time.Time {
	// TODO(gap): Nothing validates a zero-area or multi-area flight plan
	return rfc3339(p.Volumes[0].TimeEnd.Value)
}

func rfc3339(s string) time.Time {
	// TODO(gap): Nothing validates a malformed timestamp
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
