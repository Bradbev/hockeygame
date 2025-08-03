package hg

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

type PlayerGroup struct {
	player []*Player
}

func (s *PlayerGroup) Add(player *Player) {
	s.player = append(s.player, player)
}

func (s *PlayerGroup) Remove(player *Player) {
	s.player = slices.DeleteFunc(s.player, func(sp *Player) bool {
		return player == sp
	})
}

func (s *PlayerGroup) Draw(screen *ebiten.Image) {
	for _, spr := range s.player {
		spr.Draw(screen)
	}
}

func (s *PlayerGroup) Under(x, y int) *Player {
	for _, player := range s.player {
		if player.In(x, y) {
			return player
		}
	}
	return nil
}
