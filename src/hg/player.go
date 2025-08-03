package hg

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// cribbed from https://ebitengine.org/en/examples/drag.html

type Player struct {
	image      *ebiten.Image
	alphaImage *image.Alpha
	X          int
	Y          int
	Id         int
	Team       int
	Symbol     string
}

func NewPlayerFromPlayer(player *Player) *Player {
	s := *player
	return &s
}

func NewPlayerFromImage(img *ebiten.Image) *Player {

	// Clone an image but only with alpha values.
	// This is used to detect a user cursor touches the image.
	b := img.Bounds()
	alphaImg := image.NewAlpha(b)
	for j := b.Min.Y; j < b.Max.Y; j++ {
		for i := b.Min.X; i < b.Max.X; i++ {
			alphaImg.Set(i, j, img.At(i, j))
		}
	}

	return &Player{
		image:      img,
		alphaImage: alphaImg,
	}
}

func (p *Player) CopyImagesFrom(player *Player) {
	p.alphaImage = player.alphaImage
	p.image = player.image
}

func (s *Player) Draw(screen *ebiten.Image) {
	s.DrawWithAlpha(screen, 1)
}

// DrawWithAlpha draws the sprite.
func (s *Player) DrawWithAlpha(screen *ebiten.Image, alpha float32) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.X), float64(s.Y))
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(s.image, op)
}

// In returns true if (x, y) is in the sprite, and false otherwise.
func (s *Player) In(x, y int) bool {
	// Check the actual color (alpha) value at the specified position
	// so that the result of In becomes natural to users.
	//
	// Use alphaImage (*image.Alpha) instead of image (*ebiten.Image) here.
	// It is because (*ebiten.Image).At is very slow as this reads pixels from GPU,
	// and should be avoided whenever possible.
	ret := s.alphaImage.At(x-s.X, y-s.Y).(color.Alpha).A > 0
	return ret
}
