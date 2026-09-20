package scd

import "bwawan.com/openuss/sdk/api/scdussv1"

type IntentParams struct {
	Volumes  []scdussv1.Volume4D
	State    scdussv1.OperationalIntentState
	Priority scdussv1.Priority
}
