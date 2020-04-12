package ppu

import (
	"fmt"
	"image/color"
)

const (
	width        = 256
	height       = 240
	cyclePerLine = 341
	vBlank       = 20
)

var colors = [...]color.RGBA{
	{0x80, 0x80, 0x80, 0xFF}, {0x00, 0x3D, 0xA6, 0xFF}, {0x00, 0x12, 0xB0, 0xFF}, {0x44, 0x00, 0x96, 0xFF},
	{0xA1, 0x00, 0x5E, 0xFF}, {0xC7, 0x00, 0x28, 0xFF}, {0xBA, 0x06, 0x00, 0xFF}, {0x8C, 0x17, 0x00, 0xFF},
	{0x5C, 0x2F, 0x00, 0xFF}, {0x10, 0x45, 0x00, 0xFF}, {0x05, 0x4A, 0x00, 0xFF}, {0x00, 0x47, 0x2E, 0xFF},
	{0x00, 0x41, 0x66, 0xFF}, {0x00, 0x00, 0x00, 0xFF}, {0x05, 0x05, 0x05, 0xFF}, {0x05, 0x05, 0x05, 0xFF},
	{0xC7, 0xC7, 0xC7, 0xFF}, {0x00, 0x77, 0xFF, 0xFF}, {0x21, 0x55, 0xFF, 0xFF}, {0x82, 0x37, 0xFA, 0xFF},
	{0xEB, 0x2F, 0xB5, 0xFF}, {0xFF, 0x29, 0x50, 0xFF}, {0xFF, 0x22, 0x00, 0xFF}, {0xD6, 0x32, 0x00, 0xFF},
	{0xC4, 0x62, 0x00, 0xFF}, {0x35, 0x80, 0x00, 0xFF}, {0x05, 0x8F, 0x00, 0xFF}, {0x00, 0x8A, 0x55, 0xFF},
	{0x00, 0x99, 0xCC, 0xFF}, {0x21, 0x21, 0x21, 0xFF}, {0x09, 0x09, 0x09, 0xFF}, {0x09, 0x09, 0x09, 0xFF},
	{0xFF, 0xFF, 0xFF, 0xFF}, {0x0F, 0xD7, 0xFF, 0xFF}, {0x69, 0xA2, 0xFF, 0xFF}, {0xD4, 0x80, 0xFF, 0xFF},
	{0xFF, 0x45, 0xF3, 0xFF}, {0xFF, 0x61, 0x8B, 0xFF}, {0xFF, 0x88, 0x33, 0xFF}, {0xFF, 0x9C, 0x12, 0xFF},
	{0xFA, 0xBC, 0x20, 0xFF}, {0x9F, 0xE3, 0x0E, 0xFF}, {0x2B, 0xF0, 0x35, 0xFF}, {0x0C, 0xF0, 0xA4, 0xFF},
	{0x05, 0xFB, 0xFF, 0xFF}, {0x5E, 0x5E, 0x5E, 0xFF}, {0x0D, 0x0D, 0x0D, 0xFF}, {0x0D, 0x0D, 0x0D, 0xFF},
	{0xFF, 0xFF, 0xFF, 0xFF}, {0xA6, 0xFC, 0xFF, 0xFF}, {0xB3, 0xEC, 0xFF, 0xFF}, {0xDA, 0xAB, 0xEB, 0xFF},
	{0xFF, 0xA8, 0xF9, 0xFF}, {0xFF, 0xAB, 0xB3, 0xFF}, {0xFF, 0xD2, 0xB0, 0xFF}, {0xFF, 0xEF, 0xA6, 0xFF},
	{0xFF, 0xF7, 0x9C, 0xFF}, {0xD7, 0xE8, 0x95, 0xFF}, {0xA6, 0xED, 0xAF, 0xFF}, {0xA2, 0xF2, 0xDA, 0xFF},
	{0x99, 0xFF, 0xFC, 0xFF}, {0xDD, 0xDD, 0xDD, 0xFF}, {0x11, 0x11, 0x11, 0xFF}, {0x11, 0x11, 0x11, 0xFF},
}

type background [height][width]*color.RGBA
type sprite [8][8]uint8
type palette [4]uint8

type PPU struct {
	bus        *PPUBus
	cycle      uint
	line       uint
	ppuctrl    uint8
	ppumask    uint8
	ppuscroll  uint8
	ppuaddr    uint16
	background *background
}

func New(ppuBus *PPUBus) *PPU {
	return &PPU{bus: ppuBus, background: &background{}}
}

func (p *PPU) ReadRegister(addr uint16) uint8 {
	switch addr {
	case 0x2002: // TODO
		return 0
	case 0x2007:
		tmp := p.ppuaddr
		p.ppuaddr++
		return p.bus.Read(tmp)
	default:
		fmt.Printf("ppu / ReadRegister 0x%x\n", addr)
		panic(1)
	}
}

func (p *PPU) WriteRegister(addr uint16, data uint8) {
	switch addr {
	case 0x2000:
		p.ppuctrl = data
	case 0x2001:
		p.ppumask = data
	case 0x2005:
		p.ppuscroll = data
	case 0x2006:
		p.ppuaddr = p.ppuaddr<<8 + uint16(data)
	case 0x2007:
		p.bus.Write(p.ppuaddr, data)
		p.ppuaddr++
	default:
		fmt.Printf("ppu / WriteRegister 0x%x\n", addr)
		panic(1)
	}
}

func (p *PPU) readByte(addr uint16) uint8 {
	return p.bus.Read(addr)
}

func (p *PPU) getAttribute(x uint, y uint) uint8 {
	addr := 0x23C0 + uint16(x/32+(y/32)*0x08)
	return p.readByte(addr)
}

func (p *PPU) getSpriteAddress(x uint, y uint) uint16 {
	addr := 0x2000 + uint16(x/8+(y/8)*0x20)
	return uint16(p.readByte(addr)) * 0x10
}

func (p *PPU) getSprite(x uint, y uint) *sprite {
	baseAddr := p.getSpriteAddress(x, y)
	s := sprite{}
	for i := uint16(0); i < 16; i++ {
		d := p.readByte(baseAddr + i)
		for j := 0; j < 8; j++ {
			s[i%8][j] += ((d >> (7 - j)) & 0b01) << (i / 8)
		}
	}
	return &s
}

func (p *PPU) getPalette(index uint8) *palette {
	baseAddr := 0x3F00 + uint16(0x04*index)
	palette := palette{}
	for i := uint16(0); i < 4; i++ {
		palette[i] = p.readByte(baseAddr + i)
	}
	return &palette
}

func (p *PPU) calcRGBA(x uint, y uint) *color.RGBA {
	attr := p.getAttribute(x, y)
	sprite := p.getSprite(x, y)
	index := (attr >> sprite[y%8][x%8]) & 0b11
	palette := p.getPalette(index)
	return &colors[palette[sprite[y%8][x%8]]]
}

func (p *PPU) Run(cycle uint) *background {
	p.cycle += cycle

	if p.cycle < cyclePerLine {
		return nil
	}
	p.cycle -= cyclePerLine

	if p.line < height {
		for i := uint(0); i < width; i++ {
			p.background[p.line][i] = p.calcRGBA(i, p.line)
		}
	}
	p.line++

	if p.line < height+vBlank {
		return nil
	}

	p.line = 0
	return p.background
}
