package hg

import "github.com/hajimehoshi/ebiten/v2"

type SpriteGroup struct {
	sprites []*Sprite
}

func (s *SpriteGroup) Add(sprite *Sprite) {
	s.sprites = append(s.sprites, sprite)
}

func (s *SpriteGroup) Draw(screen *ebiten.Image) {
	for _, spr := range s.sprites {
		spr.Draw(screen)
	}
}

func (s *SpriteGroup) Under(x, y int) *Sprite {
	for _, spr := range s.sprites {
		if spr.In(x, y) {
			return spr
		}
	}
	return nil
}
