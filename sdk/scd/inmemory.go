package scd

import (
	"context"
	"maps"
	"slices"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

// TODO(gap): Not safe for concurrent use
type InMemoryIntentStore struct {
	intents map[scdussv1.EntityID]OperationalIntent
}

func NewInMemoryIntentStore() *InMemoryIntentStore {
	return &InMemoryIntentStore{
		intents: map[scdussv1.EntityID]OperationalIntent{},
	}
}

func (s *InMemoryIntentStore) Get(ctx context.Context, id scdussv1.EntityID) (OperationalIntent, error) {
	mustHaveContext(ctx)
	intent, ok := s.intents[id]
	if !ok {
		return OperationalIntent{}, ErrNotFound
	}
	return intent, nil
}

func (s *InMemoryIntentStore) List(context.Context) ([]OperationalIntent, error) {
	return slices.Collect(maps.Values(s.intents)), nil
}

func (s *InMemoryIntentStore) Upsert(_ context.Context, intent OperationalIntent) error {
	s.intents[intent.EntityID] = intent
	return nil
}

func (s *InMemoryIntentStore) Delete(ctx context.Context, id scdussv1.EntityID) error {
	mustHaveContext(ctx)
	delete(s.intents, id)
	return nil
}

func mustHaveContext(ctx context.Context) {
	if ctx == nil {
		panic("inmemory dss: missing Context")
	}
}
