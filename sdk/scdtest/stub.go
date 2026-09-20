package scdtest

import (
	"context"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/scd"
)

type Stub struct {
	CreateFn func(context.Context, scd.IntentParams) (scd.OperationalIntent, error)
	UpdateFn func(context.Context, scdussv1.EntityID, scd.IntentParams) (scd.OperationalIntent, error)
	DeleteFn func(context.Context, scdussv1.EntityID) error
}

func (s Stub) CreateOperationalIntent(
	ctx context.Context,
	params scd.IntentParams,
) (scd.OperationalIntent, error) {
	return s.CreateFn(ctx, params)
}

func (s Stub) UpdateOperationalIntent(
	ctx context.Context,
	id scdussv1.EntityID,
	params scd.IntentParams,
) (scd.OperationalIntent, error) {
	return s.UpdateFn(ctx, id, params)
}

func (s Stub) DeleteOperationalIntent(ctx context.Context, id scdussv1.EntityID) error {
	return s.DeleteFn(ctx, id)
}
