package testutil

import (
	"time"

	"bwawan.com/openuss/internal/api/scdussv1"
)

func NewVolumes4D() []scdussv1.Volume4D {
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
				OutlineCircle: &scdussv1.Circle{
					Radius: &scdussv1.Radius{
						Value: 100,
						Units: "M",
					},
					Center: &scdussv1.LatLngPoint{
						Lng: -80.6,
						Lat: 37.2,
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
