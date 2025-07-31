package hg

import (
	"encoding/json"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"strings"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenW = 1300
	ScreenH = 800
)

var rink = mustLoadImage("assets/rink.png")

type Game struct {
	debugui          debugui.DebugUI
	initDone         bool
	fixedSprites     *SpriteGroup
	placedSprites    *SpriteGroup
	activeDragSprite *Sprite
	dragController   *DragController
}

type spriteExtra struct {
	symbol string
	team   int
}

func NewGame() *Game {
	g := &Game{
		fixedSprites:   &SpriteGroup{},
		placedSprites:  &SpriteGroup{},
		dragController: &DragController{},
	}
	return g
}

func (g *Game) init() {
	g.initDone = true
	colors := []color.RGBA{
		{0x80, 0, 0, 0},
		{0, 0, 0xf0, 0}}
	items := strings.Split("LW,RW,C,F,F1,F2,F3,LD,RD,D,X", ",")
	for team, col := range colors {
		for i, item := range items {
			s, _ := MakeSymbol(item, 20, col)
			sprite := NewSpriteFromImage(s)
			sprite.extra = &spriteExtra{item, team}
			sprite.x = i * (40 + 2)
			sprite.y = 610 + team*42
			g.fixedSprites.Add(sprite)
		}
	}
	g.Load()
}

func (g *Game) Update() error {
	if !g.initDone {
		g.init()
	}
	g.debugui.Update(func(ctx *debugui.Context) error {
		ctx.Window("Test", image.Rect(526, 609, 875, 700), func(layout debugui.ContainerLayout) {
			ctx.Button("Save").On(func() { g.Save() })
			ctx.Button("Load").On(func() { g.Load() })
		})
		return nil
	})
	g.handleDragging()
	return nil
}

type saveLoadSprite struct {
	Symbol string
	Team   int
	X, Y   int
}

type saveLoadData struct {
	Sprites []saveLoadSprite
}

func (g *Game) Save() {
	sld := saveLoadData{}
	for _, s := range g.placedSprites.sprites {
		extra := s.extra.(*spriteExtra)
		sld.Sprites = append(sld.Sprites, saveLoadSprite{
			Symbol: extra.symbol,
			Team:   extra.team,
			X:      s.x,
			Y:      s.y,
		})
	}
	data, _ := json.MarshalIndent(sld, "", " ")
	os.WriteFile("saved.json", data, os.ModePerm)
}

func (g *Game) Load() {
	data, err := os.ReadFile("saved.json")
	if err != nil {
		return
	}
	sld := saveLoadData{}
	err = json.Unmarshal(data, &sld)
	if err != nil {
		return
	}
	extras := map[spriteExtra]*Sprite{}
	for _, sprite := range g.fixedSprites.sprites {
		extras[*(sprite.extra.(*spriteExtra))] = sprite
	}
	for _, toLoad := range sld.Sprites {
		s := extras[spriteExtra{symbol: toLoad.Symbol, team: toLoad.Team}]
		if s != nil {
			toAdd := *s
			toAdd.x = toLoad.X
			toAdd.y = toLoad.Y
			g.placedSprites.Add(&toAdd)
		}
	}
}

func (g *Game) handleDragging() {
	if g.dragController.DragActive() {
		x, y := g.dragController.Position()
		if g.dragController.DragStart() {
			if placed := g.placedSprites.Under(x, y); placed != nil {
				g.activeDragSprite = placed
				g.placedSprites.Remove(placed)
				x, y = g.dragController.SetOffset(x-placed.x, y-placed.y)
			} else if fixed := g.fixedSprites.Under(x, y); fixed != nil {
				g.activeDragSprite = NewSpriteSprite(fixed)
				x, y = g.dragController.SetOffset(x-fixed.x, y-fixed.y)
			}
		}
		if g.activeDragSprite != nil {
			g.activeDragSprite.x = x
			g.activeDragSprite.y = y
		}
	} else if g.dragController.Dropped() {
		if g.activeDragSprite != nil {
			_, y := g.dragController.Position()
			if y < 590 {
				g.placedSprites.Add(g.activeDragSprite)
			}
			g.activeDragSprite = nil
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)

	screen.DrawImage(rink, &ebiten.DrawImageOptions{})
	g.fixedSprites.Draw(screen)
	g.placedSprites.Draw(screen)
	if g.activeDragSprite != nil {
		g.activeDragSprite.DrawWithAlpha(screen, 0.8)
	}
	g.debugui.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenW, ScreenH
}

func mustLoadImage(name string) *ebiten.Image {
	f, err := os.Open(name)
	if err != nil {
		//panic(err)
		return nil
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}
