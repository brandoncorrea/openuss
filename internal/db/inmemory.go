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

func (db *InMemoryDB) SaveIntent(intent OperationalIntent) error {
	db.Intents[intent.EntityID] = intent
	return nil
}

func (db *InMemoryDB) GetIntent(id scdussv1.EntityID) *OperationalIntent {
	if e, ok := db.Intents[id]; ok {
		return &e
	}
	return nil
}

func (db *InMemoryDB) DeleteIntent(id scdussv1.EntityID) {
	delete(db.Intents, id)
}

func (db *InMemoryDB) GetAllIntents() iter.Seq[OperationalIntent] {
	return maps.Values(db.Intents)
}

func (db *InMemoryDB) SaveFlight(flight FlightPlan) error {
	db.Flights[flight.Id] = flight
	return nil
}

func (db *InMemoryDB) GetFlight(id uuid.UUID) *FlightPlan {
	if e, ok := db.Flights[id]; ok {
		return &e
	}
	return nil
}

func (db *InMemoryDB) DeleteFlight(id uuid.UUID) {
	delete(db.Flights, id)
}

func (db *InMemoryDB) GetAllFlights() iter.Seq[FlightPlan] {
	return maps.Values(db.Flights)
}
