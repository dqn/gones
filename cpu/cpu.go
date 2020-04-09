package cpu

const (
	memorySize            = 0xFFFF + 1
	resetInterruptAddress = 0xFFFC
)

type StatusRegister struct {
	N, V, R, B, D, I, Z, C bool
}

type Registers struct {
	A, X, Y, S byte
	P          StatusRegister
	PC         uint16
}

type CPU struct {
	Registers *Registers
	Memory    []byte
}

func (c *CPU) readWord(addr uint16) uint16 {
	return uint16(c.Memory[addr])<<8 + uint16(c.Memory[addr+1])
}

func (c *CPU) Reset() {
	c.Registers.PC = c.readWord(resetInterruptAddress)
}

func New(programROM []byte) *CPU {
	m := make([]byte, memorySize)
	for i := 0; i < len(programROM); i++ {
		m[0x8000+i] = programROM[i]
	}

	c := &CPU{&Registers{}, m}
	c.Reset()

	return c
}
