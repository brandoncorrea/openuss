package scdtest

import (
	"time"
	"uuid"

	"bwawan.com/openuss/sdk/api/scdussv1"
)

func NewEntityID() scdussv1.EntityID {
	return scdussv1.EntityID(uuid.New().String())
}

const squareSideDegrees = 0.001

func NewVolumes4D() []scdussv1.Volume4D {
	return NewSquareVolumes4DAt(scdussv1.LatLngPoint{Lng: -80.6, Lat: 37.2})
}

func NewSquareVolumes4DAt(corner scdussv1.LatLngPoint) []scdussv1.Volume4D {
	timeStart := time.Now().Add(time.Hour)
	timeEnd := timeStart.Add(time.Hour)

	return []scdussv1.Volume4D{
		{
			TimeStart: &scdussv1.Time{
				Value:  timeStart.Format(time.RFC3339Nano),
				Format: "RFC3339",
			},
			TimeEnd: &scdussv1.Time{
				Value:  timeEnd.Format(time.RFC3339Nano),
				Format: "RFC3339",
			},
			Volume: scdussv1.Volume3D{
				OutlinePolygon: &scdussv1.Polygon{
					Vertices: []scdussv1.LatLngPoint{
						corner,
						{Lng: corner.Lng + squareSideDegrees, Lat: corner.Lat},
						{Lng: corner.Lng + squareSideDegrees, Lat: corner.Lat + squareSideDegrees},
						{Lng: corner.Lng, Lat: corner.Lat + squareSideDegrees},
					},
				},
				AltitudeLower: &scdussv1.Altitude{
					Value:     0,
					Reference: "W84",
					Units:     "M",
				},
				AltitudeUpper: &scdussv1.Altitude{
					Value:     100,
					Reference: "W84",
					Units:     "M",
				},
			},
		},
	}
}
