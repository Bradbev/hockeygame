package hg

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

type PlayerGroup struct {
	Players []*Player
}

func (p *PlayerGroup) Clone() *PlayerGroup {
	ret := &PlayerGroup{}
	for _, p := range p.Players {
		player := *p
		ret.Players = append(ret.Players, &player)
	}
	return ret
}

func (p *PlayerGroup) Add(player *Player) {
	p.Players = append(p.Players, player)
}

func (p *PlayerGroup) Remove(player *Player) {
	p.Players = slices.DeleteFunc(p.Players, func(sp *Player) bool {
		return player == sp
	})
}

func (p *PlayerGroup) Interpolate(fraction float32) {
	for _, player := range p.Players {
		player.Interpolate(fraction)
	}
}

func (p *PlayerGroup) Draw(screen *ebiten.Image) {
	for _, player := range p.Players {
		player.Draw(screen)
	}
}

func (p *PlayerGroup) Under(x, y int) *Player {
	for _, player := range p.Players {
		if player.In(x, y) {
			return player
		}
	}
	return nil
}
