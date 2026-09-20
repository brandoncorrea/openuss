package db

import (
	"maps"
	"slices"
	"uuid"
)

type InMemoryDB struct {
	Flights map[uuid.UUID]FlightPlan
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		Flights: map[uuid.UUID]FlightPlan{},
	}
}

func (s *InMemoryDB) SaveFlight(flight FlightPlan) error {
	s.Flights[flight.ID] = flight
	return nil
}

func (s *InMemoryDB) GetFlight(id uuid.UUID) *FlightPlan {
	if e, ok := s.Flights[id]; ok {
		return &e
	}
	return nil
}

func (s *InMemoryDB) DeleteFlight(id uuid.UUID) {
	delete(s.Flights, id)
}

func (s *InMemoryDB) GetAllFlights() []FlightPlan {
	return slices.Collect(maps.Values(s.Flights))
}
