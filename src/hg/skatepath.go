package hg

import (
	"image"
	"math"
	"slices"

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
	// Todo - break out to SkatePathWithRadiusEditor struct
	editPointIndex  int
	editRadiusIndex int
}

func (sp *SkatePathWithRadius) DistancePathToPoint(p SkatePoint) float32 {
	dist := float32(100000)
	end := len(sp.Points) - 1
	for i := 0; i < end; i++ {
		d := pointToLineSegmentDistSquared(p, sp.Points[i], sp.Points[i+1])
		dist = min(dist, d)
	}
	return float32(math.Sqrt(float64(dist)))
}

func (sp *SkatePathWithRadius) UpdateForEdit(mouseController *MouseController) {
	const selectRadius = 10
	const selectRadius2 = selectRadius * selectRadius
	mp := SkatePoint{}
	if mouseController.DragStart() {
		mx, my := mouseController.Position()
		mp = SkatePoint{X: float32(mx), Y: float32(my)}
	}
	if mouseController.Dropped() {
		sp.editPointIndex = -1
		sp.editRadiusIndex = -1
	}
	if sp.editPointIndex > -1 {
		mx, my := mouseController.Position()
		mp = SkatePoint{X: float32(mx), Y: float32(my)}
		sp.Points[sp.editPointIndex] = mp
	}
	if sp.editRadiusIndex > -1 {
		mx, my := mouseController.Position()
		mp = SkatePoint{X: float32(mx), Y: float32(my)}

		i := sp.editRadiusIndex
		p := sp.Points[i]
		prev := sp.Points[i-1]
		next := sp.Points[i+1]
		radius := sp.PointRadiuses[i]
		radius = min(radius, p.Sub(prev).Length()/2)
		radius = min(radius, p.Sub(next).Length()/2)
		//_, centre := findPointOnRadiusCircle(next, p, prev, radius*10)
		sp.PointRadiuses[i] = p.Sub(mp).Length()

	}
	insertPointIndex := -1
	for i, p := range sp.Points {
		if p.Sub(mp).LengthSq() < selectRadius2 {
			sp.editPointIndex = i
		}
		if i < len(sp.Points)-1 {
			pt := p.Sub(sp.Points[i+1]).Mul(0.5).Add(sp.Points[i+1])
			if pt.Sub(mp).LengthSq() < selectRadius2 {
				insertPointIndex = i
			}
		}
		if i > 0 && i < len(sp.Points)-1 {
			prev := sp.Points[i-1]
			next := sp.Points[i+1]
			radius := sp.PointRadiuses[i]
			radius = min(radius, p.Sub(prev).Length()/2)
			radius = min(radius, p.Sub(next).Length()/2)
			_, centre := findPointOnRadiusCircle(next, p, prev, radius)
			if centre.Sub(mp).LengthSq() < selectRadius2 {
				sp.editRadiusIndex = i
				sp.editPointIndex = -1
			}
		}

	}
	if insertPointIndex >= 0 {
		i := insertPointIndex + 1
		sp.Points = slices.Insert(sp.Points, i, mp)
		sp.PointRadiuses = slices.Insert(sp.PointRadiuses, i, 5)
		sp.editPointIndex = insertPointIndex + 1
	}
}

func (sp *SkatePathWithRadius) DrawForEdit(screen *ebiten.Image) {
	const diamondRadius = 10
	sp.Draw(screen)
	for i, p := range sp.Points {
		drawDiamond(screen, p, diamondRadius)
		if i < len(sp.Points)-1 {
			pt := p.Sub(sp.Points[i+1]).Mul(0.5).Add(sp.Points[i+1])
			drawCross(screen, pt, diamondRadius-5)
		}

		if i > 0 && i < len(sp.Points)-1 {
			prev := sp.Points[i-1]
			next := sp.Points[i+1]
			radius := sp.PointRadiuses[i]
			radius = min(radius, p.Sub(prev).Length()/2)
			radius = min(radius, p.Sub(next).Length()/2)
			_, centre := findPointOnRadiusCircle(next, p, prev, radius)
			drawCross(screen, centre, diamondRadius)
		}
	}
}

func (sp *SkatePathWithRadius) Draw(screen *ebiten.Image) {
	path := vector.Path{}
	points := sp.pathPoints()
	path.MoveTo(float32(points[0].X), float32(points[0].Y))
	for _, p := range points[1:] {
		path.LineTo(float32(p.X), float32(p.Y))
	}

	dispatchPath(screen, &path, 3)
}

func (sp *SkatePathWithRadius) pathPoints() []SkatePoint {
	if len(sp.Points) < 2 {
		return nil
	}
	result := make([]SkatePoint, 0, 200)
	if len(sp.Points) == 2 {
		result = append(result, sp.Points[0])
		result = append(result, sp.Points[1])
		return result
	}
	for i, pt := range sp.Points {
		if i == 0 {
			// first point
			result = append(result, pt)
			continue
		}
		if i == len(sp.Points)-1 {
			// last point
			result = append(result, pt)
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

			//path.LineTo(entryPoint.X, entryPoint.Y)
			result = append(result, entryPoint)

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
					//path.LineTo(p.X, p.Y)
					result = append(result, p)
				}
			}
			sweep(toEntry, toMid)
			sweep(toMid, toExit)

			result = append(result, exitPoint)
			//path.LineTo(exitPoint.X, exitPoint.Y)
		}
	}
	return result
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

func pointToLineSegmentDist(p, v, w SkatePoint) float32 {
	return float32(math.Sqrt(float64(pointToLineSegmentDistSquared(p, v, w))))
}
func pointToLineSegmentDistSquared(p, v, w SkatePoint) float32 {
	d2 := func(a, b SkatePoint) float32 {
		return a.Sub(b).LengthSq()
	}
	l2 := d2(v, w)
	if l2 == 0 {
		return d2(p, v)
	}
	t := ((p.X-v.X)*(w.X-v.X) + (p.Y-v.Y)*(w.Y-v.Y)) / l2
	t = float32(math.Max(0, math.Min(1, float64(t))))
	return d2(p, SkatePoint{
		X: v.X + t*(w.X-v.X),
		Y: v.Y + t*(w.Y-v.Y),
	})
}

func drawDiamond(screen *ebiten.Image, p SkatePoint, radius float32) {
	path := vector.Path{}
	px, py := p.X, p.Y
	path.MoveTo(px, py-radius)
	path.LineTo(px+radius, py)
	path.LineTo(px, py+radius)
	path.LineTo(px-radius, py)
	path.Close()
	dispatchPath(screen, &path, 3)
}

func drawCross(screen *ebiten.Image, p SkatePoint, radius float32) {
	path := vector.Path{}
	px, py := p.X, p.Y
	path.MoveTo(px+radius, py+radius)
	path.LineTo(px-radius, py-radius)
	path.MoveTo(px-radius, py+radius)
	path.LineTo(px+radius, py-radius)
	dispatchPath(screen, &path, 3)
}
