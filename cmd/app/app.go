package main

import (
	"log"

	"github.com/Bradbev/hockeygame/src"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(src.ScreenW, src.ScreenH)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(src.NewGame()); err != nil {
		log.Fatal(err)
	}
}
