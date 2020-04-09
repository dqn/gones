package main

import (
	"log"

	"github.com/hajimehoshi/ebiten"
)

func update(screen *ebiten.Image) error {
	if ebiten.IsDrawingSkipped() {
		return nil
	}

	return nil
}

func main() {
	if err := ebiten.Run(update, 320, 240, 1, "gones"); err != nil {
		log.Fatal(err)
	}
}
