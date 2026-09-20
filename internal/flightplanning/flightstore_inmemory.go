package flightplanning

import (
	"maps"
	"slices"
	"uuid"
)

type InMemoryFlightStore struct {
	flights map[uuid.UUID]FlightPlanRecord
}

func NewInMemoryFlightStore() *InMemoryFlightStore {
	return &InMemoryFlightStore{
		flights: map[uuid.UUID]FlightPlanRecord{},
	}
}

func (s *InMemoryFlightStore) Upsert(flight FlightPlanRecord) error {
	s.flights[flight.ID] = flight
	return nil
}

func (s *InMemoryFlightStore) Get(id uuid.UUID) *FlightPlanRecord {
	if e, ok := s.flights[id]; ok {
		return &e
	}
	return nil
}

func (s *InMemoryFlightStore) Delete(id uuid.UUID) {
	delete(s.flights, id)
}

func (s *InMemoryFlightStore) List() []FlightPlanRecord {
	return slices.Collect(maps.Values(s.flights))
}
