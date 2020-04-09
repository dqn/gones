package cpu

type StatusRegister struct {
	N, V, R, B, D, I, Z, C bool
}

type Registers struct {
	A, X, Y byte
	P       *StatusRegister
	SP      uint16
	PC      uint16
}

type CPU struct {
	Registers *Registers
	Memory    []byte
}

func (c *CPU) readWord(addr uint16) uint16 {
	return uint16(c.Memory[addr])<<8 + uint16(c.Memory[addr+1])
}

func (c *CPU) Reset() {
	c.Registers = &Registers{
		A: 0x00,
		X: 0x00,
		Y: 0x00,
		P: &StatusRegister{
			N: false,
			V: false,
			R: true,
			B: true,
			D: false,
			I: true,
			Z: false,
			C: false,
		},
		SP: 0x01FD,
		PC: 0x0000,
	}
	c.Registers.PC = c.readWord(0xFFFC)
}

func New(programROM []byte) *CPU {
	m := make([]byte, 0xFFFF+1)

	for i := 0; i < len(programROM); i++ {
		m[0x8000+i] = programROM[i]
	}
	c := &CPU{Memory: m}
	c.Reset()

	return c
}
