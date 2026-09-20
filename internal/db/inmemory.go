package db

import (
	"maps"
	"slices"
	"uuid"
)

type InMemoryFlightStore struct {
	Flights map[uuid.UUID]FlightPlan
}

func NewInMemoryFlightStore() *InMemoryFlightStore {
	return &InMemoryFlightStore{
		Flights: map[uuid.UUID]FlightPlan{},
	}
}

func (s *InMemoryFlightStore) Upsert(flight FlightPlan) error {
	s.Flights[flight.ID] = flight
	return nil
}

func (s *InMemoryFlightStore) Get(id uuid.UUID) *FlightPlan {
	if e, ok := s.Flights[id]; ok {
		return &e
	}
	return nil
}

func (s *InMemoryFlightStore) Delete(id uuid.UUID) {
	delete(s.Flights, id)
}

func (s *InMemoryFlightStore) List() []FlightPlan {
	return slices.Collect(maps.Values(s.Flights))
}
