package scd_test

import (
	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/dss"
	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/scdtest"
)

const ussBaseURL = "http://openuss.localutm"

const squareSideDegrees = 0.001

func sharedCorner() scdussv1.LatLngPoint {
	return scdussv1.LatLngPoint{Lng: -80.6, Lat: 37.2}
}

func distantCorner() scdussv1.LatLngPoint {
	return scdussv1.LatLngPoint{Lng: -80.5, Lat: 37.2}
}

func newSquareVolumes(corner scdussv1.LatLngPoint) []scdussv1.Volume4D {
	volumes := scdtest.NewVolumes4D()
	volumes[0].Volume.OutlineCircle = nil
	volumes[0].Volume.OutlinePolygon = &scdussv1.Polygon{
		Vertices: []scdussv1.LatLngPoint{
			corner,
			{Lng: corner.Lng + squareSideDegrees, Lat: corner.Lat},
			{Lng: corner.Lng + squareSideDegrees, Lat: corner.Lat + squareSideDegrees},
			{Lng: corner.Lng, Lat: corner.Lat + squareSideDegrees},
		},
	}
	return volumes
}

func newIntent() scd.OperationalIntent {
	return scd.OperationalIntent{
		EntityID: scdtest.NewEntityID(),
		Volumes:  newSquareVolumes(sharedCorner()),
		Priority: 2,
	}
}

func newService() (*scd.Service, *dss.InMemoryDSS) {
	dssClient := dss.NewInMemoryDSS()
	service := scd.New(dssClient, nil, scd.NewInMemoryIntentStore(), ussBaseURL)
	return service, dssClient
}
