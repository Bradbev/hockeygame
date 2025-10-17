package hg

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"slices"
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
	buttons          *ButtonGroup
	nextPlayerId     int
	activeDragPlayer *Player
	mouseController  *MouseController
	frames           []frame
	activeFrameIndex int
	currentTime      float64

	dragMovesPlayer bool

	activeSkatePath *SkatePath

	testSkatePath *SkatePathWithRadius
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
		fixedPlayers:    &PlayerGroup{},
		buttons:         &ButtonGroup{},
		mouseController: &MouseController{},
		frames: []frame{{
			Players:         &PlayerGroup{},
			DurationSeconds: 1}},
		testSkatePath: &SkatePathWithRadius{
			Points: []SkatePoint{
				{X: 200, Y: 100},
				{X: 400, Y: 200},
				{X: 300, Y: 300},
			},
			PointRadiuses:   []float32{0, 5, 0},
			editPointIndex:  -1,
			editRadiusIndex: -1,
		},
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
			s, _ := MakeCircle(symbol, 20, col)
			player := NewPlayerFromImage(s)
			player.Team = team
			player.Symbol = symbol
			player.X = i*(40+2) + 5
			player.Y = 610 + team*42
			g.fixedPlayers.Add(player)
		}
	}

	g.makeButtons()
	g.Load()
	g.dragMovesPlayer = true
}

func (g *Game) makeButtons() {
	w := float32(0)
	h := float32(30)
	x := 5
	y := 698

	button := func(s string, cb func()) {
		b, err := MakeButton(s, w, h, x, y, color.RGBA{0x80, 0x80, 0x80, 1}, cb)
		if err == nil {
			g.buttons.Add(b)
			y += int(h) + 5
		}
	}
	newCol := func(newWidth float32) {
		y = 698
		x += int(w) + 5
		w = newWidth
	}
	newCol(100)
	button("Save", g.Save)
	button("Load", g.Load)

	newCol(150)
	button("Move", func() { g.dragMovesPlayer = true })
	button("Prev Frame", g.PreviousFrame)
	button("New Frame", g.NewFrame)

	newCol(150)
	button("Skate", func() { g.dragMovesPlayer = false })
	button("Next Frame", g.NextFrame)
	button("Delete Frame", g.DeleteFrame)
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
	if g.mouseController.DragActive() {
		x, y := g.mouseController.Position()
		if g.mouseController.DragStart() {
			if player := g.activeFrame().Players.Under(x, y); player != nil {
				g.activeDragPlayer = player
				g.activeFrame().Players.Remove(player)
				x, y = g.mouseController.SetOffset(x-player.X, y-player.Y)
				if g.dragMovesPlayer {
					g.activeDragPlayer.SkatePath = nil
				} else {
					sp := &SkatePath{TargetId: g.activeDragPlayer.Id}
					if player.SkatePath != nil {
						// If the player already has a skate path, use it.
						sp = player.SkatePath
						sp.TruncatePathToFraction(float32(g.currentTime))
					}
					//g.activeSkatePath = &SkatePath{TargetId: g.activeDragPlayer.Id}
					g.activeSkatePath = sp
				}

			} else if fixed := g.fixedPlayers.Under(x, y); fixed != nil {
				g.activeDragPlayer = NewPlayerFromPlayer(fixed)
				g.activeDragPlayer.Id = g.nextPlayerId
				g.nextPlayerId++
				x, y = g.mouseController.SetOffset(x-fixed.X, y-fixed.Y)
			}
			g.buttons.OnDragStart(x, y)
		}
		g.buttons.OnDrag(x, y)
		if g.activeDragPlayer != nil {
			g.activeDragPlayer.X = x
			g.activeDragPlayer.Y = y
			pt := g.activeDragPlayer.CenterPoint()
			if g.activeSkatePath != nil {
				g.activeSkatePath.AddPt(pt)
			}
		}
	} else if g.mouseController.Dropped() {
		x, y := g.mouseController.Position()
		g.buttons.Dropped(x, y)
		if g.activeDragPlayer != nil {
			if y < 590 {
				g.activeFrame().Players.Add(g.activeDragPlayer)
				if g.activeSkatePath != nil {
					g.activeSkatePath.AddClosingPt(g.activeDragPlayer.CenterPoint())
					g.activeDragPlayer.SkatePath = g.activeSkatePath
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
			ctx.Text(fmt.Sprintf("Frame: %d (%d)", g.activeFrameIndex+1, len(g.frames)))
			ctx.NumberFieldF(&g.activeFrame().DurationSeconds, 0.01, 1)
			if g.activeFrame().DurationSeconds < 0 {
				g.activeFrame().DurationSeconds = 0
			}
			ctx.NumberFieldF(&g.currentTime, 0.01, 1)
		})
		return nil
	})
	g.mouseController.Update()
	if g.mouseController.IsDoubleClick() {
		fmt.Println("Double click")
	}
	g.handleDragging()
	g.activeFrame().Players.Interpolate(float32(g.currentTime))

	g.testSkatePath.UpdateForEdit(g.mouseController)
	return nil
}

func (g *Game) NewFrame() {
	frame := *g.activeFrame()
	frame.Players = frame.Players.CloneForNewFrame()
	g.frames = append(g.frames, frame)
	g.currentTime = 0
	g.NextFrame()
}

func (g *Game) DeleteFrame() {
	if len(g.frames) > 1 {
		g.frames = g.frames[:len(g.frames)-1]
		if g.activeFrameIndex >= len(g.frames) {
			g.activeFrameIndex = len(g.frames) - 1
		}
	}
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
	g.buttons.Draw(screen)

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

	g.DrawTest(screen)

	g.testSkatePath.DrawForEdit(screen)

	g.debugui.Draw(screen)
}

func (g *Game) DrawTest(screen *ebiten.Image) {
	spath := &SkatePathWithRadius{}
	players := make([]*Player, len(g.activeFrame().Players.Players))
	if g.activeDragPlayer != nil {
		players = append(players, g.activeDragPlayer)
	}
	copy(players, g.activeFrame().Players.Players)
	slices.SortFunc(players, func(a, b *Player) int { return a.Id - b.Id })
	for _, p := range players {
		spath.Points = append(spath.Points, SkatePoint{float32(p.CenterPoint().X), float32(p.CenterPoint().Y)})
		spath.PointRadiuses = append(spath.PointRadiuses, 80)
	}
	spath.Draw(screen)
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
