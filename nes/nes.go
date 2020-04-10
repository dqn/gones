package nes

import (
	"io/ioutil"

	"github.com/dqn/gones/cpu"
	"github.com/dqn/gones/ppu"
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

func New(path string) (*NES, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	programROMPages, characterROMPages := int(buf[4]), int(buf[5])
	programROMEnd := nesHeaderSize + programROMSizePerPage*programROMPages
	characterROMEnd := programROMEnd + characterROMSizePerPage*characterROMPages

	nes := &NES{
		cpu: cpu.New(buf[nesHeaderSize:programROMEnd]),
		ppu: ppu.New(buf[programROMEnd:characterROMEnd]),
	}

	return nes, nil
}

func (n *NES) Run() error {
	for i := 0; i < 65535; i++ {
		_, err := n.cpu.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
