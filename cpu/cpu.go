package cpu

import (
	"fmt"
)

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
	registers *Registers
	bus       *CPUBus
}

func nthBit(v uint8, n uint8) uint8 {
	return uint8((v >> n) & 1)
}

func isNegative(v uint8) bool {
	return nthBit(v, 7) == 1
}

func New(cpuBus *CPUBus) *CPU {
	c := &CPU{bus: cpuBus}
	c.Reset()

	return c
}

func (c *CPU) readByte(addr uint16) uint8 {
	return c.bus.Read(addr)
}

func (c *CPU) readWord(addr uint16) uint16 {
	return uint16(c.readByte(addr)) + uint16(c.readByte(addr+1))<<8
}

func (c *CPU) writeByte(addr uint16, data uint8) {
	c.bus.Write(addr, data)
}

func (n *CPU) fetchByte() uint8 {
	addr := n.registers.PC
	n.registers.PC++
	return n.readByte(addr)
}

func (n *CPU) fetchWord() uint16 {
	addr := n.registers.PC
	n.registers.PC += 2
	return n.readWord(addr)
}

func (c *CPU) push(data uint8) {
	c.writeByte(c.registers.SP, data)
	c.registers.SP--
}

func (c *CPU) pop() uint8 {
	c.registers.SP++
	return c.readByte(c.registers.SP)
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
		opeland = c.fetchWord() + uint16(c.registers.X)
	case "Absolute, Y":
		opeland = c.fetchWord() + uint16(c.registers.Y)
	case "Zeropage":
		opeland = uint16(c.fetchByte())
	case "Zeropage, X":
		opeland = uint16(c.fetchByte() + c.registers.X)
	case "Zeropage, Y":
		opeland = uint16(c.fetchByte() + c.registers.Y)
	case "Relative":
		baseAddr := uint16(c.fetchByte())
		if baseAddr < 0x80 {
			opeland = baseAddr
		} else {
			opeland = baseAddr - 0x0100
		}
		opeland += c.registers.PC
	case "(Indirect, X)":
		baseAddr := uint16(c.fetchByte()) + uint16(c.registers.X)
		opeland = uint16(c.readByte(baseAddr)) + uint16(c.readByte(baseAddr+1))<<8
	case "(Indirect), Y":
		baseAddr := uint16(c.fetchByte())
		opeland = uint16(c.readByte(baseAddr)) + uint16(c.readByte(baseAddr+1))<<8 + uint16(c.registers.Y)
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
		if c.registers.P.C {
			m++
		}
		c.registers.P.V = c.registers.A < 0x80 && c.registers.A+m > 0x7F
		c.registers.P.C = c.registers.A+m <= c.registers.A
		c.registers.A += m
		c.registers.P.N = isNegative(c.registers.A)
		c.registers.P.Z = c.registers.A == 0
	case "SBC":
		m := c.getByteByAddressing(opeland, addressing)
		if c.registers.P.C {
			m++
		}
		c.registers.P.V = c.registers.A > 0x7F && c.registers.A-m < 0x80
		c.registers.P.C = c.registers.A-m >= c.registers.A
		c.registers.A -= m
		c.registers.P.N = isNegative(c.registers.A)
		c.registers.P.Z = c.registers.A == 0
	case "AND":
		c.registers.A &= c.getByteByAddressing(opeland, addressing)
		c.registers.P.N = isNegative(c.registers.A)
		c.registers.P.Z = c.registers.A == 0
	case "ORA":
		c.registers.A |= c.getByteByAddressing(opeland, addressing)
		c.registers.P.N = isNegative(c.registers.A)
		c.registers.P.Z = c.registers.A == 0
	case "EOR":
		c.registers.A ^= c.getByteByAddressing(opeland, addressing)
		c.registers.P.N = isNegative(c.registers.A)
		c.registers.P.Z = c.registers.A == 0
	case "ASL":
		if addressing == "Accumulator" {
			c.registers.P.C = nthBit(c.registers.A, 7) == 1
			c.registers.A <<= 1
			c.registers.P.N = isNegative(c.registers.A)
			c.registers.P.Z = c.registers.A == 0
		} else {
			data := c.readByte(opeland)
			c.registers.P.C = nthBit(data, 7) == 1
			data <<= 1
			c.writeByte(opeland, data)
			c.registers.P.N = isNegative(data)
			c.registers.P.Z = data == 0
		}
	case "LSR":
		if addressing == "Accumulator" {
			c.registers.P.C = nthBit(c.registers.A, 0) == 1
			c.registers.A >>= 1
			c.registers.P.N = isNegative(c.registers.A)
			c.registers.P.Z = c.registers.A == 0
		} else {
			data := c.readByte(opeland)
			c.registers.P.C = nthBit(data, 0) == 1
			data >>= 1
			c.writeByte(opeland, data)
			c.registers.P.N = isNegative(data)
			c.registers.P.Z = data == 0
		}
	case "ROL":
		if addressing == "Accumulator" {
			b := nthBit(c.registers.A, 7)
			c.registers.P.C = b == 1
			c.registers.A = c.registers.A<<1 + b
			c.registers.P.N = isNegative(c.registers.A)
			c.registers.P.Z = c.registers.A == 0
		} else {
			data := c.readByte(opeland)
			b := nthBit(data, 7)
			c.registers.P.C = b == 1
			data = data<<1 + b
			c.writeByte(opeland, data)
			c.registers.P.N = isNegative(data)
			c.registers.P.Z = data == 0
		}
	case "ROR":
		if addressing == "Accumulator" {
			b := nthBit(c.registers.A, 0)
			c.registers.P.C = b == 1
			c.registers.A = c.registers.A>>1 + b<<7
			c.registers.P.N = isNegative(c.registers.A)
			c.registers.P.Z = c.registers.A == 0
		} else {
			data := c.readByte(opeland)
			b := nthBit(data, 0)
			c.registers.P.C = b == 1
			data = data>>1 + b<<7
			c.writeByte(opeland, data)
			c.registers.P.N = isNegative(data)
			c.registers.P.Z = data == 0
		}
	case "BCC":
		if !c.registers.P.C {
			c.registers.PC = opeland
		}
	case "BCS":
		if c.registers.P.C {
			c.registers.PC = opeland
		}
	case "BNE":
		if !c.registers.P.Z {
			c.registers.PC = opeland
		}
	case "BEQ":
		if c.registers.P.Z {
			c.registers.PC = opeland
		}
	case "BVC":
		if !c.registers.P.V {
			c.registers.PC = opeland
		}
	case "BVS":
		if c.registers.P.V {
			c.registers.PC = opeland
		}
	case "BPL":
		if !c.registers.P.N {
			c.registers.PC = opeland
		}
	case "BMI":
		if c.registers.P.N {
			c.registers.PC = opeland
		}
	case "BIT":
		data := c.readByte(opeland)
		c.registers.P.Z = (c.registers.A & data) == 0
		c.registers.P.N = nthBit(data, 7) == 1
		c.registers.P.V = nthBit(data, 6) == 1
	case "JMP":
		c.registers.PC = opeland
	case "JSR":
		pc := c.registers.PC - 1
		c.push(uint8(pc >> 8))
		c.push(uint8(pc))
	case "RTS":
		c.registers.PC = uint16(c.pop()) + uint16(c.pop())<<8 + 1
	case "BRK": // TODO
		// if !c.registers.P.I {
		// 	c.registers.P.B = true
		// 	c.registers.PC++
		// 	c.push(uint8(c.registers.PC >> 8))
		// 	c.push(uint8(c.registers.PC))
		// 	c.push(uint8(c.registers.P))
		// }
	case "RTI": // TODO
		// c.registers.P = c.pop()
		// c.registers.PC = uint16(c.pop()) + uint16(c.pop())<<8 + 1
	case "CMP":
		data := c.registers.A - c.getByteByAddressing(opeland, addressing)
		c.registers.P.N = isNegative(data)
		c.registers.P.Z = data == 0
		c.registers.P.C = data >= c.registers.A
	case "CPX":
		data := c.registers.X - c.getByteByAddressing(opeland, addressing)
		c.registers.P.N = isNegative(data)
		c.registers.P.Z = data == 0
		c.registers.P.C = data >= c.registers.X
	case "CPY":
		data := c.registers.Y - c.getByteByAddressing(opeland, addressing)
		c.registers.P.N = isNegative(data)
		c.registers.P.Z = data == 0
		c.registers.P.C = data >= c.registers.Y
	case "INC":
		m := c.readByte(opeland) + 1
		c.registers.P.N = isNegative(m)
		c.registers.P.Z = m == 0
		c.writeByte(opeland, m)
	case "DEC":
		m := c.readByte(opeland) - 1
		c.registers.P.N = isNegative(m)
		c.registers.P.Z = m == 0
		c.writeByte(opeland, m)
	case "INX":
		c.registers.X++
		c.registers.P.N = isNegative(c.registers.X)
		c.registers.P.Z = c.registers.X == 0
	case "DEX":
		c.registers.X--
		c.registers.P.N = isNegative(c.registers.X)
		c.registers.P.Z = c.registers.X == 0
	case "INY":
		c.registers.Y++
		c.registers.P.N = isNegative(c.registers.Y)
		c.registers.P.Z = c.registers.Y == 0
	case "DEY":
		c.registers.Y--
		c.registers.P.N = isNegative(c.registers.Y)
		c.registers.P.Z = c.registers.Y == 0
	case "CLC":
		c.registers.P.C = false
	case "SEC":
		c.registers.P.C = true
	case "CLI":
		c.registers.P.I = false
	case "SEI":
		c.registers.P.I = true
	case "CLD":
		c.registers.P.D = false
	case "SED":
		c.registers.P.D = true
	case "CLV":
		c.registers.P.V = false
	case "LDA":
		c.registers.A = c.getByteByAddressing(opeland, addressing)
		c.registers.P.N = isNegative(c.registers.A)
		c.registers.P.Z = c.registers.A == 0
	case "LDX":
		c.registers.X = c.getByteByAddressing(opeland, addressing)
		c.registers.P.N = isNegative(c.registers.X)
		c.registers.P.Z = c.registers.X == 0
	case "LDY":
		c.registers.Y = c.getByteByAddressing(opeland, addressing)
		c.registers.P.N = isNegative(c.registers.Y)
		c.registers.P.Z = c.registers.Y == 0
	case "STA":
		c.writeByte(opeland, c.registers.A)
	case "STX":
		c.writeByte(opeland, c.registers.X)
	case "STY":
		c.writeByte(opeland, c.registers.Y)
	case "TXA":
		c.registers.A = c.registers.X
		c.registers.P.N = isNegative(c.registers.A)
		c.registers.P.Z = c.registers.A == 0
	case "TAX":
		c.registers.X = c.registers.A
		c.registers.P.N = isNegative(c.registers.X)
		c.registers.P.Z = c.registers.X == 0
	case "TAY":
		c.registers.Y = c.registers.A
		c.registers.P.N = isNegative(c.registers.Y)
		c.registers.P.Z = c.registers.Y == 0
	case "TYA":
		c.registers.A = c.registers.Y
		c.registers.P.N = isNegative(c.registers.A)
		c.registers.P.Z = c.registers.A == 0
	case "TSX":
		c.registers.X = uint8(c.registers.SP)
		c.registers.P.N = isNegative(c.registers.X)
		c.registers.P.Z = c.registers.X == 0
	case "TXS":
		c.registers.SP = uint16(c.registers.X)
	case "PHA":
		c.push(c.registers.A)
	case "PLA":
		c.registers.A = c.pop()
		c.registers.P.N = isNegative(c.registers.A)
		c.registers.P.Z = c.registers.A == 0
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
	c.registers = &Registers{
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
	c.registers.PC = c.readWord(0xFFFC)
}

func (c *CPU) Run() (uint, error) {
	b := c.fetchByte()
	i := instructionSets[b]
	// fmt.Printf("%x %x: %v\n", c.registers.PC-1, b, i)
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
