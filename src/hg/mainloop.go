package hg

import (
	"encoding/json"
	"fmt"
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
	debugui  debugui.DebugUI
	initDone bool

	fixedPlayers     *PlayerGroup
	nextPlayerId     int
	activeDragPlayer *Player
	dragController   *DragController
	frames           []frame
	activeFrameIndex int
	currentTime      float64

	activeSkatePath *SkatePath
}

type frame struct {
	Players *PlayerGroup
	// DurationSeconds is how long this frame plays for
	DurationSeconds float64
}

type playerSaveKey struct {
	symbol string
	team   int
}

func NewGame() *Game {
	g := &Game{
		fixedPlayers: &PlayerGroup{},
		//players:        &PlayerGroup{},
		dragController: &DragController{},
		frames: []frame{{
			Players:         &PlayerGroup{},
			DurationSeconds: 1}},
	}
	g.activeFrameIndex = 0
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
	NextPlayerId int
	Frames       []frame
}

func (g *Game) Save() {
	sld := saveLoadData{
		NextPlayerId: g.nextPlayerId,
		Frames:       g.frames,
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

	extras := map[playerSaveKey]*Player{}
	for _, player := range g.fixedPlayers.Players {
		extras[playerSaveKey{player.Symbol, player.Team}] = player
	}
	g.frames = sld.Frames
	g.activeFrameIndex = 0
	for _, frame := range g.frames {
		for _, toLoad := range frame.Players.Players {
			fixedPlayer := extras[playerSaveKey{symbol: toLoad.Symbol, team: toLoad.Team}]
			if fixedPlayer != nil {
				toLoad.CopyImagesFrom(fixedPlayer)
			}
		}
	}
}

func (g *Game) activeFrame() *frame {
	return &g.frames[g.activeFrameIndex]
}

func (g *Game) handleDragging() {
	if g.dragController.DragActive() {
		x, y := g.dragController.Position()
		if g.dragController.DragStart() {
			if placed := g.activeFrame().Players.Under(x, y); placed != nil {
				g.activeDragPlayer = placed
				g.activeFrame().Players.Remove(placed)
				x, y = g.dragController.SetOffset(x-placed.X, y-placed.Y)
				g.activeSkatePath = &SkatePath{TargetId: g.activeDragPlayer.Id}

			} else if fixed := g.fixedPlayers.Under(x, y); fixed != nil {
				g.activeDragPlayer = NewPlayerFromPlayer(fixed)
				g.activeDragPlayer.Id = g.nextPlayerId
				g.nextPlayerId++
				x, y = g.dragController.SetOffset(x-fixed.X, y-fixed.Y)
			}
		}
		if g.activeDragPlayer != nil {
			g.activeDragPlayer.X = x
			g.activeDragPlayer.Y = y
			pt := g.activeDragPlayer.CenterPoint()
			if g.activeSkatePath != nil {
				g.activeSkatePath.AddPt(pt)
			}
		}
	} else if g.dragController.Dropped() {
		if g.activeDragPlayer != nil {
			_, y := g.dragController.Position()
			if y < 590 {
				g.activeFrame().Players.Add(g.activeDragPlayer)
				if g.activeSkatePath != nil {
					g.activeDragPlayer.SkatePath = g.activeSkatePath
					g.activeDragPlayer.Interpolate(0)
				}
			}
			g.activeDragPlayer = nil
			g.activeSkatePath = nil
		}
	}
}

func (g *Game) Update() error {
	if !g.initDone {
		g.init()
	}
	g.debugui.Update(func(ctx *debugui.Context) error {
		ctx.Window("Test", image.Rect(526, 609, 875, 790), func(layout debugui.ContainerLayout) {
			ctx.Button("Save").On(func() { g.Save() })
			ctx.Button("Load").On(func() { g.Load() })
			ctx.Button("Next Frame").On(func() { g.NextFrame() })
			ctx.Button("Prev Frame").On(func() { g.PreviousFrame() })
			ctx.Button("New Frame").On(func() { g.NewFrame() })
			ctx.Text(fmt.Sprintf("Frame: %d (%d)", g.activeFrameIndex+1, len(g.frames)))
			ctx.NumberFieldF(&g.activeFrame().DurationSeconds, 0.01, 1)
			if g.activeFrame().DurationSeconds < 0 {
				g.activeFrame().DurationSeconds = 0
			}
			ctx.NumberFieldF(&g.currentTime, 0.01, 1)
		})
		return nil
	})
	g.handleDragging()
	g.activeFrame().Players.Interpolate(float32(g.currentTime))
	return nil
}

func (g *Game) NewFrame() {
	frame := *g.activeFrame()
	frame.Players = frame.Players.Clone()
	g.frames = append(g.frames, frame)
}

func (g *Game) PreviousFrame() {
	if g.activeFrameIndex > 0 {
		g.activeFrameIndex--
	}
}

func (g *Game) NextFrame() {
	if g.activeFrameIndex < len(g.frames)-1 {
		g.activeFrameIndex++
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.White)

	screen.DrawImage(rink, &ebiten.DrawImageOptions{})
	g.fixedPlayers.Draw(screen)
	g.activeFrame().Players.Draw(screen)
	if g.activeDragPlayer != nil {
		g.activeDragPlayer.DrawWithAlpha(screen, 0.8)
	}

	if g.activeSkatePath != nil {
		if g.activeDragPlayer != nil {
			g.activeSkatePath.DrawActive(screen, g.activeDragPlayer.CenterPoint())
		} else {
			g.activeSkatePath.Draw(screen)
		}
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
