package db

import (
	"iter"
	"maps"
	"uuid"

	"bwawan.com/openuss/internal/api/scdussv1"
)

type InMemoryDB struct {
	Intents map[scdussv1.EntityID]OperationalIntent
	Flights map[uuid.UUID]FlightPlan
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		Intents: map[scdussv1.EntityID]OperationalIntent{},
		Flights: map[uuid.UUID]FlightPlan{},
	}
}

func (s *InMemoryDB) SaveIntent(intent OperationalIntent) error {
	s.Intents[intent.EntityID] = intent
	return nil
}

func (s *InMemoryDB) GetIntent(id scdussv1.EntityID) *OperationalIntent {
	if e, ok := s.Intents[id]; ok {
		return &e
	}
	return nil
}

func (s *InMemoryDB) DeleteIntent(id scdussv1.EntityID) {
	delete(s.Intents, id)
}

func (s *InMemoryDB) GetAllIntents() iter.Seq[OperationalIntent] {
	return maps.Values(s.Intents)
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

func (s *InMemoryDB) GetAllFlights() iter.Seq[FlightPlan] {
	return maps.Values(s.Flights)
}
