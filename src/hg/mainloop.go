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
	fixedPlayers     *PlayerGroup
	players          *PlayerGroup
	nextPlayerId     int
	activeDragSprite *Player
	dragController   *DragController
	frames           []frame
	activeFrame      *frame
}

type frame struct {
	Players *PlayerGroup
	// DurationSeconds is how long this frame plays for
	DurationSeconds float32
}

type playerSaveKey struct {
	symbol string
	team   int
}

func NewGame() *Game {
	g := &Game{
		fixedPlayers:   &PlayerGroup{},
		players:        &PlayerGroup{},
		dragController: &DragController{},
		frames:         []frame{{Players: &PlayerGroup{}}},
	}
	g.activeFrame = &g.frames[0]
	return g
}

func (g *Game) init() {
	g.initDone = true
	colors := []color.RGBA{
		{0x80, 0, 0, 0},
		{0, 0, 0xf0, 0}}
	symbols := strings.Split("LW,RW,C,F,F1,F2,F3,LD,RD,D,X", ",")
	for team, col := range colors {
		for i, symbol := range symbols {
			s, _ := MakeSymbol(symbol, 20, col)
			player := NewPlayerFromImage(s)
			player.Team = team
			player.Symbol = symbol
			player.X = i * (40 + 2)
			player.Y = 610 + team*42
			g.fixedPlayers.Add(player)
		}
	}
	g.Load()
}

type saveLoadSprite struct {
	Symbol string
	Team   int
	X, Y   int
	Id     int
}

type saveLoadData struct {
	Players      []saveLoadSprite
	NextPlayerId int
	Frames       []frame
}

func (g *Game) Save() {
	sld := saveLoadData{
		NextPlayerId: g.nextPlayerId,
		Frames:       g.frames,
	}
	for _, s := range g.players.player {
		sld.Players = append(sld.Players, saveLoadSprite{
			Symbol: s.Symbol,
			Team:   s.Team,
			Id:     s.Id,
			X:      s.X,
			Y:      s.Y,
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
	g.nextPlayerId = sld.NextPlayerId
	g.players = &PlayerGroup{}
	//g.frames = sld.Frames
	//g.activeFrame = &sld.Frames[0]

	extras := map[playerSaveKey]*Player{}
	for _, player := range g.fixedPlayers.player {
		extras[playerSaveKey{player.Symbol, player.Team}] = player
	}
	for _, toLoad := range sld.Players {
		s := extras[playerSaveKey{symbol: toLoad.Symbol, team: toLoad.Team}]
		if s != nil {
			toAdd := *s
			toAdd.Id = toLoad.Id
			toAdd.X = toLoad.X
			toAdd.Y = toLoad.Y
			g.players.Add(&toAdd)
		}
	}
}

func (g *Game) handleDragging() {
	if g.dragController.DragActive() {
		x, y := g.dragController.Position()
		if g.dragController.DragStart() {
			if placed := g.players.Under(x, y); placed != nil {
				g.activeDragSprite = placed
				g.players.Remove(placed)
				x, y = g.dragController.SetOffset(x-placed.X, y-placed.Y)
			} else if fixed := g.fixedPlayers.Under(x, y); fixed != nil {
				g.activeDragSprite = NewPlayerFromPlayer(fixed)
				g.activeDragSprite.Id = g.nextPlayerId
				g.nextPlayerId++
				x, y = g.dragController.SetOffset(x-fixed.X, y-fixed.Y)
			}
		}
		if g.activeDragSprite != nil {
			g.activeDragSprite.X = x
			g.activeDragSprite.Y = y
		}
	} else if g.dragController.Dropped() {
		if g.activeDragSprite != nil {
			_, y := g.dragController.Position()
			if y < 590 {
				g.players.Add(g.activeDragSprite)
			}
			g.activeDragSprite = nil
		}
	}
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

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)

	screen.DrawImage(rink, &ebiten.DrawImageOptions{})
	g.fixedPlayers.Draw(screen)
	g.players.Draw(screen)
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
		return nil
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}
