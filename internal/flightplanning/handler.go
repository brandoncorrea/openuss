package flightplanning

import (
	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/sdk/scd"
)

type Handler struct {
	SCD *scd.Service
	DB  db.DB
}

func New(service *scd.Service, db db.DB) *Handler {
	return &Handler{
		SCD: service,
		DB:  db,
	}
}
