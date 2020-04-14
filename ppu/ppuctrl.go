package ppu

type ppuctrl uint8

func (p ppuctrl) getBGPatternBaseAddress() uint16 {
	if p&0b00010000 == 0 {
		return 0x0000
	} else {
		return 0x1000
	}
}
