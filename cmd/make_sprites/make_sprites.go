package main

import (
	"fmt"
	"image/color"
	"image/png"
	"log"
	"os"

	"github.com/Bradbev/hockeygame/src/hg"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type Game struct {
	counter int
}

func (g *Game) Update() error {
	g.counter++
	if g.counter == 1 {
		return nil
	}
	return ebiten.Termination
}

func (g *Game) Draw(screen *ebiten.Image) {
	dst := screen

	dst.Fill(color.RGBA{0x0, 0x0, 0x0, 0xff})
	s, err := hg.MakeSymbol("LW", 20, color.RGBA{0xff, 0, 0, 0})
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 0)
	screen.DrawImage(s, op)

	f, err := os.Create("test_output/test1.png")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	defer f.Close()
	err = png.Encode(f, screen)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	g := &Game{counter: 0}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
