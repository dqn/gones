package ppu

// https://wiki.nesdev.com/w/index.php/PPU_registers
// PPUCTRL
// bit7[V]:    NMI enable
// bit6[P]:    PPU master/slave
// bit5[H]:    sprite height
// bit4[B]:    background tile select
// bit3[S]:    sprite tile select
// bit2[I]:    increment mode
// bit1-0[NN]: nametable select

type ppuctrl uint8

func (p *ppuctrl) GetBGPatternBaseAddress() uint16 {
	if *p&0b00010000 == 0 {
		return 0x0000
	} else {
		return 0x1000
	}
}

func (p *ppuctrl) GetSpritePatternBaseAddress() uint16 {
	if *p&0b00001000 == 0 {
		return 0x0000
	} else {
		return 0x1000
	}
}

func (p *ppuctrl) GetIncrementSize() uint16 {
	if *p&0b00000100 == 0 {
		return 1
	} else {
		return 32
	}
}

func (p *ppuctrl) GetNameTableBaseAddress() uint16 {
	switch *p & 0b00000011 {
	case 0b00:
		return 0x2000
	case 0b01:
		return 0x2400
	case 0b10:
		return 0x2800
	case 0b11:
		return 0x2C00
	default:
		panic("system error")
	}
}
