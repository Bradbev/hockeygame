package hg

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// cribbed from https://ebitengine.org/en/examples/drag.html

type Sprite struct {
	image      *ebiten.Image
	alphaImage *image.Alpha
	x          int
	y          int
	extra      any
}

func NewSpriteSprite(sprite *Sprite) *Sprite {
	s := *sprite
	return &s
}

func NewSpriteFromImage(img *ebiten.Image) *Sprite {

	// Clone an image but only with alpha values.
	// This is used to detect a user cursor touches the image.
	b := img.Bounds()
	alphaImg := image.NewAlpha(b)
	for j := b.Min.Y; j < b.Max.Y; j++ {
		for i := b.Min.X; i < b.Max.X; i++ {
			alphaImg.Set(i, j, img.At(i, j))
		}
	}

	return &Sprite{
		image:      img,
		alphaImage: alphaImg,
	}
}

func (s *Sprite) Draw(screen *ebiten.Image) {
	s.DrawWithAlpha(screen, 1)
}

// DrawWithAlpha draws the sprite.
func (s *Sprite) DrawWithAlpha(screen *ebiten.Image, alpha float32) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.x), float64(s.y))
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(s.image, op)
}

// In returns true if (x, y) is in the sprite, and false otherwise.
func (s *Sprite) In(x, y int) bool {
	// Check the actual color (alpha) value at the specified position
	// so that the result of In becomes natural to users.
	//
	// Use alphaImage (*image.Alpha) instead of image (*ebiten.Image) here.
	// It is because (*ebiten.Image).At is very slow as this reads pixels from GPU,
	// and should be avoided whenever possible.
	ret := s.alphaImage.At(x-s.x, y-s.y).(color.Alpha).A > 0
	return ret
}
