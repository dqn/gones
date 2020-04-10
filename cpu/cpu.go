package cpu

import "fmt"

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

func nthBit(v byte, n uint8) uint8 {
	return uint8((v >> n) & 1)
}

func isNegative(v byte) bool {
	return nthBit(v, 7) == 1
}

func (c *CPU) readByte(addr word) byte {
	return c.Memory[addr]
}

func (c *CPU) readWord(addr word) word {
	return word(c.readByte(addr))<<8 + word(c.readByte(addr+1))
}

func (c *CPU) writeByte(addr word, data byte) {
	c.Memory[addr] = data
}

func (n *CPU) fetchByte() byte {
	addr := n.Registers.PC
	n.Registers.PC++
	return n.readByte(addr)
}

func (n *CPU) fetchWord() word {
	addr := n.Registers.PC
	n.Registers.PC += 2
	return n.readWord(addr)
}

func (c *CPU) push(data byte) {
	// c.writeByte(c.Registers.SP|0x0100, data)
	c.writeByte(c.Registers.SP, data)
	c.Registers.SP--
}

func (c *CPU) pop() byte {
	c.Registers.SP++
	// return c.readByte(c.Registers.SP|0x0100)
	return c.readByte(c.Registers.SP)
}

func (c *CPU) getByteByAddressing(opeland word, addressing string) byte {
	switch addressing {
	case "immediate":
		return byte(opeland)
	default:
		return c.readByte(opeland)
	}
}

func (c *CPU) fetchOpeland(addressing string) (word, error) {
	switch addressing {
	case "accumulator":
		return 0, nil
	case "implied":
		return 0, nil
	case "immediate":
		return word(c.fetchByte()), nil
	case "zeroPage":
		return word(c.fetchByte()), nil
	case "zeroPageX":
		addr := c.fetchByte()
		return word(addr + c.Registers.X), nil
	case "zeroPageY":
		addr := c.fetchByte()
		return word(addr + c.Registers.Y), nil
	case "absolute":
		return c.fetchWord(), nil
	case "absoluteX":
		addr := c.fetchWord()
		return addr + word(c.Registers.X), nil
	case "absoluteY":
		addr := c.fetchWord()
		return addr + word(c.Registers.Y), nil
	case "preIndexedIndirect":
		baseAddr := word(c.fetchByte()) + word(c.Registers.X)
		addr := word(c.readByte(baseAddr)) + word(c.readByte(baseAddr+1))<<8
		return addr, nil
	case "postIndexedIndirect":
		baseAddr := word(c.fetchByte())
		addr := word(c.readByte(baseAddr)) + word(c.readByte(baseAddr+1))<<8 + word(c.Registers.Y)
		return addr, nil
	case "indirectAbsolute":
		baseAddr := c.fetchWord()
		addr := word(c.readByte(baseAddr)) + word(c.readByte(baseAddr|((baseAddr)+1)))<<8
		return addr, nil
	default:
		return 0, fmt.Errorf("unknown addressing %s", addressing)
	}
}

func (c *CPU) exec(opcode string, opeland word, addressing string) error {
	switch opcode {
	case "ADC":
		m := c.getByteByAddressing(opeland, addressing)
		if c.Registers.P.C {
			m++
		}
		c.Registers.P.V = c.Registers.A < 0x80 && c.Registers.A+m > 0x7F
		c.Registers.P.C = c.Registers.A+m <= c.Registers.A
		c.Registers.A += m
		c.Registers.P.N = isNegative(c.Registers.A)
		c.Registers.P.Z = c.Registers.A == 0
	case "SBC":
		m := c.getByteByAddressing(opeland, addressing)
		if c.Registers.P.C {
			m++
		}
		c.Registers.P.V = c.Registers.A > 0x7F && c.Registers.A-m < 0x80
		c.Registers.P.C = c.Registers.A-m >= c.Registers.A
		c.Registers.A -= m
		c.Registers.P.N = isNegative(c.Registers.A)
		c.Registers.P.Z = c.Registers.A == 0
	case "AND":
		c.Registers.A &= c.getByteByAddressing(opeland, addressing)
		c.Registers.P.N = isNegative(c.Registers.A)
		c.Registers.P.Z = c.Registers.A == 0
	case "ORA":
		c.Registers.A |= c.getByteByAddressing(opeland, addressing)
		c.Registers.P.N = isNegative(c.Registers.A)
		c.Registers.P.Z = c.Registers.A == 0
	case "EOR":
		c.Registers.A ^= c.getByteByAddressing(opeland, addressing)
		c.Registers.P.N = isNegative(c.Registers.A)
		c.Registers.P.Z = c.Registers.A == 0
	case "ASL":
		if addressing == "accumulator" {
			c.Registers.P.C = nthBit(c.Registers.A, 7) == 1
			c.Registers.A <<= 1
			c.Registers.P.N = isNegative(c.Registers.A)
			c.Registers.P.Z = c.Registers.A == 0
		} else {
			data := c.readByte(opeland)
			c.Registers.P.C = nthBit(data, 7) == 1
			data <<= 1
			c.writeByte(opeland, data)
			c.Registers.P.N = isNegative(data)
			c.Registers.P.Z = data == 0
		}
	case "LSR":
		if addressing == "accumulator" {
			c.Registers.P.C = nthBit(c.Registers.A, 0) == 1
			c.Registers.A >>= 1
			c.Registers.P.N = isNegative(c.Registers.A)
			c.Registers.P.Z = c.Registers.A == 0
		} else {
			data := c.readByte(opeland)
			c.Registers.P.C = nthBit(data, 0) == 1
			data >>= 1
			c.writeByte(opeland, data)
			c.Registers.P.N = isNegative(data)
			c.Registers.P.Z = data == 0
		}
	case "ROL":
		if addressing == "accumulator" {
			b := nthBit(c.Registers.A, 7)
			c.Registers.P.C = b == 1
			c.Registers.A = c.Registers.A<<1 + b
			c.Registers.P.N = isNegative(c.Registers.A)
			c.Registers.P.Z = c.Registers.A == 0
		} else {
			data := c.readByte(opeland)
			b := nthBit(data, 7)
			c.Registers.P.C = b == 1
			data = data<<1 + b
			c.writeByte(opeland, data)
			c.Registers.P.N = isNegative(data)
			c.Registers.P.Z = data == 0
		}
	case "ROR":
		if addressing == "accumulator" {
			b := nthBit(c.Registers.A, 0)
			c.Registers.P.C = b == 1
			c.Registers.A = c.Registers.A>>1 + b<<7
			c.Registers.P.N = isNegative(c.Registers.A)
			c.Registers.P.Z = c.Registers.A == 0
		} else {
			data := c.readByte(opeland)
			b := nthBit(data, 0)
			c.Registers.P.C = b == 1
			data = data>>1 + b<<7
			c.writeByte(opeland, data)
			c.Registers.P.N = isNegative(data)
			c.Registers.P.Z = data == 0
		}
	case "BCC":
		if !c.Registers.P.C {
			c.Registers.PC = opeland
		}
	case "BCS":
		if c.Registers.P.C {
			c.Registers.PC = opeland
		}
	case "BNE":
		if !c.Registers.P.Z {
			c.Registers.PC = opeland
		}
	case "BEQ":
		if c.Registers.P.Z {
			c.Registers.PC = opeland
		}
	case "BVC":
		if !c.Registers.P.V {
			c.Registers.PC = opeland
		}
	case "BVS":
		if c.Registers.P.V {
			c.Registers.PC = opeland
		}
	case "BPL":
		if !c.Registers.P.N {
			c.Registers.PC = opeland
		}
	case "BMI":
		if c.Registers.P.N {
			c.Registers.PC = opeland
		}
	case "BIT":
		data := c.readByte(opeland)
		c.Registers.P.Z = (c.Registers.A & data) == 0
		c.Registers.P.N = nthBit(data, 7) == 1
		c.Registers.P.V = nthBit(data, 6) == 1
	case "JMP":
		c.Registers.PC = opeland
	case "JSR":
		pc := c.Registers.PC - 1
		c.push(byte(pc >> 8))
		c.push(byte(pc))
	case "RTS":
		c.Registers.PC = word(c.pop()) + word(c.pop())<<8 + 1
	case "BRK": // TODO
		// if !c.Registers.P.I {
		// 	c.Registers.P.B = true
		// 	c.Registers.PC++
		// 	c.push(byte(c.Registers.PC >> 8))
		// 	c.push(byte(c.Registers.PC))
		// 	c.push(byte(c.Registers.P))
		// }
	case "RTI": // TODO
		// c.Registers.P = c.pop()
		// c.Registers.PC = word(c.pop()) + word(c.pop())<<8 + 1
	case "CMP":
		data := c.Registers.A - c.getByteByAddressing(opeland, addressing)
		c.Registers.P.N = isNegative(data)
		c.Registers.P.Z = data == 0
		c.Registers.P.C = data >= c.Registers.A
	case "CPX":
		data := c.Registers.X - c.getByteByAddressing(opeland, addressing)
		c.Registers.P.N = isNegative(data)
		c.Registers.P.Z = data == 0
		c.Registers.P.C = data >= c.Registers.X
	case "CPY":
		data := c.Registers.Y - c.getByteByAddressing(opeland, addressing)
		c.Registers.P.N = isNegative(data)
		c.Registers.P.Z = data == 0
		c.Registers.P.C = data >= c.Registers.Y
	case "INC":
		m := c.readByte(opeland) + 1
		c.Registers.P.N = isNegative(m)
		c.Registers.P.Z = m == 0
		c.writeByte(opeland, m)
	case "DEC":
		m := c.readByte(opeland) - 1
		c.Registers.P.N = isNegative(m)
		c.Registers.P.Z = m == 0
		c.writeByte(opeland, m)
	case "INX":
		c.Registers.X++
		c.Registers.P.N = isNegative(c.Registers.X)
		c.Registers.P.Z = c.Registers.X == 0
	case "DEX":
		c.Registers.X--
		c.Registers.P.N = isNegative(c.Registers.X)
		c.Registers.P.Z = c.Registers.X == 0
	case "INY":
		c.Registers.Y++
		c.Registers.P.N = isNegative(c.Registers.Y)
		c.Registers.P.Z = c.Registers.Y == 0
	case "DEY":
		c.Registers.Y--
		c.Registers.P.N = isNegative(c.Registers.Y)
		c.Registers.P.Z = c.Registers.Y == 0
	case "CLC":
		c.Registers.P.C = false
	case "SEC":
		c.Registers.P.C = true
	case "CLI":
		c.Registers.P.I = false
	case "SEI":
		c.Registers.P.I = true
	case "CLD":
		c.Registers.P.D = false
	case "SED":
		c.Registers.P.D = true
	case "CLV":
		c.Registers.P.V = false
	case "LDA":
		c.Registers.A = c.getByteByAddressing(opeland, addressing)
		c.Registers.P.N = isNegative(c.Registers.A)
		c.Registers.P.Z = c.Registers.A == 0
	case "LDX":
		c.Registers.X = c.getByteByAddressing(opeland, addressing)
		c.Registers.P.N = isNegative(c.Registers.X)
		c.Registers.P.Z = c.Registers.X == 0
	case "LDY":
		c.Registers.Y = c.getByteByAddressing(opeland, addressing)
		c.Registers.P.N = isNegative(c.Registers.Y)
		c.Registers.P.Z = c.Registers.Y == 0
	case "STA":
		c.writeByte(opeland, c.Registers.A)
	case "STX":
		c.writeByte(opeland, c.Registers.X)
	case "STY":
		c.writeByte(opeland, c.Registers.Y)
	case "TXA":
		c.Registers.A = c.Registers.X
		c.Registers.P.N = isNegative(c.Registers.A)
		c.Registers.P.Z = c.Registers.A == 0
	case "TAX":
		c.Registers.X = c.Registers.A
		c.Registers.P.N = isNegative(c.Registers.X)
		c.Registers.P.Z = c.Registers.X == 0
	case "TAY":
		c.Registers.Y = c.Registers.A
		c.Registers.P.N = isNegative(c.Registers.Y)
		c.Registers.P.Z = c.Registers.Y == 0
	case "TYA":
		c.Registers.A = c.Registers.Y
		c.Registers.P.N = isNegative(c.Registers.A)
		c.Registers.P.Z = c.Registers.A == 0
	case "TSX":
		c.Registers.X = byte(c.Registers.SP)
		c.Registers.P.N = isNegative(c.Registers.X)
		c.Registers.P.Z = c.Registers.X == 0
	case "TXS":
		c.Registers.SP = word(c.Registers.X)
	case "PHA":
		c.push(c.Registers.A)
	case "PLA":
		c.Registers.A = c.pop()
		c.Registers.P.N = isNegative(c.Registers.A)
		c.Registers.P.Z = c.Registers.A == 0
	case "PHP":
		// TODO
	case "PLP":
		// TODO
	case "NOP":
		// no operation
	default:
		return fmt.Errorf("unknown opcode: %s", opcode)
	}
	return nil
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
