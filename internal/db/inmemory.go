package db

import "bwawan.com/openuss/internal/api/scdussv1"

type InMemoryDB struct {
	IntentReferences map[scdussv1.EntityID]scdussv1.OperationalIntent
}

func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		IntentReferences: map[scdussv1.EntityID]scdussv1.OperationalIntent{},
	}
}

func (db *InMemoryDB) SaveIntent(intent scdussv1.OperationalIntent) error {
	db.IntentReferences[intent.Reference.Id] = intent
	return nil
}

func (db *InMemoryDB) GetIntent(id scdussv1.EntityID) *scdussv1.OperationalIntent {
	if e, ok := db.IntentReferences[id]; ok {
		return &e
	}
	return nil
}

func (db *InMemoryDB) DeleteIntent(id scdussv1.EntityID) {
	delete(db.IntentReferences, id)
}
