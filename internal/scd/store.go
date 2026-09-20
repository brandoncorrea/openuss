package scd

import (
	"context"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

type IntentStore interface {
	Get(context.Context, scdussv1.EntityID) (OperationalIntent, error)
	List(context.Context) ([]OperationalIntent, error)
	Upsert(context.Context, OperationalIntent) error
	Delete(context.Context, scdussv1.EntityID) error
}
