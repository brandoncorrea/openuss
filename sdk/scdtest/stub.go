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

type Call struct {
	EntityID scdussv1.EntityID
	Params   scd.IntentParams
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

func NewCreateStub() (Stub, *Call) {
	created := &Call{}
	stub := Stub{
		CreateFn: func(_ context.Context, params scd.IntentParams) (scd.OperationalIntent, error) {
			entityID := NewEntityID()
			*created = Call{EntityID: entityID, Params: params}
			intent := scd.OperationalIntent{EntityID: entityID, State: params.State}
			return intent, nil
		},
	}
	return stub, created
}

func NewUpdateStub() (Stub, *Call) {
	updated := &Call{}
	stub := Stub{
		UpdateFn: func(
			_ context.Context,
			id scdussv1.EntityID,
			params scd.IntentParams,
		) (scd.OperationalIntent, error) {
			*updated = Call{EntityID: id, Params: params}
			intent := scd.OperationalIntent{EntityID: id, State: params.State}
			return intent, nil
		},
	}
	return stub, updated
}

func NewDeleteStub() (Stub, *Call) {
	deleted := &Call{}
	stub := Stub{
		DeleteFn: func(_ context.Context, id scdussv1.EntityID) error {
			*deleted = Call{EntityID: id}
			return nil
		},
	}
	return stub, deleted
}
