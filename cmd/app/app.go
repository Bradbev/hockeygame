package main

import (
	"log"

	"github.com/Bradbev/hockeygame/src/hg"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(hg.ScreenW, hg.ScreenH)
	ebiten.SetWindowTitle("Hockey")
	if err := ebiten.RunGame(hg.NewGame()); err != nil {
		log.Fatal(err)
	}
}
