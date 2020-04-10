package main

import (
	"log"

	"github.com/dqn/gones/nes"
	"github.com/hajimehoshi/ebiten"
)

func update(screen *ebiten.Image) error {
	if ebiten.IsDrawingSkipped() {
		return nil
	}

	return nil
}

func run() error {
	n, err := nes.New("./sample1/sample1.nes")
	if err != nil {
		return err
	}
	return n.Run()
	// if err := ebiten.Run(update, 320, 240, 1, "gones"); err != nil {
	// 	log.Fatal(err)
	// }
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
