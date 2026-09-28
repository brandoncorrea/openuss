package volume

import (
	"slices"

	"bwawan.com/openuss/sdk/api/scdussv1"
	"github.com/peterstace/simplefeatures/geom"
)

func VolumesIntersect(ours, theirs []scdussv1.Volume4D) bool {
	// TODO(gap): This assumes volumes always have exactly one element
	return volume4DIntersects(ours[0], theirs[0])
}

// TODO(gap): This currently does no time-based validations
func volume4DIntersects(ours, theirs scdussv1.Volume4D) bool {
	return altitudesIntersect(ours, theirs) &&
		volume2DIntersects(ours, theirs)
}

// TODO(gap): The only check for altitudes is that 'our lower' is less than than 'their upper'
func altitudesIntersect(ours, theirs scdussv1.Volume4D) bool {
	return ours.Volume.AltitudeLower.Value < theirs.Volume.AltitudeUpper.Value
}

func volume2DIntersects(ours, theirs scdussv1.Volume4D) bool {
	ourOutline := Geometry(ours.Volume)
	theirOutline := Geometry(theirs.Volume)
	// TODO(gap): Volumes that share a zero-area space intersect
	return geom.Intersects(ourOutline, theirOutline)
}

// TODO(gap): A circular outline is unhandled
func Geometry(volume scdussv1.Volume3D) geom.Geometry {
	// TODO(gap): Nothing validates a polygon whose edges cross or whose vertices repeat
	vertices := volume.OutlinePolygon.Vertices
	ring := slices.Concat(vertices, vertices[:1])
	coordinates := make([]float64, 0, len(ring)*geom.DimXY.Dimension())
	for _, vertex := range ring {
		coordinates = append(coordinates, float64(vertex.Lng), float64(vertex.Lat))
	}
	exterior := geom.NewLineString(geom.NewSequence(coordinates, geom.DimXY))
	return geom.NewPolygon([]geom.LineString{exterior}).AsGeometry()
}
