package flightplanning

import (
	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/dss"
)

type Handler struct {
	DSS dss.USSAuthority
	DB  db.DB
}
