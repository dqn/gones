package cpu

import (
	"fmt"

	"github.com/dqn/gones/controller"
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
	controller *controller.Controller
}

func NewBus(ram *ram.RAM, programROM []uint8, ppu *ppu.PPU, controller *controller.Controller) *CPUBus {
	return &CPUBus{ram, programROM, ppu, controller}
}

func (b *CPUBus) Read(addr uint16) uint8 {
	switch {
	case addr >= 0x0000 && addr < 0x0800:
		return b.ram[addr]
	case addr >= 0x0800 && addr < 0x2000:
		return b.ram[addr-0x0800]
	case addr >= 0x2000 && addr < 0x2008:
		return b.ppu.ReadRegister(addr)
	case addr >= 0x2008 && addr < 0x4000:
		return b.ppu.ReadRegister(addr - 0x0008)
	case addr == 0x4016:
		return b.controller.ReadButton()
	case addr >= 0xC000 && addr <= 0xFFFF:
		if len(b.programROM) <= 0x4000 {
			return b.programROM[addr-0xC000]
		}
		return b.programROM[addr-0x8000]
	case addr >= 0x8000 && addr <= 0xFFFF:
		return b.programROM[addr-0x8000]
	default:
		fmt.Printf("!!! cpu bus / Read 0x%x\n", addr)
		panic(1)
	}
}

func (b *CPUBus) Write(addr uint16, data uint8) {
	switch {
	case addr >= 0x0000 && addr < 0x0800:
		b.ram[addr] = data
	case addr >= 0x0800 && addr < 0x2000:
		b.ram[addr-0x0800] = data
	case addr >= 0x2000 && addr < 0x2008:
		b.ppu.WriteRegister(addr, data)
	case addr >= 0x2008 && addr < 0x4000:
		b.ppu.WriteRegister(addr-0x0008, data)
	case addr == 0x4014:
		baseAddr := uint16(data) << 8
		b.ppu.DMA(b.ram[baseAddr : baseAddr+0x0100])
	case addr == 0x4016:
		b.controller.Clear()
	case addr >= 0x8000 && addr <= 0xFFFF:
		b.programROM[addr-0x8000] = data
	default:
		fmt.Printf("!!! cpu bus / Write 0x%x\n", addr)
		panic(1)
	}
}
