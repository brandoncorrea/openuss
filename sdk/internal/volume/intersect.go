package volume

import (
	"slices"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"github.com/peterstace/simplefeatures/geom"
)

// TODO(gap): This assumes every pair of volumes overlaps in time and altitude
// TODO(gap): This assumes volumes always have one element
// TODO(gap): Volumes that share zero are are considered intersecting
func VolumesIntersect(ours, theirs []scdussv1.Volume4D) bool {
	return volume4DIntersects(ours[0], theirs[0])
}

func volume4DIntersects(ours, theirs scdussv1.Volume4D) bool {
	ourOutline := Geometry(ours.Volume)
	theirOutline := Geometry(theirs.Volume)

	return altitudesIntersect(ours, theirs) && geom.Intersects(ourOutline, theirOutline)
}

func altitudesIntersect(ours, theirs scdussv1.Volume4D) bool {
	return (*ours.Volume.AltitudeLower).Value < (*theirs.Volume.AltitudeUpper).Value
}

// TODO(gap): A circular outline is unhandled
func Geometry(volume scdussv1.Volume3D) geom.Geometry {
	// TODO(gap): Nothing validates a polygon whose edges cross or whose vertices repeat
	vertices := volume.OutlinePolygon.Vertices
	ring := slices.Concat(vertices, vertices[:1])
	coordinates := make([]float64, len(ring)*geom.DimXY.Dimension())
	for i, vertex := range ring {
		coordinates[i*2] = float64(vertex.Lng)
		coordinates[(i*2)+1] = float64(vertex.Lat)
	}
	exterior := geom.NewLineString(geom.NewSequence(coordinates, geom.DimXY))
	return geom.NewPolygon([]geom.LineString{exterior}).AsGeometry()
}
