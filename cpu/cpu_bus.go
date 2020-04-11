package cpu

import (
	"github.com/dqn/gones/ppu"
	"github.com/dqn/gones/ram"
)

// https://qiita.com/bokuweb/items/1575337bef44ae82f4d3#%E3%83%A1%E3%83%A2%E3%83%AA%E3%83%9E%E3%83%83%E3%83%97-1

// アドレス	       サイズ   用途
// 0x0000～0x07FF	0x0800	WRAM
// 0x0800～0x1FFF	-	      WRAM のミラー
// 0x2000～0x2007	0x0008	PPU レジスタ
// 0x2008～0x3FFF	-	      PPU レジスタのミラー
// 0x4000～0x401F	0x0020	APU I/O、PAD
// 0x4020～0x5FFF	0x1FE0	拡張 ROM
// 0x6000～0x7FFF	0x2000	拡張 RAM
// 0x8000～0xBFFF	0x4000	PRG-ROM
// 0xC000～0xFFFF	0x4000	PRG-ROM

type CPUBus struct {
	ram        *ram.RAM
	programROM []uint8
	ppu        *ppu.PPU
}

func NewBus(ram *ram.RAM, programROM []uint8, ppu *ppu.PPU) *CPUBus {
	return &CPUBus{ram, programROM, ppu}
}

func (b *CPUBus) Read(addr uint16) uint8 {
	switch {
	case addr >= 0x8000:
		return b.programROM[addr-0x8000]
	default:
		return b.ram[addr]
	}
}

func (b *CPUBus) Write(addr uint16, data uint8) {
	b.ram[addr] = data
}
