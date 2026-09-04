package db

import "bwawan.com/openuss/internal/api/scdussv1"

type DB interface {
	SaveIntent(scdussv1.OperationalIntent) error
	GetIntent(scdussv1.EntityID) *scdussv1.OperationalIntent
	DeleteIntent(scdussv1.EntityID)
}
