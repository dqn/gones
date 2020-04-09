package cpu

type word = uint16

type StatusRegister struct {
	N, V, R, B, D, I, Z, C bool
}

type Registers struct {
	A, X, Y byte
	P       *StatusRegister
	SP      word
	PC      word
}

type CPU struct {
	Registers *Registers
	Memory    []byte
}

func nthBit(value byte, n uint8) uint8 {
	return uint8((value >> n) & 1)
}

func (c *CPU) readByte(addr word) byte {
	return c.Memory[addr]
}

func (c *CPU) readWord(addr word) word {
	return word(c.Memory[addr])<<8 + word(c.Memory[addr+1])
}

func (c *CPU) write(addr word, data byte) {
	c.Memory[addr] = data
}

func (n *CPU) fetch() byte {
	addr := n.Registers.PC
	n.Registers.PC++
	return n.readByte(addr)
}

func (c *CPU) getByteByMode(opeland word, mode string) byte {
	switch mode {
	case "immediate":
		return byte(opeland)
	default:
		return c.readByte(opeland)
	}
}

func (c *CPU) exec(baseName string, opeland word, mode string) {
	switch baseName {
	case "ADC":
		m := c.getByteByMode(opeland, mode)
		if c.Registers.P.C {
			m++
		}
		c.Registers.P.V = c.Registers.A < 0x80 && c.Registers.A+m > 0x7F
		c.Registers.P.C = c.Registers.A+m <= c.Registers.A
		c.Registers.A += m
		c.Registers.P.N = nthBit(c.Registers.A, 7) == 1
		c.Registers.P.Z = c.Registers.A == 0
	case "SBC":
		m := c.getByteByMode(opeland, mode)
		if c.Registers.P.C {
			m++
		}
		c.Registers.P.V = c.Registers.A > 0x7F && c.Registers.A-m < 0x80
		c.Registers.P.C = c.Registers.A-m >= c.Registers.A
		c.Registers.A -= m
		c.Registers.P.N = nthBit(c.Registers.A, 7) == 1
		c.Registers.P.Z = c.Registers.A == 0
	case "AND":
		c.Registers.A &= c.getByteByMode(opeland, mode)
		c.Registers.P.N = nthBit(c.Registers.A, 7) == 1
		c.Registers.P.Z = c.Registers.A == 0
	case "ORA":
		c.Registers.A |= c.getByteByMode(opeland, mode)
		c.Registers.P.N = nthBit(c.Registers.A, 7) == 1
		c.Registers.P.Z = c.Registers.A == 0
	case "EOR":
		c.Registers.A ^= c.getByteByMode(opeland, mode)
		c.Registers.P.N = nthBit(c.Registers.A, 7) == 1
		c.Registers.P.Z = c.Registers.A == 0
	case "ASL":
		if mode == "accumulator" {
			c.Registers.P.C = nthBit(c.Registers.A, 7) == 1
			c.Registers.A <<= 1
			c.Registers.P.N = nthBit(c.Registers.A, 7) == 1
			c.Registers.P.Z = c.Registers.A == 0
		} else {
			data := c.readByte(opeland)
			c.Registers.P.C = nthBit(data, 7) == 1
			data <<= 1
			c.write(opeland, data)
			c.Registers.P.N = nthBit(data, 7) == 1
			c.Registers.P.Z = data == 0
		}
	case "LSR":
		if mode == "accumulator" {
			c.Registers.P.C = nthBit(c.Registers.A, 0) == 1
			c.Registers.A >>= 1
			c.Registers.P.Z = c.Registers.A == 0
		} else {
			data := c.readByte(opeland)
			c.Registers.P.C = nthBit(data, 0) == 1
			data >>= 1
			c.write(opeland, data)
			c.Registers.P.Z = data == 0
		}
		c.Registers.P.N = false
	case "ROL":
		if mode == "accumulator" {
			b := nthBit(c.Registers.A, 7)
			c.Registers.P.C = b == 1
			c.Registers.A = c.Registers.A<<1 + b
			c.Registers.P.N = nthBit(c.Registers.A, 7) == 1
			c.Registers.P.Z = c.Registers.A == 0
		} else {
			data := c.readByte(opeland)
			b := nthBit(data, 7)
			c.Registers.P.C = b == 1
			data = data<<1 + b
			c.write(opeland, data)
			c.Registers.P.N = nthBit(data, 7) == 1
			c.Registers.P.Z = data == 0
		}
	case "ROR":
		if mode == "accumulator" {
			b := nthBit(c.Registers.A, 0)
			c.Registers.P.C = b == 1
			c.Registers.A = c.Registers.A>>1 + b<<7
			c.Registers.P.N = nthBit(c.Registers.A, 7) == 1
			c.Registers.P.Z = c.Registers.A == 0
		} else {
			data := c.readByte(opeland)
			b := nthBit(data, 0)
			c.Registers.P.C = b == 1
			data = data>>1 + b<<7
			c.write(opeland, data)
			c.Registers.P.N = nthBit(data, 7) == 1
			c.Registers.P.Z = data == 0
		}
	case "LDA":
		c.Registers.A = c.getByteByMode(opeland, mode)
		c.Registers.P.N = nthBit(c.Registers.A, 7) == 1
		c.Registers.P.Z = c.Registers.A == 0
	case "STA":
		c.write(opeland, c.Registers.A)
	}
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
