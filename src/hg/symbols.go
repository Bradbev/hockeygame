package hg

import (
	"bytes"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	whiteImage    = ebiten.NewImage(3, 3)
	whiteSubImage = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
)

func init() {
	whiteImage.Fill(color.White)
}

func MakeCircle(letters string, r float32, clr color.Color) (*ebiten.Image, error) {
	img := ebiten.NewImage(int(r*2), int(r*2))
	img.Fill(color.Transparent)
	var path vector.Path

	path.Arc(r, r, r, 0, 360, vector.Clockwise)
	path.Close()

	cr, cg, cb, _ := clr.RGBA()
	vertices, indices := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vertices {
		vertices[i].SrcX = 1
		vertices[i].SrcY = 1
		vertices[i].ColorR = float32(cr) / float32(0xffff)
		vertices[i].ColorG = float32(cg) / float32(0xffff)
		vertices[i].ColorB = float32(cb) / float32(0xffff)
		vertices[i].ColorA = 1
	}

	op := &ebiten.DrawTrianglesOptions{
		AntiAlias: true,
		FillRule:  ebiten.FillRuleNonZero,
	}
	img.DrawTriangles(vertices, indices, whiteSubImage, op)

	// Draw the text
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		return img, err
	}
	face := &text.GoTextFace{
		Source: s,
		Size:   float64(r),
	}
	w, h := text.Measure(letters, face, 0)
	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(float64(r)-w/2, float64(r)-h/2)
	textOp.ColorScale.ScaleWithColor(color.White)
	text.Draw(img, letters, face, textOp)

	return img, nil
}

type ButtonGroup struct {
	Buttons      []*Button
	activeButton *Button
}

// Add adds a button to the group.
func (bg *ButtonGroup) Add(b *Button) {
	bg.Buttons = append(bg.Buttons, b)
}

// Draw draws all buttons in the group.
func (bg *ButtonGroup) Draw(screen *ebiten.Image) {
	for _, b := range bg.Buttons {
		b.Draw(screen)
	}
}

func (bg *ButtonGroup) OnDragStart(x, y int) {
	if b := bg.Under(x, y); b != nil {
		b.OnDrag(x, y)
	}
}

func (bg *ButtonGroup) OnDrag(x, y int) {
	if bg.activeButton != nil {
		bg.activeButton.OnDrag(x, y)
	}
}

func (bg *ButtonGroup) Dropped(x, y int) {
	if bg.activeButton != nil {
		if bg.activeButton.OnDragRelease(x, y) {
			bg.activeButton.Callback()
		}
		bg.activeButton = nil
	}
}

// Under returns the button under the given coordinates, or nil if none.
func (bg *ButtonGroup) Under(x, y int) *Button {
	for _, b := range bg.Buttons {
		if b.In(x, y) {
			bg.activeButton = b
			return b
		}
	}
	return nil
}

type Button struct {
	X          int
	Y          int
	isDown     bool
	image      *ebiten.Image
	alphaImage *image.Alpha
	callback   func()
}

func MakeButton(str string, w float32, h float32, x, y int, clr color.Color, callback func()) (*Button, error) {
	img := ebiten.NewImage(int(w), int(h))
	img.Fill(color.Black)

	padding := float32(1.0)
	vector.DrawFilledRect(img, padding, padding, w-2*padding, h-2*padding, clr, true)

	// Draw the text
	s, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.MPlus1pRegular_ttf))
	if err != nil {
		return nil, err
	}
	face := &text.GoTextFace{
		Source: s,
		Size:   float64(h * 0.75),
	}
	tw, th := text.Measure(str, face, 0)
	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(float64(w)/2-tw/2, float64(h)/2-th/2)
	textOp.ColorScale.ScaleWithColor(color.White)
	text.Draw(img, str, face, textOp)

	b := img.Bounds()
	alphaImg := image.NewAlpha(b)
	for j := b.Min.Y; j < b.Max.Y; j++ {
		for i := b.Min.X; i < b.Max.X; i++ {
			alphaImg.Set(i, j, img.At(i, j))
		}
	}

	return &Button{image: img, alphaImage: alphaImg, X: x, Y: y, callback: callback}, nil
}

func (b *Button) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(b.X), float64(b.Y))
	alpha := float32(1)
	if b.isDown {
		alpha = 0.5
	}
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(b.image, op)
}

// In returns true if (x, y) is in the sprite, and false otherwise.
func (b *Button) In(x, y int) bool {
	return x >= b.X && x < b.X+b.image.Bounds().Dx() &&
		y >= b.Y && y < b.Y+b.image.Bounds().Dy()
}

func (b *Button) OnDrag(x, y int) {
	b.isDown = b.In(x, y)
}

func (b *Button) OnDragRelease(x, y int) bool {
	b.isDown = false
	return b.In(x, y)
}

func (b *Button) Callback() {
	if b.callback != nil {
		b.callback()
	}
}
