package src

import (
	"image"
	"image/color"
	_ "image/png"
	"os"
	"strings"

	"github.com/Bradbev/hockeygame/src/hg"
	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	ScreenW = 1300
	ScreenH = 800
)

var rink = mustLoadImage("assets/rink.png")

type Game struct {
	debugui          debugui.DebugUI
	initDone         bool
	lw               *hg.Sprite
	fixedSprites     *hg.SpriteGroup
	placedSprites    *hg.SpriteGroup
	activeDragSprite *hg.Sprite
}

func NewGame() *Game {
	g := &Game{
		fixedSprites: &hg.SpriteGroup{},
	}
	return g
}

func (g *Game) init() {
	g.initDone = true
	colors := []color.RGBA{
		{0x80, 0, 0, 0},
		{0, 0, 0xf0, 0}}
	items := strings.Split("LW,RW,C,F,F1,F2,F3,LD,RD,D,X", ",")
	for colIndex, col := range colors {
		for i, item := range items {
			s, _ := hg.MakeSymbol(item, 20, col)
			sprite := hg.NewSpriteFromImage(s)
			sprite.X = i * (40 + 2)
			sprite.Y = 610 + colIndex*42
			g.fixedSprites.Add(sprite)
		}
	}
}

func (g *Game) Update() error {
	if !g.initDone {
		g.init()
	}
	x, y := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if fixed := g.fixedSprites.Under(x, y); fixed != nil {
			g.activeDragSprite = hg.NewSpriteSprite(fixed)
		}
	}
	if g.activeDragSprite != nil {
		g.activeDragSprite.X = x
		g.activeDragSprite.Y = y
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}
	screen.DrawImage(rink, &op)
	op.GeoM.Translate(100, 40)
	g.fixedSprites.Draw(screen)
	if g.activeDragSprite != nil {
		g.activeDragSprite.DrawWithAlpha(screen, 0.8)
	}
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
