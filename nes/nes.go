package nes

import (
	"image/color"
	"io/ioutil"

	"github.com/dqn/gones/cpu"
	"github.com/dqn/gones/ppu"
	"github.com/dqn/gones/ram"
	"github.com/hajimehoshi/ebiten"
)

const (
	nesHeaderSize           = 0x0010 // 16 Byte
	programROMSizePerPage   = 0x4000 // 16 KiB
	characterROMSizePerPage = 0x2000 //  8 KiB
)

type NES struct {
	cpu *cpu.CPU
	ppu *ppu.PPU
}

func splitROM(buf []byte) ([]uint8, []uint8) {
	programROMPages, characterROMPages := int(buf[4]), int(buf[5])
	programROMEnd := nesHeaderSize + programROMSizePerPage*programROMPages
	characterROMEnd := programROMEnd + characterROMSizePerPage*characterROMPages

	return buf[nesHeaderSize:programROMEnd], buf[programROMEnd:characterROMEnd]
}

func New(path string) (*NES, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	programROM, characterROM := splitROM(buf)
	ppuBus := ppu.NewBus(characterROM)
	ppu := ppu.New(ppuBus)
	cpuBus := cpu.NewBus(&ram.RAM{}, programROM, ppu)
	cpu := cpu.New(cpuBus)

	nes := &NES{cpu, ppu}

	return nes, nil
}

func (n *NES) update(screen *ebiten.Image) error {
	if ebiten.IsDrawingSkipped() {
		return nil
	}

	for {
		cycle, err := n.cpu.Run()
		if err != nil {
			return err
		}

		b := n.ppu.Run(cycle * 3)
		if b == nil {
			continue
		}

		for i := 0; i < 240; i++ {
			for j := 0; j < 256; j++ {
				screen.Set(j, i, color.Color(b[i][j]))
			}
		}
		break
	}

	return nil
}

func (n *NES) Run() error {
	if err := ebiten.Run(n.update, 256, 240, 1, "gones"); err != nil {
		return err
	}
	return nil
}
