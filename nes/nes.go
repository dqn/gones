package nes

import (
	"io/ioutil"

	"github.com/dqn/gones/cpu"
	"github.com/dqn/gones/ppu"
	"github.com/dqn/gones/ram"
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
	var ram ram.RAM

	ppuBus := ppu.NewBus(characterROM)
	ppu := ppu.New(ppuBus)
	cpuBus := cpu.NewBus(&ram, programROM, ppu)
	cpu := cpu.New(cpuBus)

	nes := &NES{cpu, ppu}

	return nes, nil
}

func (n *NES) Run() error {
	for i := 0; i < 200; i++ {
		_, err := n.cpu.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
