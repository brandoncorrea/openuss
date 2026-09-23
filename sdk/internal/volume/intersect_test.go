package volume_test

import (
	"testing"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"bwawan.com/openuss/sdk/internal/volume"
	"github.com/peterstace/simplefeatures/geom"
	"github.com/stretchr/testify/require"
)

func volume3D(vertices []scdussv1.LatLngPoint) scdussv1.Volume3D {
	return scdussv1.Volume3D{
		OutlinePolygon: &scdussv1.Polygon{
			Vertices: vertices,
		},
	}
}

func volume4Ds(vertices []scdussv1.LatLngPoint, lower, upper float64) []scdussv1.Volume4D {
	return []scdussv1.Volume4D{
		{
			Volume: scdussv1.Volume3D{
				OutlinePolygon: &scdussv1.Polygon{
					Vertices: vertices,
				},
				AltitudeLower: &scdussv1.Altitude{
					Value: lower,
				},
				AltitudeUpper: &scdussv1.Altitude{
					Value: upper,
				},
			},
		},
	}
}

func TestGeometryOfSinglePoint(t *testing.T) {
	volume3D := volume3D([]scdussv1.LatLngPoint{
		{Lng: 0, Lat: 0},
	})

	line := geom.NewLineString(geom.NewSequence([]float64{0, 0, 0, 0}, geom.DimXY))
	expected := geom.NewPolygon([]geom.LineString{line}).AsGeometry()
	require.Equal(t, expected, volume.Geometry(volume3D))
}

func TestGeometryOfThreePoints(t *testing.T) {
	volume3D := volume3D([]scdussv1.LatLngPoint{
		{Lng: 0, Lat: 0},
		{Lng: 0, Lat: 1},
		{Lng: 1, Lat: 2},
	})

	line := geom.NewLineString(geom.NewSequence([]float64{0, 0, 0, 1, 1, 2, 0, 0}, geom.DimXY))
	expected := geom.NewPolygon([]geom.LineString{line}).AsGeometry()
	require.Equal(t, expected, volume.Geometry(volume3D))
}

func TestVolumesThatShareBorderIntersect(t *testing.T) {
	volume1 := volume4Ds([]scdussv1.LatLngPoint{
		{Lng: 0, Lat: 0},
		{Lng: 0, Lat: 1},
		{Lng: 1, Lat: 0},
	}, 0, 100)
	volume2 := volume4Ds([]scdussv1.LatLngPoint{
		{Lng: 0, Lat: 0},
		{Lng: 0, Lat: 1},
		{Lng: -1, Lat: 0},
	}, 0, 100)

	require.True(t, volume.VolumesIntersect(volume1, volume2))
}

func TestVolumesThatCrossIntersect(t *testing.T) {
	volume1 := volume4Ds([]scdussv1.LatLngPoint{
		{Lng: 0, Lat: 0},
		{Lng: 0, Lat: 1},
		{Lng: 1, Lat: 0},
	}, 0, 100)
	volume2 := volume4Ds([]scdussv1.LatLngPoint{
		{Lng: 0.5, Lat: 0},
		{Lng: 0.5, Lat: 1},
		{Lng: -1, Lat: 0},
	}, 0, 100)

	require.True(t, volume.VolumesIntersect(volume1, volume2))
}

func TestVolumesSharingSinglePointIntersect(t *testing.T) {
	volume1 := volume4Ds([]scdussv1.LatLngPoint{
		{Lng: 0, Lat: 0},
		{Lng: 0, Lat: 1},
		{Lng: 1, Lat: 0},
	}, 0, 100)
	volume2 := volume4Ds([]scdussv1.LatLngPoint{
		{Lng: 0, Lat: 0},
		{Lng: 0, Lat: -1},
		{Lng: -1, Lat: 0},
	}, 0, 100)

	require.True(t, volume.VolumesIntersect(volume1, volume2))
}

func TestVolumesDoNotCrossOrIntersectHorizontally(t *testing.T) {
	volume1 := volume4Ds([]scdussv1.LatLngPoint{
		{Lng: 0, Lat: 0},
		{Lng: 0, Lat: 1},
		{Lng: 1, Lat: 0},
	}, 0, 100)
	volume2 := volume4Ds([]scdussv1.LatLngPoint{
		{Lng: -1, Lat: 0},
		{Lng: -1, Lat: 1},
		{Lng: -2, Lat: 0},
	}, 0, 100)

	require.False(t, volume.VolumesIntersect(volume1, volume2))
}

func TestVolume1FloorIsLowerThanVolume2Ceiling(t *testing.T) {
	volume1 := volume4Ds([]scdussv1.LatLngPoint{
		{Lng: 0, Lat: 0},
		{Lng: 0, Lat: 1},
		{Lng: 1, Lat: 0},
	}, 99, 100)
	volume2 := volume4Ds([]scdussv1.LatLngPoint{
		{Lng: 0, Lat: 0},
		{Lng: 0, Lat: 1},
		{Lng: 1, Lat: 0},
	}, 100, 200)

	require.True(t, volume.VolumesIntersect(volume1, volume2))
}

func TestFloorEqualToCeilingDoesNotIntersect(t *testing.T) {
	volume1 := volume4Ds([]scdussv1.LatLngPoint{
		{Lng: 0, Lat: 0},
		{Lng: 0, Lat: 1},
		{Lng: 1, Lat: 0},
	}, 200, 300)
	volume2 := volume4Ds([]scdussv1.LatLngPoint{
		{Lng: 0, Lat: 0},
		{Lng: 0, Lat: 1},
		{Lng: 1, Lat: 0},
	}, 100, 200)

	require.False(t, volume.VolumesIntersect(volume1, volume2))
}
