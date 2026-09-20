package scd

import (
	"context"
	"maps"
	"slices"

	"bwawan.com/openuss/internal/api/scdussv1"
)

type InMemoryIntentStore struct {
	intents map[scdussv1.EntityID]OperationalIntent
}

func NewInMemoryIntentStore() *InMemoryIntentStore {
	return &InMemoryIntentStore{
		intents: map[scdussv1.EntityID]OperationalIntent{},
	}
}

func (s *InMemoryIntentStore) Get(_ context.Context, id scdussv1.EntityID) (OperationalIntent, error) {
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

func (s *InMemoryIntentStore) Delete(_ context.Context, id scdussv1.EntityID) error {
	delete(s.intents, id)
	return nil
}
