package cpu

import (
	// "fmt"
	// "log"
	"math/bits"
)

/*
https://web.archive.org/web/20240118230916/http://www.emulator101.com/finishing-the-cpu-emulator.html

reference implementation:
https://github.com/hlboehm/i8080-emulator/blob/main/src/emulator/processor.c
*/

/*
TODO:
- make flag setter function
- make psw register
- make helper functions for d16 ops
*/
const memSize = 0xffff

type CPU struct {
	memory []uint8

	A, B, C, D, E, H, L uint8
	PSW                 uint8
	SP, PC              uint16
	portMap             map[uint8]uint8

	// Flags               uint8
	/*FLAGS*/
	sign, zero, auxCarry, parity, carry bool
	interrupt                           bool
	Cycles                               int

	/*
		spinCounter uint8 // number of cycles to spin
	*/
}

/*
	GETTER FUNCTIONS

For packages reading from the cpu
*/
func (c *CPU) GetFlags() []bool {
	return []bool{c.sign, c.zero, c.auxCarry, c.parity, c.carry}
}

func (c *CPU) GetRegs() ([]uint8, []uint16) {
	return []uint8{c.A, c.B, c.C, c.D, c.E, c.H, c.L}, []uint16{c.SP, c.PC}
}

func (c *CPU) GetMemory() []uint8 { return c.memory }

func New(memory []uint8) *CPU {
	return &CPU{
		// memory: [memSize]uint8{0},
		memory: memory,
		A:      0,
		B:      0,
		C:      0,
		D:      0,
		E:      0,
		H:      0,
		L:      0,
		PSW:    0,
		// Flags:  0,
		SP:        0,
		PC:        0,
		sign:      false,
		zero:      false,
		auxCarry:  false,
		parity:    false,
		carry:     false,
		interrupt: false,
		portMap:   make(map[uint8]uint8),
	}
}

/****  MAIN RUN FUNCTION  ****/
func (c *CPU) Run() {
	// run 1 cycle
	c.Cycles += 1
	var op uint8

	// TODO check if PC is in range
	// if c.PC >= 0xFFFF{
	// 	log.Fatalf("ERROR: PC out of range: %X", c.PC)
	// }
	op = c.memory[c.PC]

	switch op {
	case 0x00:
		// NOP
		c.PC += 1
	case 0x01:
		c.B = c.memory[c.PC+2]
		c.C = c.memory[c.PC+1]
		c.PC += 3
	case 0x05:
		c.B -= 1
		c.zero = c.B == 0
		c.sign = (c.B & 0x80) != 0 // if the msb is NOT zero, then negative
		c.parity = (bits.OnesCount8(c.B) % 2) == 0
		c.auxCarry = false // TODO: aux carry
		c.PC += 1
	case 0x06:
		c.B = c.memory[c.PC+1]
		c.PC += 2
	case 0x09:
		// HL = HL + BC
		sum, carryBit := Add16(c.getHL(), c.getBC())
		c.setHL(sum)
		c.carry = carryBit == 1
		c.PC += 1
	case 0x0d:
		c.C -= 1
		c.zero = c.C == 0
		c.sign = (c.C & 0x80) != 0 // if the msb is NOT zero, then negative
		c.parity = (bits.OnesCount8(c.C) % 2) == 0
		c.auxCarry = false // TODO: aux carry
		c.PC += 1
	case 0x0e:
		c.C = c.memory[c.PC+1]
		c.PC += 1
	case 0x0f:
		// RRC
		// carry = previous bit 0
		c.carry = (c.A & 0b1) == 1 // check if smallest bit is 1
		c.A = bits.RotateLeft8(c.A, -1)
		c.PC += 1
	case 0x11:
		c.D = c.memory[c.PC+2]
		c.E = c.memory[c.PC+1]
		c.PC += 3
	case 0x13:
		c.setDE(c.getDE() + 1)
		c.PC += 1
	case 0x19:
		// I start using carry functions here
		sum, carryBit := Add16(c.getHL(), c.getDE())
		c.setHL(sum)
		c.carry = carryBit == 1
		c.PC += 1
	case 0x1a:
		c.A = c.memory[c.getDE()]
		c.PC += 1
	case 0x21:
		c.H = c.memory[c.PC+2]
		c.L = c.memory[c.PC+1]
		c.PC += 3
	case 0x23:
		c.setHL(c.getHL() + 1)
		c.PC += 1
	case 0x26:
		c.H = c.memory[c.PC+1]
		c.PC += 2
	case 0x29:
		hl := c.getHL()
		sum, carryBit := Add16(hl, hl)
		c.setHL(sum)
		c.carry = carryBit == 1
		c.PC += 1
	case 0x31:
		c.SP = uint16(c.memory[c.PC+2])<<8 | uint16(c.memory[c.PC+1])
		c.PC += 3
	case 0x32:
		adr := uint16(c.memory[c.PC+2])<<8 | uint16(c.memory[c.PC+1])
		c.memory[adr] = c.A
		c.PC += 3
	case 0x36:
		c.memory[c.getHL()] = c.memory[c.PC+1]
		c.PC += 2
	case 0x3a:
		adr := uint16(c.memory[c.PC+2])<<8 | uint16(c.memory[c.PC+1])
		c.A = c.memory[adr]
		c.PC += 3
	case 0x3e:
		c.A = c.memory[c.PC+1]
		c.PC += 2
	case 0x56:
		c.D = c.memory[c.getHL()]
		c.PC += 1
	case 0x5e:
		c.E = c.memory[c.getHL()]
		c.PC += 1
	case 0x66:
		c.H = c.memory[c.getHL()]
		c.PC += 1
	case 0x6f:
		c.L = c.A
		c.PC += 1
	case 0x77:
		c.memory[c.getHL()] = c.A
		c.PC += 1
	case 0x7a:
		c.A = c.D
		c.PC += 1
	case 0x7b:
		c.A = c.E
		c.PC += 1
	case 0x7c:
		c.A = c.H
		c.PC += 1
	case 0x7e:
		c.A = c.memory[c.getHL()]
		c.PC += 1
	case 0xa7:
		c.A = c.A & c.A
		// TODO: write a function to set the flags for A register ops
		c.zero = c.A == 0
		c.sign = (c.A & 0x80) != 0 // if the msb is NOT zero, then negative
		c.parity = (bits.OnesCount8(c.A) % 2) == 0
		c.carry = false
		c.auxCarry = false
		c.PC += 1
	case 0xaf:
		c.A = c.A ^ c.A
		c.zero = c.A == 0
		c.sign = (c.A & 0x80) != 0 // if the msb is NOT zero, then negative
		c.parity = (bits.OnesCount8(c.A) % 2) == 0
		c.carry = false
		c.auxCarry = false
		c.PC += 1
	case 0xc1:
		// POP B
		c.C = c.memory[c.SP]
		c.B = c.memory[c.SP+1]
		c.SP += 2
		c.PC += 1
	case 0xc2:
		// JNZ: if the zero flag is NOT set, jump to adr
		if !c.zero {
			adr := uint16(c.memory[c.PC+2])<<8 | uint16(c.memory[c.PC+1])
			c.PC = adr
		} else {
			c.PC += 3
		}
	case 0xc3:
		adr := uint16(c.memory[c.PC+2])<<8 | uint16(c.memory[c.PC+1])
		c.PC = adr
	case 0xc5:
		// PUSH B
		c.memory[c.SP-2] = c.C
		c.memory[c.SP-1] = c.B
		c.SP -= 2
		c.PC += 1
	case 0xc6:
		c.A += c.memory[c.PC+1]
		c.zero = c.A == 0
		c.sign = (c.A & 0x80) != 0 // if the msb is NOT zero, then negative
		c.parity = (bits.OnesCount8(c.A) % 2) == 0
		c.carry = false
		c.auxCarry = false
		c.PC += 2
	case 0xc9:
		//RET
		adr := uint16(c.memory[c.SP+1])<<8 | uint16(c.memory[c.SP])
		c.PC = adr
		c.SP += 2
	case 0xcd:
		//CALL adr
		// (SP-1)<-PC.hi;(SP-2)<-PC.lo;SP<-SP-2;PC=adr
		adr := uint16(c.memory[c.PC+2])<<8 | uint16(c.memory[c.PC+1])
		c.memory[c.SP-1] = uint8(c.PC >> 8)
		c.memory[c.SP-2] = uint8(c.PC & 0xff)
		c.SP -= 2
		c.PC = adr
	case 0xd1:
		//POP D
		//E <- (sp); D <- (sp+1); sp <- sp+2
		c.E = c.memory[c.SP]
		c.D = c.memory[c.SP+1]
		c.SP += 2
		c.PC += 1
	case 0xd3:
		// OUT - content of A is placed on the bus to be transmitted to the specified port.
		port := c.memory[c.PC+1]
		c.redirectPort(c.A, port)
		c.PC += 2
	case 0xd5:
		// PUSH D
		c.memory[c.SP-2] = c.E
		c.memory[c.SP-1] = c.D
		c.SP -= 2
		c.PC += 1
	case 0xe1:
		// POP H
		c.L = c.memory[c.SP]
		c.H = c.memory[c.SP+1]
		c.SP += 2
		c.PC += 1
	case 0xe5:
		// PUSH H
		c.memory[c.SP-2] = c.L
		c.memory[c.SP-1] = c.H
		c.SP -= 2
		c.PC += 1
	case 0xe6:
		// A & d8
		c.A &= c.memory[c.PC+1]
		c.zero = c.A == 0
		c.sign = (c.A & 0x80) != 0 // if the msb is NOT zero, then negative
		c.parity = (bits.OnesCount8(c.A) % 2) == 0
		c.carry = false
		c.auxCarry = false
		c.PC += 2
	case 0xeb:
		c.H, c.D = c.D, c.H
		c.L, c.E = c.E, c.L
		c.PC += 1
	case 0xf1:
		//POP PSW
		c.PSW = c.memory[c.SP]
		c.A = c.memory[c.SP+1]
		// adjust flags TODO: check this again
		c.sign = (c.PSW & 0x01) == 0x01
		c.zero = (c.PSW & 0x02) == 0x02
		c.auxCarry = (c.PSW & 0x08) == 0x08
		c.parity = (c.PSW & 0x20) == 0x20
		c.carry = (c.PSW & 0x80) == 0x80
		c.SP += 2
		c.PC += 1
	case 0xf5:
		//PUSH PSW
		c.memory[c.SP-2] = c.PSW
		c.memory[c.SP-1] = c.A
		c.SP -= 2
		c.PC += 1
	case 0xfb:
		c.interrupt = true
		c.PC += 1

	case 0xfe:
		c.A -= c.memory[c.PC+1]
		c.zero = c.A == 0
		c.sign = (c.A & 0x80) != 0 // if the msb is NOT zero, then negative
		c.parity = (bits.OnesCount8(c.A) % 2) == 0
		c.carry = false
		c.auxCarry = false

		c.PC += 2

	default:
		c.unimplementedInstruction()
	}

}

func (c *CPU) RunCycles(cycles int) {
	for range cycles {
		c.Run()
	}
}

/*START OF BIT FUNCTIONS*/
// refer to here
// https://cs.opensource.google/go/go/+/refs/tags/go1.25.0:src/math/bits/bits.go
func Add16(x, y uint16) (sum, carry uint16) {
	sum32 := uint32(x) + uint32(y)
	sum = uint16(sum32)
	carry = uint16(sum32 >> 16)
	return
}

/*ENDOF BIT FUNCTIONS*/

/*START OF CPU HELPER FUNCTIONS*/

func (c *CPU) unimplementedInstruction() {}
func (c *CPU) setBC(val uint16) {
	c.B = uint8(val >> 8)
	c.C = uint8(val & 0xff)
}
func (c *CPU) setDE(val uint16) {
	c.D = uint8(val >> 8)
	c.E = uint8(val & 0xff)
}
func (c *CPU) setHL(val uint16) {
	c.H = uint8(val >> 8)
	c.L = uint8(val & 0xff)
}

func (c *CPU) getBC() uint16 {
	return uint16((c.B << 8) | c.C)
}
func (c *CPU) getDE() uint16 {
	return uint16((c.D << 8) | c.E)
}
func (c *CPU) getHL() uint16 {
	return uint16((c.H << 8) | c.L)
}

func (c *CPU) redirectPort(data, port uint8) {
	c.portMap[port] = data
}

// TODO: make flag helper functions
// func (c *CPU) stateFlagsA() {}
/*END OF CPU HELPER FUNCTIONS*/

func (c *CPU) printState() string { return "lol" }
func TestCpu(c *CPU) uint8        { return c.A }
func TestString() string          { return "hi hello" }
