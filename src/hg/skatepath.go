package hg

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// SkatePoint represents a 2D point with float32 coordinates
type SkatePoint struct {
	X, Y float32
}

// Heading in radians
func (p SkatePoint) Heading() float32 {
	d1 := p.Normalize()
	angle := math.Atan2(float64(d1.Y), float64(d1.X))
	if angle < 0 {
		angle += 2 * math.Pi
	}
	return float32(angle)
}

func (p SkatePoint) Normalize() SkatePoint {
	len := p.Length()
	if len == 0 {
		return SkatePoint{0, 0}
	}
	return SkatePoint{X: p.X / len, Y: p.Y / len}
}

// Sub returns the vector p-q.
func (p SkatePoint) Sub(q SkatePoint) SkatePoint {
	return SkatePoint{X: p.X - q.X, Y: p.Y - q.Y}
}

// Add returns the vector p+q.
func (p SkatePoint) Add(q SkatePoint) SkatePoint {
	return SkatePoint{X: p.X + q.X, Y: p.Y + q.Y}
}

// Mul returns the vector p*s.
func (p SkatePoint) Mul(s float32) SkatePoint {
	return SkatePoint{X: p.X * s, Y: p.Y * s}
}

// LengthSq returns p's squared length from the origin.
func (p SkatePoint) LengthSq() float32 {
	return p.X*p.X + p.Y*p.Y
}

// Length returns p's length from the origin.
func (p SkatePoint) Length() float32 {
	return float32(math.Sqrt(float64(p.LengthSq())))
}

type SkatePath struct {
	TargetId int
	Points   []SkatePoint
}

func (sp *SkatePath) Draw(screen *ebiten.Image) {
	sp.drawActive(screen, nil)
}

func (sp *SkatePath) DrawActive(screen *ebiten.Image, lastPoint image.Point) {
	sp.drawActive(screen, &lastPoint)
}

func (sp *SkatePath) drawActive(screen *ebiten.Image, lastPoint *image.Point) {
	path := vector.Path{}
	pt := sp.Points[0]
	path.MoveTo(float32(pt.X), float32(pt.Y))
	for _, pt := range sp.Points[1:] {
		path.LineTo(float32(pt.X), float32(pt.Y))
		//vector.DrawFilledCircle(screen, float32(pt.X), float32(pt.Y), 5, color.Black, true)
	}

	// connect from the player to the last point
	if lastPoint != nil {
		pt := *lastPoint
		path.LineTo(float32(pt.X), float32(pt.Y))
	}

	vertices, indices := path.AppendVerticesAndIndicesForStroke(nil, nil, &vector.StrokeOptions{
		Width: 3,
	})
	for i := range vertices {
		vertices[i].SrcX = 1
		vertices[i].SrcY = 1
		vertices[i].ColorR = 0
		vertices[i].ColorG = 0
		vertices[i].ColorB = 0
		vertices[i].ColorA = 0.6
	}
	op := &ebiten.DrawTrianglesOptions{
		AntiAlias: true,
		FillRule:  ebiten.FillRuleNonZero,
	}
	screen.DrawTriangles(vertices, indices, whiteSubImage, op)
}

// farEnoughToAddPoint checks if a point is far enough from the last point in the path.
func (sp *SkatePath) farEnoughToAddPoint(p SkatePoint) bool {
	const minDist = 10
	if len(sp.Points) == 0 {
		return true
	}
	lastPt := sp.Points[len(sp.Points)-1]
	return p.Sub(lastPt).LengthSq() >= minDist*minDist
}

// AddSkatePoint adds a point to the path.
// To avoid having too many points, it only adds the point if it is
// a certain distance away from the previous point in the path.
func (sp *SkatePath) AddSkatePoint(p SkatePoint) {
	if !sp.farEnoughToAddPoint(p) {
		return
	}
	sp.Points = append(sp.Points, p)
}

// AddPt is a helper to add a point from an image.Point.
func (sp *SkatePath) AddPt(p image.Point) {
	sp.AddSkatePoint(SkatePoint{X: float32(p.X), Y: float32(p.Y)})
}

func (sp *SkatePath) AddClosingPt(p image.Point) {
	pt := SkatePoint{
		X: float32(p.X),
		Y: float32(p.Y),
	}
	sp.Points = append(sp.Points, pt)
}

// TotalLength calculates the total length of the path.
func (sp *SkatePath) TotalLength() float32 {
	var total float32
	if len(sp.Points) < 2 {
		return 0
	}
	for i := 0; i < len(sp.Points)-1; i++ {
		p1 := sp.Points[i]
		p2 := sp.Points[i+1]
		total += p2.Sub(p1).Length()
	}
	return total
}

// Interpolate returns a point along the path at a given fraction of its total length.
// The fraction should be between 0.0 and 1.0.
func (sp *SkatePath) Interpolate(fraction float32) SkatePoint {
	if len(sp.Points) == 0 {
		return SkatePoint{}
	}
	if len(sp.Points) == 1 {
		return sp.Points[0]
	}

	targetDist := sp.TotalLength() * fraction
	if targetDist <= 0 {
		return sp.Points[0]
	}

	var distCovered float32
	for i := 0; i < len(sp.Points)-1; i++ {
		p1 := sp.Points[i]
		p2 := sp.Points[i+1]
		segmentVector := p2.Sub(p1)
		segmentLength := segmentVector.Length()

		if distCovered+segmentLength >= targetDist {
			distIntoSegment := targetDist - distCovered
			if segmentLength == 0 {
				return p1
			}
			segmentFraction := distIntoSegment / segmentLength
			// The position is interpolated.
			interpolatedPoint := p1.Add(segmentVector.Mul(segmentFraction))

			return interpolatedPoint
		}
		distCovered += segmentLength
	}

	// If fraction is 1.0 or slightly more due to float inaccuracies, return the last point.
	return sp.Points[len(sp.Points)-1]
}

func (sp *SkatePath) TruncatePathToFraction(fraction float32) {
	if len(sp.Points) <= 1 {
		sp.Points = nil
		return
	}

	targetDist := sp.TotalLength() * fraction
	if targetDist <= 0 {
		sp.Points = nil
		return
	}

	var distCovered float32
	for i := 0; i < len(sp.Points)-1; i++ {
		p1 := sp.Points[i]
		p2 := sp.Points[i+1]
		segmentVector := p2.Sub(p1)
		segmentLength := segmentVector.Length()

		if distCovered+segmentLength >= targetDist {
			// We found the segment where the target distance falls.
			sp.Points = sp.Points[:i+1] // Keep points up to this segment
			return
		}
		distCovered += segmentLength
	}
}

type SkatePathWithRadius struct {
	TargetId      int
	Points        []SkatePoint
	PointRadiuses []float32
}

func (sp *SkatePathWithRadius) Draw(screen *ebiten.Image) {
	if len(sp.Points) < 3 {
		return
	}
	path := vector.Path{}
	for i, pt := range sp.Points {
		if i == 0 {
			// first point
			path.MoveTo(float32(pt.X), float32(pt.Y))
			continue
		}
		if i == len(sp.Points)-1 {
			// last point
			path.LineTo(float32(pt.X), float32(pt.Y))
			continue
		}
		prev := sp.Points[i-1]
		next := sp.Points[i+1]

		// Interior points.  The tangents are not exactly correct because we are finding between
		// the points to the current point circle - not circle to circle.  But it's close enough
		// to look OK.
		if sp.PointRadiuses[i] > 0 {
			radius := sp.PointRadiuses[i]
			radius = min(radius, pt.Sub(prev).Length()/2)
			radius = min(radius, pt.Sub(next).Length()/2)
			entryPoint, _ := findPointOnRadiusCircle(prev, pt, next, radius)
			exitPoint, centre := findPointOnRadiusCircle(next, pt, prev, radius)

			path.LineTo(entryPoint.X, entryPoint.Y)

			toEntry := entryPoint.Sub(centre).Heading()
			toMid := pt.Sub(centre).Heading()
			toExit := exitPoint.Sub(centre).Heading()

			sweep := func(a1, a2 float32) {
				points := 10
				inc := AngleShortestSweep(a1, a2) / float32(points)

				for i := 0; i < points; i++ {
					angle := a1 + inc*float32(i)
					p := SkatePoint{X: radius * float32(math.Cos(float64(angle))), Y: radius * float32(math.Sin(float64(angle)))}
					p = p.Add(centre)
					path.LineTo(p.X, p.Y)
				}
			}
			sweep(toEntry, toMid)
			sweep(toMid, toExit)

			path.LineTo(exitPoint.X, exitPoint.Y)
		}
	}
	dispatchPath(screen, &path, 3)
}

func findPointOnRadiusCircle(prev, current, next SkatePoint, radius float32) (radiusPoint, centre SkatePoint) {
	// find the circle centre point.  This point is halfway (by angle) between the prev and next points
	dirToPrev := prev.Sub(current).Normalize()
	dirToNext := next.Sub(current).Normalize()
	circleCentre := dirToNext.Sub(dirToPrev).Mul(0.5).Add(dirToPrev).Normalize().Mul(radius).Add(current)

	// form a triangle from the previous point (moved to 0,0), the circle centre and the tangent point to be found
	hyp := circleCentre.Sub(prev)
	opp := radius
	a := math.Asin(float64(opp / hyp.Length()))
	b := math.Atan2(float64(hyp.Y), float64(hyp.X))
	t1 := b - a
	p1 := SkatePoint{X: opp * float32(math.Sin(t1)), Y: opp * float32(-math.Cos(t1))}
	t2 := b + a
	p2 := SkatePoint{X: opp * float32(-math.Sin(t2)), Y: opp * float32(math.Cos(t2))}
	currentMoved := current.Sub(circleCentre)
	if currentMoved.Sub(p1).LengthSq() < currentMoved.Sub(p2).LengthSq() {
		return p1.Add(circleCentre), circleCentre
	}
	return p2.Add(circleCentre), circleCentre
}

func dispatchPath(screen *ebiten.Image, path *vector.Path, w float32) {
	vertices, indices := path.AppendVerticesAndIndicesForStroke(nil, nil, &vector.StrokeOptions{
		Width: w,
	})
	for i := range vertices {
		vertices[i].SrcX = 1
		vertices[i].SrcY = 1
		vertices[i].ColorR = 0
		vertices[i].ColorG = 0
		vertices[i].ColorB = 0
		vertices[i].ColorA = 0.6
	}
	op := &ebiten.DrawTrianglesOptions{
		AntiAlias: true,
		FillRule:  ebiten.FillRuleNonZero,
	}
	screen.DrawTriangles(vertices, indices, whiteSubImage, op)
}
