package hg

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// SkatePoint represents a 2D point with float32 coordinates and a heading.
type SkatePoint struct {
	X, Y, HeadingRadians float32
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
		Width: 1,
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
	const minDist = 20
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

			h1 := p1.HeadingRadians
			h2 := p2.HeadingRadians

			// Find the shortest angle between h1 and h2 to handle wrapping at +/- PI.
			delta := h2 - h1
			if delta > math.Pi {
				delta -= 2 * math.Pi
			} else if delta < -math.Pi {
				delta += 2 * math.Pi
			}
			interpolatedPoint.HeadingRadians = h1 + segmentFraction*delta
			return interpolatedPoint
		}
		distCovered += segmentLength
	}

	// If fraction is 1.0 or slightly more due to float inaccuracies, return the last point.
	return sp.Points[len(sp.Points)-1]
}
