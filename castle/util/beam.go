package util

type BeamBoundingBox struct {
	UpperRight BeamPoint
	UpperLeft  BeamPoint
	LowerRight BeamPoint
	LowerLeft  BeamPoint
}

type BeamPoint struct {
	X, Y float64
}

func (bp BeamPoint) New(x, y float64) BeamPoint {
	return BeamPoint{
		X: x,
		Y: y,
	}
}

func (b BeamBoundingBox) Min() (x float64, y float64) {
	minx := b.UpperRight.X
	if b.UpperLeft.X < minx {
		minx = b.UpperLeft.X
	}
	if b.LowerRight.X < minx {
		minx = b.LowerRight.X
	}
	if b.LowerLeft.X < minx {
		minx = b.LowerLeft.X
	}

	miny := b.UpperRight.Y
	if b.UpperLeft.Y < miny {
		miny = b.UpperLeft.Y
	}
	if b.LowerRight.Y < miny {
		miny = b.LowerRight.Y
	}
	if b.LowerLeft.Y < miny {
		miny = b.LowerLeft.Y
	}

	return minx, miny
}

// https://golang.org/src/image/geom.go?s=4792:4837#L199
func (b BeamBoundingBox) Overlaps(s BeamBoundingBox) bool {

	bminx, bminy := b.Min()
	bmaxx, bmaxy := b.Max()

	sminx, sminy := s.Min()
	smaxx, smaxy := s.Max()

	return bminx < smaxx && sminx < bmaxx &&
		bminy < smaxy && sminy < bmaxy
}

func (b BeamBoundingBox) Max() (x float64, y float64) {
	maxx := b.UpperRight.X
	if b.UpperLeft.X > maxx {
		maxx = b.UpperLeft.X
	}
	if b.LowerRight.X > maxx {
		maxx = b.LowerRight.X
	}
	if b.LowerLeft.X > maxx {
		maxx = b.LowerLeft.X
	}

	maxy := b.UpperRight.Y
	if b.UpperLeft.Y > maxy {
		maxy = b.UpperLeft.Y
	}
	if b.LowerRight.Y > maxy {
		maxy = b.LowerRight.Y
	}
	if b.LowerLeft.Y > maxy {
		maxy = b.LowerLeft.Y
	}

	return maxx, maxy
}
