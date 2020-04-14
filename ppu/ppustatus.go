package ppu

// https://wiki.nesdev.com/w/index.php/PPU_registers
// PPUSTATUS
// bit7[V]: vblank
// bit6[S]: sprite 0 hit
// bit5[O]: sprite overflow

type ppustatus uint8

func (p *ppustatus) Uint8() uint8 {
	return uint8(*p)
}

func (p *ppustatus) SetVBlank(b bool) {
	*p |= 0b10000000
}
