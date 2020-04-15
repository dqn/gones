package controller

import "github.com/hajimehoshi/ebiten"

type Controller struct {
	buttons uint8
}

func (c *Controller) Clear() {
	c.buttons = 0
}

func (c *Controller) ReadButton() uint8 {
	var k ebiten.Key
	switch c.buttons {
	case 0: // A
		k = ebiten.KeyZ
	case 1: // B
		k = ebiten.KeyC
	case 2: // Select
		k = ebiten.KeySpace
	case 3: // Start
		k = ebiten.KeyEnter
	case 4: // Up
		k = ebiten.KeyUp
	case 5: // Down
		k = ebiten.KeyDown
	case 6: // Left
		k = ebiten.KeyLeft
	case 7: // Right
		k = ebiten.KeyRight
	}
	c.buttons++
	if ebiten.IsKeyPressed(k) {
		return 1
	}
	return 0
}
