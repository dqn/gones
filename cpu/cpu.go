package cpu

import (
	"fmt"
)

type memory = [0x10000]uint8

type StatusRegister struct {
	N, V, R, B, D, I, Z, C bool
}

type Registers struct {
	A, X, Y uint8
	P       *StatusRegister
	SP      uint16
	PC      uint16
}

type CPU struct {
	Registers *Registers
	Memory    *memory
}

func nthBit(v uint8, n uint8) uint8 {
	return uint8((v >> n) & 1)
}

func isNegative(v uint8) bool {
	return nthBit(v, 7) == 1
}

func New(programROM []uint8) *CPU {
	var m memory
	for i := 0; i < len(programROM); i++ {
		m[0x8000+i] = programROM[i]
	}
	c := &CPU{Memory: &m}
	c.Reset()

	return c
}

func (c *CPU) readByte(addr uint16) uint8 {
	return c.Memory[addr]
}

func (c *CPU) readWord(addr uint16) uint16 {
	return uint16(c.readByte(addr)) + uint16(c.readByte(addr+1))<<8
}

func (c *CPU) writeByte(addr uint16, data uint8) {
	c.Memory[addr] = data
}

func (n *CPU) fetchByte() uint8 {
	addr := n.Registers.PC
	n.Registers.PC++
	return n.readByte(addr)
}

func (n *CPU) fetchWord() uint16 {
	addr := n.Registers.PC
	n.Registers.PC += 2
	return n.readWord(addr)
}

func (c *CPU) push(data uint8) {
	// c.writeByte(c.Registers.SP|0x0100, data)
	c.writeByte(c.Registers.SP, data)
	c.Registers.SP--
}

func (c *CPU) pop() uint8 {
	c.Registers.SP++
	// return c.readByte(c.Registers.SP|0x0100)
	return c.readByte(c.Registers.SP)
}

func (c *CPU) getByteByAddressing(opeland uint16, addressing string) uint8 {
	switch addressing {
	case "Immediate":
		return uint8(opeland)
	default:
		return c.readByte(opeland)
	}
}

func (c *CPU) fetchOpeland(addressing string) (uint16, error) {
	var opeland uint16
	switch addressing {
	case "Accumulator":
		// no opeland
	case "Implied":
		// no opeland
	case "Immediate":
		opeland = uint16(c.fetchByte())
	case "Absolute":
		opeland = c.fetchWord()
	case "Absolute, X":
		opeland = c.fetchWord() + uint16(c.Registers.X)
	case "Absolute, Y":
		opeland = c.fetchWord() + uint16(c.Registers.Y)
	case "Zeropage":
		opeland = uint16(c.fetchByte())
	case "Zeropage, X":
		opeland = uint16(c.fetchByte() + c.Registers.X)
	case "Zeropage, Y":
		opeland = uint16(c.fetchByte() + c.Registers.Y)
	case "Relative":
		opeland = uint16(c.fetchByte()) + c.Registers.PC - 0xFF
	case "(Indirect, X)":
		baseAddr := uint16(c.fetchByte()) + uint16(c.Registers.X)
		opeland = uint16(c.readByte(baseAddr)) + uint16(c.readByte(baseAddr+1))<<8
	case "(Indirect), Y":
		baseAddr := uint16(c.fetchByte())
		opeland = uint16(c.readByte(baseAddr)) + uint16(c.readByte(baseAddr+1))<<8 + uint16(c.Registers.Y)
	case "(Indirect)":
		baseAddr := c.fetchWord()
		opeland = uint16(c.readByte(baseAddr)) + uint16(c.readByte((baseAddr&0xFF00)|(((baseAddr&0xFF)+1)&0xFF)))<<8
	default:
		return 0, fmt.Errorf("unknown addressing %s", addressing)
	}
	return opeland, nil
}

func (c *CPU) exec(opcode string, opeland uint16, addressing string) error {
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
		if addressing == "Accumulator" {
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
		if addressing == "Accumulator" {
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
		if addressing == "Accumulator" {
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
		if addressing == "Accumulator" {
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
		c.push(uint8(pc >> 8))
		c.push(uint8(pc))
	case "RTS":
		c.Registers.PC = uint16(c.pop()) + uint16(c.pop())<<8 + 1
	case "BRK": // TODO
		// if !c.Registers.P.I {
		// 	c.Registers.P.B = true
		// 	c.Registers.PC++
		// 	c.push(uint8(c.Registers.PC >> 8))
		// 	c.push(uint8(c.Registers.PC))
		// 	c.push(uint8(c.Registers.P))
		// }
	case "RTI": // TODO
		// c.Registers.P = c.pop()
		// c.Registers.PC = uint16(c.pop()) + uint16(c.pop())<<8 + 1
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
		c.Registers.X = uint8(c.Registers.SP)
		c.Registers.P.N = isNegative(c.Registers.X)
		c.Registers.P.Z = c.Registers.X == 0
	case "TXS":
		c.Registers.SP = uint16(c.Registers.X)
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

func (c *CPU) Run() (int, error) {
	b := c.fetchByte()
	i := instructionSets[b]
	// fmt.Printf("%x %x: %v\n", c.Registers.PC-1, b, i)
	opeland, err := c.fetchOpeland(i.Addressing)
	if err != nil {
		return 0, err
	}
	err = c.exec(i.Opcode, opeland, i.Addressing)
	if err != nil {
		return 0, err
	}

	return i.Cycle, nil
}
