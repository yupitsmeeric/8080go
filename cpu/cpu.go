package cpu

import (
	// "fmt"
	// "log"
	// "fmt"
	"math/bits"
)

/*
https://web.archive.org/web/20240118230916/http://www.emulator101.com/finishing-the-cpu-emulator.html

reference implementation:
https://github.com/hlboehm/i8080-emulator/blob/main/src/emulator/processor.c
*/

/*
TODO:
- Redo whole thing with subfunctions
  - retain the cpu structure, but make a bunch more functions as getters/setters
    /aux add/sub functions
  - just copy the c implementation

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
	Ports               [8]uint8
	shiftReg            uint16
	offset              uint8
	// inPorts             [4]uint8
	// outPorts						[7]uint8
	// portMap             map[uint8]uint8

	// Flags               uint8
	/*FLAGS*/
	sign, zero, auxCarry, parity, carry bool
	Cycles                              int
	interrupt                           bool
	// interruptVector                     uint16

	// Debug stuff
	Log         string
	LogUpdated  bool
	Running     bool
	history     []uint16
	histCounter int
	Breakpoint  uint16

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
		/*
			Port 0
			 bit 0 DIP4 (Seems to be self-test-request read at power up)
			 bit 1 Always 1
			 bit 2 Always 1
			 bit 3 Always 1
			 bit 4 Fire
			 bit 5 Left
			 bit 6 Right
			 bit 7 ? tied to demux port 7 ?

			Port 1
			 bit 0 = CREDIT (1 if deposit)
			 bit 1 = 2P start (1 if pressed)
			 bit 2 = 1P start (1 if pressed)
			 bit 3 = Always 1
			 bit 4 = 1P shot (1 if pressed)
			 bit 5 = 1P left (1 if pressed)
			 bit 6 = 1P right (1 if pressed)
			 bit 7 = Not connected

			Port 2
			 bit 0 = DIP3 00 = 3 ships  10 = 5 ships
			 bit 1 = DIP5 01 = 4 ships  11 = 6 ships
			 bit 2 = Tilt
			 bit 3 = DIP6 0 = extra ship at 1500, 1 = extra ship at 1000
			 bit 4 = P2 shot (1 if pressed)
			 bit 5 = P2 left (1 if pressed)
			 bit 6 = P2 right (1 if pressed)
			 bit 7 = DIP7 Coin info displayed in demo screen 0=ON

			Port 3
			  bit 0-7 Shift register data
		*/
		// inPorts: [4]uint8{
		// 	0b01110000,
		// 	0b00010000,
		// 	0b00000001,
		// 	0,
		// },
		// outPorts: [7]uint8{0, 0, 0, 0, 0, 0, 0},
		Ports: [8]uint8{0, 0, 0b11, 0, 0, 0, 0, 0},
		// portMap:   make(map[uint8]uint8),

		Log:         "",
		LogUpdated:  false,
		Running:     false,
		history:     make([]uint16, 10000),
		histCounter: 0,
		Breakpoint:  0xFFFF,
	}
}

/*
INTERRUPT HANDLING
*/
func (c *CPU) CallRST(vector uint16) {
	// TODO rewrite this to make the interrupts stay if needed,
	// or however else to manage the interrupt enable/disable
	if !c.interrupt {
		return
	}
	c.call(vector)
}

/****  MAIN RUN FUNCTION  ****/
func (c *CPU) Run() {
	// run 1 cycle
	c.Cycles += 1

	// update PC history
	c.history[c.histCounter] = c.PC
	c.histCounter += 1
	// check if the counter is too big
	// then reset it
	if c.histCounter >= 9000 {
		copy(c.history[:15], c.history[c.histCounter-15:c.histCounter])
		c.histCounter = 15
	}

	var op uint8
	op = c.memory[c.PC]
	c.PC += 1

	switch op {
	case 0x7F: //c.A = c.A // MOV A,A
	case 0x78:
		c.A = c.B // MOV A,B
	case 0x79:
		c.A = c.C // MOV A,C
	case 0x7A:
		c.A = c.D // MOV A,D
	case 0x7B:
		c.A = c.E // MOV A,E
	case 0x7C:
		c.A = c.H // MOV A,H
	case 0x7D:
		c.A = c.L // MOV A,L
	case 0x7E:
		c.A = c.readByte(c.getHL()) // MOV A,M

	case 0x0A:
		c.A = c.readByte(c.getBC()) // LDAX B
	case 0x1A:
		c.A = c.readByte(c.getDE()) // LDAX D
	case 0x3A:
		c.A = c.readByte(c.nextWord()) // LDA word

	case 0x47:
		c.B = c.A // MOV B,A
	case 0x40: // MOV B,B
	// case 0x40: c.B = c.B // MOV B,B
	case 0x41:
		c.B = c.C // MOV B,C
	case 0x42:
		c.B = c.D // MOV B,D
	case 0x43:
		c.B = c.E // MOV B,E
	case 0x44:
		c.B = c.H // MOV B,H
	case 0x45:
		c.B = c.L // MOV B,L
	case 0x46:
		c.B = c.readByte(c.getHL()) // MOV B,M

	case 0x4F:
		c.C = c.A // MOV C,A
	case 0x48:
		c.C = c.B // MOV C,B
	case 0x49: // MOV C,C
	// case 0x49:
	// 	c.C = c.C // MOV C,C
	case 0x4A:
		c.C = c.D // MOV C,D
	case 0x4B:
		c.C = c.E // MOV C,E
	case 0x4C:
		c.C = c.H // MOV C,H
	case 0x4D:
		c.C = c.L // MOV C,L
	case 0x4E:
		c.C = c.readByte(c.getHL()) // MOV C,M

	case 0x57:
		c.D = c.A // MOV D,A
	case 0x50:
		c.D = c.B // MOV D,B
	case 0x51:
		c.D = c.C // MOV D,C
	case 0x52: // MOV D,D
	// case 0x52:
	// 	c.D = c.D // MOV D,D
	case 0x53:
		c.D = c.E // MOV D,E
	case 0x54:
		c.D = c.H // MOV D,H
	case 0x55:
		c.D = c.L // MOV D,L
	case 0x56:
		c.D = c.readByte(c.getHL()) // MOV D,M

	case 0x5F:
		c.E = c.A // MOV E,A
	case 0x58:
		c.E = c.B // MOV E,B
	case 0x59:
		c.E = c.C // MOV E,C
	case 0x5A:
		c.E = c.D // MOV E,D
	case 0x5B: // MOV E,E
	// case 0x5B: c.E = c.E // MOV E,E
	case 0x5C:
		c.E = c.H // MOV E,H
	case 0x5D:
		c.E = c.L // MOV E,L
	case 0x5E:
		c.E = c.readByte(c.getHL()) // MOV E,M

	case 0x67:
		c.H = c.A // MOV H,A
	case 0x60:
		c.H = c.B // MOV H,B
	case 0x61:
		c.H = c.C // MOV H,C
	case 0x62:
		c.H = c.D // MOV H,D
	case 0x63:
		c.H = c.E // MOV H,E
	case 0x64: // MOV H,H
	// case 0x64: c.H = c.H // MOV H,H
	case 0x65:
		c.H = c.L // MOV H,L
	case 0x66:
		c.H = c.readByte(c.getHL()) // MOV H,M

	case 0x6F:
		c.L = c.A // MOV L,A
	case 0x68:
		c.L = c.B // MOV L,B
	case 0x69:
		c.L = c.C // MOV L,C
	case 0x6A:
		c.L = c.D // MOV L,D
	case 0x6B:
		c.L = c.E // MOV L,E
	case 0x6C:
		c.L = c.H // MOV L,H
	case 0x6D: // MOV L,L
	// case 0x6D: c.L = c.L // MOV L,L
	case 0x6E:
		c.L = c.readByte(c.getHL()) // MOV L,M

	case 0x77:
		c.writeByte(c.getHL(), c.A) // MOV M,A
	case 0x70:
		c.writeByte(c.getHL(), c.B) // MOV M,B
	case 0x71:
		c.writeByte(c.getHL(), c.C) // MOV M,C
	case 0x72:
		c.writeByte(c.getHL(), c.D) // MOV M,D
	case 0x73:
		c.writeByte(c.getHL(), c.E) // MOV M,E
	case 0x74:
		c.writeByte(c.getHL(), c.H) // MOV M,H
	case 0x75:
		c.writeByte(c.getHL(), c.L) // MOV M,L

	case 0x3E:
		c.A = c.nextByte() // MVI A,byte
	case 0x06:
		c.B = c.nextByte() // MVI B,byte
	case 0x0E:
		c.C = c.nextByte() // MVI C,byte
	case 0x16:
		c.D = c.nextByte() // MVI D,byte
	case 0x1E:
		c.E = c.nextByte() // MVI E,byte
	case 0x26:
		c.H = c.nextByte() // MVI H,byte
	case 0x2E:
		c.L = c.nextByte() // MVI L,byte
	case 0x36:
		c.writeByte(c.getHL(), c.nextByte()) // MVI M, byte

	case 0x02:
		c.writeByte(c.getBC(), c.A) // STAX B
	case 0x12:
		c.writeByte(c.getDE(), c.A) // STAX D
	case 0x32:
		c.writeByte(c.nextWord(), c.A) // STA word

	case 0x01:
		c.setBC(c.nextWord()) // LXI B,word
	case 0x11:
		c.setDE(c.nextWord()) // LXI D,word
	case 0x21:
		c.setHL(c.nextWord()) // LXI H,word
	case 0x31:
		c.SP = c.nextWord() // LXI SP,word
	case 0x2A:
		c.setHL(c.readWord(c.nextWord())) // LHLD
	case 0x22:
		c.writeWord(c.nextWord(), c.getHL()) // SHLD
	case 0xF9:
		c.SP = c.getHL() // SPHL

	case 0xEB:
		c.xchg() // XCHG
	case 0xE3:
		c.xthl() // XTHL

	case 0x87:
		c.A = c.add(c.A, c.A, false) // ADD A
	case 0x80:
		c.A = c.add(c.A, c.B, false) // ADD B
	case 0x81:
		c.A = c.add(c.A, c.C, false) // ADD C
	case 0x82:
		c.A = c.add(c.A, c.D, false) // ADD D
	case 0x83:
		c.A = c.add(c.A, c.E, false) // ADD E
	case 0x84:
		c.A = c.add(c.A, c.H, false) // ADD H
	case 0x85:
		c.A = c.add(c.A, c.L, false) // ADD L
	case 0x86:
		c.A = c.add(c.A, c.readByte(c.getHL()), false) // ADD M
	case 0xC6:
		c.A = c.add(c.A, c.nextByte(), false) // ADI byte

	case 0x8F:
		c.A = c.add(c.A, c.A, c.carry) // ADC A
	case 0x88:
		c.A = c.add(c.A, c.B, c.carry) // ADC B
	case 0x89:
		c.A = c.add(c.A, c.C, c.carry) // ADC C
	case 0x8A:
		c.A = c.add(c.A, c.D, c.carry) // ADC D
	case 0x8B:
		c.A = c.add(c.A, c.E, c.carry) // ADC E
	case 0x8C:
		c.A = c.add(c.A, c.H, c.carry) // ADC H
	case 0x8D:
		c.A = c.add(c.A, c.L, c.carry) // ADC L
	case 0x8E:
		c.A = c.add(c.A, c.readByte(c.getHL()), c.carry) // ADC M
	case 0xCE:
		c.A = c.add(c.A, c.nextByte(), c.carry) // ACI byte

	case 0x97:
		c.A = c.sub(c.A, c.A, false) // SUB A
	case 0x90:
		c.A = c.sub(c.A, c.B, false) // SUB B
	case 0x91:
		c.A = c.sub(c.A, c.C, false) // SUB C
	case 0x92:
		c.A = c.sub(c.A, c.D, false) // SUB D
	case 0x93:
		c.A = c.sub(c.A, c.E, false) // SUB E
	case 0x94:
		c.A = c.sub(c.A, c.H, false) // SUB H
	case 0x95:
		c.A = c.sub(c.A, c.L, false) // SUB L
	case 0x96:
		c.A = c.sub(c.A, c.readByte(c.getHL()), false) // SUB M
	case 0xD6:
		c.A = c.sub(c.A, c.nextByte(), false) // SUI byte

	case 0x9F:
		c.A = c.sub(c.A, c.A, c.carry) // SBB A
	case 0x98:
		c.A = c.sub(c.A, c.B, c.carry) // SBB B
	case 0x99:
		c.A = c.sub(c.A, c.C, c.carry) // SBB C
	case 0x9A:
		c.A = c.sub(c.A, c.D, c.carry) // SBB D
	case 0x9B:
		c.A = c.sub(c.A, c.E, c.carry) // SBB E
	case 0x9C:
		c.A = c.sub(c.A, c.H, c.carry) // SBB H
	case 0x9D:
		c.A = c.sub(c.A, c.L, c.carry) // SBB L
	case 0x9E:
		c.A = c.sub(c.A, c.readByte(c.getHL()), c.carry) // SBB M
	case 0xDE:
		c.A = c.sub(c.A, c.nextByte(), c.carry) // SBI byte

	case 0x09:
		c.dad(c.getBC()) // DAD B
	case 0x19:
		c.dad(c.getDE()) // DAD D
	case 0x29:
		c.dad(c.getHL()) // DAD H
	case 0x39:
		c.dad(c.SP) // DAD SP

		// TODO interrupts
	case 0xF3:
		c.interrupt = false // DI - disable interrupts
	case 0xFB:
		c.interrupt = true // EI - enable interrupts
		// TODO get a hold of the timing stuff
	case 0x00: // NOP
	case 0x76:
		c.unimplementedInstruction() //HLT

	case 0x3C:
		c.A = c.inr(c.A) // INR A
	case 0x04:
		c.B = c.inr(c.B) // INR B
	case 0x0C:
		c.C = c.inr(c.C) // INR C
	case 0x14:
		c.D = c.inr(c.D) // INR D
	case 0x1C:
		c.E = c.inr(c.E) // INR E
	case 0x24:
		c.H = c.inr(c.H) // INR H
	case 0x2C:
		c.L = c.inr(c.L) // INR L
	case 0x34:
		c.writeByte(c.getHL(), c.inr(c.readByte(c.getHL()))) // INR M

	case 0x3D:
		c.A = c.dcr(c.A) // DCR A
	case 0x05:
		c.B = c.dcr(c.B) // DCR B
	case 0x0D:
		c.C = c.dcr(c.C) // DCR C
	case 0x15:
		c.D = c.dcr(c.D) // DCR D
	case 0x1D:
		c.E = c.dcr(c.E) // DCR E
	case 0x25:
		c.H = c.dcr(c.H) // DCR H
	case 0x2D:
		c.L = c.dcr(c.L) // DCR L
	case 0x35:
		c.writeByte(c.getHL(), c.dcr(c.readByte(c.getHL()))) // INR M

	case 0x03:
		c.setBC(c.getBC() + 1) // INX B
	case 0x13:
		c.setDE(c.getDE() + 1) // INX D
	case 0x23:
		c.setHL(c.getHL() + 1) // INX H
	case 0x33:
		c.SP += 1 // INX SP

	case 0x0B:
		c.setBC(c.getBC() - 1) // DCX B
	case 0x1B:
		c.setDE(c.getDE() - 1) // DCX D
	case 0x2B:
		c.setHL(c.getHL() - 1) // DCX H
	case 0x3B:
		c.SP -= 1 // DCX SP

	case 0x27:
		c.daa() // DAA
	case 0x2F:
		c.A = ^c.A // CMA
	case 0x37:
		c.carry = true // STC
	case 0x3F:
		c.carry = !c.carry // CMC

	case 0x07:
		c.rlc() // RLC (rotate left)
	case 0x0F:
		c.rrc() // RRC (rotate right)
	case 0x17:
		c.ral() // RAL
	case 0x1F:
		c.rar() // RAR

	case 0xA7:
		c.andA(c.A) // ANA A
	case 0xA0:
		c.andA(c.B) // ANA B
	case 0xA1:
		c.andA(c.C) // ANA C
	case 0xA2:
		c.andA(c.D) // ANA D
	case 0xA3:
		c.andA(c.E) // ANA E
	case 0xA4:
		c.andA(c.H) // ANA H
	case 0xA5:
		c.andA(c.L) // ANA L
	case 0xA6:
		c.andA(c.readByte(c.getHL())) // ANA M
	case 0xE6:
		c.andA(c.nextByte()) // ANI byte

	case 0xAF:
		c.xorA(c.A) // XRA A
	case 0xA8:
		c.xorA(c.B) // XRA B
	case 0xA9:
		c.xorA(c.C) // XRA C
	case 0xAA:
		c.xorA(c.D) // XRA D
	case 0xAB:
		c.xorA(c.E) // XRA E
	case 0xAC:
		c.xorA(c.H) // XRA H
	case 0xAD:
		c.xorA(c.L) // XRA L
	case 0xAE:
		c.xorA(c.readByte(c.getHL())) // XRA M
	case 0xEE:
		c.xorA(c.nextByte()) // XRI byte

	case 0xB7:
		c.orA(c.A) // ORA A
	case 0xB0:
		c.orA(c.B) // ORA B
	case 0xB1:
		c.orA(c.C) // ORA C
	case 0xB2:
		c.orA(c.D) // ORA D
	case 0xB3:
		c.orA(c.E) // ORA E
	case 0xB4:
		c.orA(c.H) // ORA H
	case 0xB5:
		c.orA(c.L) // ORA L
	case 0xB6:
		c.orA(c.readByte(c.getHL())) // ORA M
	case 0xF6:
		c.orA(c.nextByte()) // ORI byte

	case 0xBF:
		c.cmpA(c.A) // CMP A
	case 0xB8:
		c.cmpA(c.B) // CMP B
	case 0xB9:
		c.cmpA(c.C) // CMP C
	case 0xBA:
		c.cmpA(c.D) // CMP D
	case 0xBB:
		c.cmpA(c.E) // CMP E
	case 0xBC:
		c.cmpA(c.H) // CMP H
	case 0xBD:
		c.cmpA(c.L) // CMP L
	case 0xBE:
		c.cmpA(c.readByte(c.getHL())) // CMP M
	case 0xFE:
		c.cmpA(c.nextByte()) // CPI byte

	case 0xC3:
		c.jmp(c.nextWord()) // JMP
	case 0xC2:
		c.condJmp(!c.zero) // JNZ
	case 0xCA:
		c.condJmp(c.zero) // JZ
	case 0xD2:
		c.condJmp(!c.carry) // JNC
	case 0xDA:
		c.condJmp(c.carry) // JC
	case 0xE2:
		c.condJmp(!c.parity) // JPO - parity means even
	case 0xEA:
		c.condJmp(c.parity) // JPE
	case 0xF2:
		c.condJmp(!c.sign) // JP
	case 0xFA:
		c.condJmp(c.sign) // JM

	case 0xE9:
		c.PC = c.getHL() // PCHL
	case 0xCD:
		c.call(c.nextWord())

	case 0xC4:
		c.condCall(!c.zero) // CNZ
	case 0xCC:
		c.condCall(c.zero) // CZ
	case 0xD4:
		c.condCall(!c.carry) // CNC
	case 0xDC:
		c.condCall(c.carry) // CC
	case 0xE4:
		c.condCall(!c.parity) // CPO
	case 0xEC:
		c.condCall(c.parity) // CPE
	case 0xF4:
		c.condCall(!c.sign) // CP
	case 0xFC:
		c.condCall(c.sign) // CM

	case 0xC9:
		c.ret() // RET
	case 0xC0:
		c.condRet(!c.zero) // RNZ
	case 0xC8:
		c.condRet(c.zero) // RZ
	case 0xD0:
		c.condRet(!c.carry) // RNC
	case 0xD8:
		c.condRet(c.carry) // RC
	case 0xE0:
		c.condRet(!c.parity) // RPO
	case 0xE8:
		c.condRet(c.parity) // RPE
	case 0xF0:
		c.condRet(!c.sign) // RP
	case 0xF8:
		c.condRet(c.sign) // RM

	case 0xC7:
		c.call(0x00) // RST 0
	case 0xCF:
		c.call(0x08) // RST 1
	case 0xD7:
		c.call(0x10) // RST 2
	case 0xDF:
		c.call(0x18) // RST 3
	case 0xE7:
		c.call(0x20) // RST 4
	case 0xEF:
		c.call(0x28) // RST 5
	case 0xF7:
		c.call(0x30) // RST 6
	case 0xFF:
		c.call(0x38) // RST 7

	case 0xC5:
		c.pushStack(c.getBC()) // PUSH B
	case 0xD5:
		c.pushStack(c.getDE()) // PUSH D
	case 0xE5:
		c.pushStack(c.getHL()) // PUSH H
	case 0xF5:
		c.pushPSW() // PUSH PSW
	case 0xC1:
		c.setBC(c.popStack()) // POP B
	case 0xD1:
		c.setDE(c.popStack()) // POP D
	case 0xE1:
		c.setHL(c.popStack()) // POP H
	case 0xF1:
		c.popPSW() // POP PSW

		// case 0xDB: c->a = c->port_in(c->userdata, i8080_next_byte(c)); break; // IN
		// case 0xD3: c->port_out(c->userdata, i8080_next_byte(c), c->a); break; // OUT

	case 0xDB: // IN
		// c.A = c.inPorts[c.nextByte()]
		portNo := c.nextByte()
		switch portNo {
		case 3:
			c.A = uint8(c.shiftReg >> (8 - c.offset)) // write to the correct port
		default:
			c.A = c.Ports[portNo]

		}
	case 0xD3: // OUT
		portNo := c.nextByte()
		// c.ports[portNo] = c.A
		// might be kind of redundant for the shift regs
		// since theyre being handled manually but just keep it
		switch portNo {
		case 4:
			c.shiftReg = (uint16(c.A) << 8) | (c.shiftReg >> 8)
		case 2:
			// set the offset for the 8 bit result returned from the shift register
			c.offset = c.A & 0b111
		}

		// c.unimplementedInstruction() // OUT

	case 0x08:
	case 0x10:
	case 0x18:
	case 0x20:
	case 0x28:
	case 0x30:
	case 0x38: // undocumented NOPs

	case 0xD9:
		c.ret() // undocumented RET

	case 0xDD:
		c.call(c.nextWord())
	case 0xED:
		c.call(c.nextWord())
	case 0xFD:
		c.call(c.nextWord()) // undocumented CALLs

	case 0xCB:
		c.jmp(c.nextWord()) // undocumented JMP
	default:
		c.unimplementedInstruction()
	}
	// c.PC += 1 // default increase pc by 1. for other longer instructions see individual instructions

}

func (c *CPU) RunCycles(cycles int) {
	for range cycles {
		if c.Running && !(c.PC == c.Breakpoint) {
			c.Run()
		} else {
			c.Running = false // debug stop running
			break
		}
	}
}

/*START OF BIT FUNCTIONS*/
// refer to here
// https://cs.opensource.google/go/go/+/refs/tags/go1.25.0:src/math/bits/bits.go
func add8(x, y uint8, carry bool) (sum uint8, carryOut, carryHalf bool) {
	c := uint16(0)
	if carry {
		c = 1
	}
	sum16 := uint16(x) + uint16(y) + c
	sum = uint8(sum16)
	carryOut = uint8(sum16>>8) == 1
	carryHalf = ((sum ^ x ^ y) & 0b1000) != 0
	return
}

func add16(x, y uint16, carry bool) (sum uint16, carryOut bool) {
	c := uint32(0)
	if carry {
		c = 1
	}
	sum32 := uint32(x) + uint32(y) + c
	sum = uint16(sum32)
	carryOut = uint16(sum32>>16) == 1
	// carryHalf = ((sum ^ x ^ y) & 0b1000) != 0
	return
}

func parity(val uint8) bool {
	return (bits.OnesCount8(val) % 2) == 0
}

// TODO some kind of half carry flag check
// func auxCarry(a, b uint8) bool
/*ENDOF BIT FUNCTIONS*/

/*START OF CPU HELPER FUNCTIONS*/

func (c *CPU) readByte(adr uint16) uint8 { return c.memory[adr] }
func (c *CPU) readWord(adr uint16) uint16 {
	return (uint16(c.memory[adr+1])<<8 | uint16(c.memory[adr]))
}

func (c *CPU) writeByte(adr uint16, val uint8) {
	c.memory[adr] = val
}
func (c *CPU) writeWord(adr uint16, val uint16) {
	c.writeByte(adr, uint8(val&0xFF))
	c.writeByte(adr+1, uint8(val>>8))
}

// read the byte/word AFTER the opcode
func (c *CPU) nextByte() (res uint8) {
	res = c.readByte(c.PC)
	c.PC += 1
	return

}
func (c *CPU) nextWord() (res uint16) {
	res = c.readWord(c.PC)
	c.PC += 2
	return
}

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
	return ((uint16(c.B) << 8) | uint16(c.C))
}
func (c *CPU) getDE() uint16 {
	return ((uint16(c.D) << 8) | uint16(c.E))
}
func (c *CPU) getHL() uint16 {
	return ((uint16(c.H) << 8) | uint16(c.L))
}

// stack functions
func (c *CPU) pushStack(val uint16) {
	c.SP -= 2
	c.writeWord(c.SP, val)
}

func (c *CPU) popStack() uint16 {
	c.SP += 2
	return c.readWord(c.SP - 2)
}

// Arithmetic
func (c *CPU) add(a, b uint8, carry bool) (res uint8) {
	res, carryOut, carryHalf := add8(a, b, carry)
	c.carry = carryOut
	c.auxCarry = carryHalf
	// add half carry flag
	c.setZSP(res)
	return
}

func (c *CPU) sub(a, b uint8, carry bool) (res uint8) {
	res = c.add(a, (0xFF ^ b), !carry)
	if carry {
		c.carry = b+1 > a
	} else {
		c.carry = b > a
	}

	return
}

func (c *CPU) dad(val uint16) { // add word to HL
	res, carry := add16(c.getHL(), val, false)
	c.carry = carry
	c.setHL(res)
}

func (c *CPU) inr(val uint8) (res uint8) { // increment a byte
	res = val + 1
	c.auxCarry = (res & 0xF) == 0
	c.setZSP(res)
	return
}

func (c *CPU) dcr(val uint8) (res uint8) {
	res = val - 1
	c.auxCarry = (res & 0xF) == 0
	c.setZSP(res)
	return
}

// AND on a reg, then store in A
func (c *CPU) andA(val uint8) {
	res := c.A & val
	c.carry = false
	c.auxCarry = ((c.A | val) & 0x08) != 0
	c.setZSP(res)
	c.A = res
}

// XOR on a reg then store in a
func (c *CPU) xorA(val uint8) {
	res := c.A ^ val
	c.carry = false
	c.auxCarry = false
	c.setZSP(res)
	c.A = res
}

// OR on a reg then store in a
func (c *CPU) orA(val uint8) {
	res := c.A | val
	c.carry = false
	c.auxCarry = false
	c.setZSP(res)
	c.A = res
}

func (c *CPU) cmpA(val uint8) {
	res := uint16(c.A) - uint16(val)
	c.carry = c.A < val
	// c.carry = (res >> 8) != 0
	// c.auxCarry = (^(uint16(c.A) ^ res ^ uint16(val)) & 0x10) != 0
	c.setZSP(uint8(res & 0xFF))
}

// all jmp functions modify the pc for you
func (c *CPU) jmp(adr uint16) {
	c.PC = adr
}

func (c *CPU) condJmp(cond bool) {
	adr := c.nextWord()
	if cond {
		c.PC = adr
	} //else {
	// c.PC += 3
	// }
}

func (c *CPU) call(adr uint16) {
	switch adr {
	/* FOR TEST ROM
	case 0x0005:
		// Printing function
		c.Running = false // debug stop running
		c.LogUpdated = true
		switch c.C {
		case 9:
			c.Log += "history: "
			for _, val := range c.history[c.histCounter-15 : c.histCounter] {
				c.Log += fmt.Sprintf("0x%02X ", val)
			}
			c.Log += "\n"

			strAdr := c.getDE() + 3
			for c.memory[strAdr] != 0 {
				c.Log += fmt.Sprintf("%c", c.memory[strAdr])
				strAdr += 1
			}
			c.Log += "\n"

		case 2:
			c.Log += "print char routine called\n"
		}
	case 0x068B:
		// CPUER
		c.LogUpdated = true
		c.Log += fmt.Sprintf("CPUER called from loc: %X, op: %X\n", c.PC-3, c.memory[c.PC-3:c.PC])
		c.pushStack(c.PC)
		c.jmp(adr)
	*/
	default:
		c.pushStack(c.PC)
		c.jmp(adr)
	}
}

func (c *CPU) condCall(cond bool) {
	adr := c.nextWord()
	if cond {
		c.call(adr)
	} //else {
	// c.PC += 3
	// }
}

func (c *CPU) ret() {
	c.PC = c.popStack()
}

func (c *CPU) condRet(cond bool) {
	if cond {
		c.ret()
	} //else {
	// not too sure about this, make sure that the opcodes using this
	// deal with the pc in case of fail or smth
	// or make it consistent at least
	// c.PC += 1
	// }
}

// Push A | flags onto the stack
func (c *CPU) pushPSW() {
	var psw uint16 = 0
	if c.sign {
		psw |= 1 << 7
	}
	if c.zero {
		psw |= 1 << 6
	}
	if c.auxCarry {
		psw |= 1 << 4
	}
	if c.parity {
		psw |= 1 << 2
	}
	psw |= 1 << 1
	if c.carry {
		psw |= 1 << 0
	}
	c.pushStack((uint16(c.A) << 8) | psw)
}

// opposite of above - pop the A/flags from the stack onto the reg
func (c *CPU) popPSW() {
	val := c.popStack()
	c.A = uint8(val >> 8)
	c.sign = (val & (1 << 7)) != 0
	c.zero = (val & (1 << 6)) != 0
	c.auxCarry = (val & (1 << 4)) != 0
	c.parity = (val & (1 << 2)) != 0
	c.carry = (val & (1 << 0)) != 0
}

// rotate left carry on reg A
func (c *CPU) rlc() {
	carry := c.A >> 7
	c.carry = carry != 0
	c.A = (c.A << 1) | carry
}

// rotate right carry on reg A
func (c *CPU) rrc() {
	carry := c.A & 1
	c.carry = carry != 0
	c.A = (c.A >> 1) | (carry << 7)
}

// rotate left on reg A but keep the carry from the carry flag
func (c *CPU) ral() {
	var oldCarry uint8 = 0
	if c.carry {
		oldCarry = 1
	}

	carry := c.A >> 7
	c.carry = carry != 0
	c.A = (c.A << 1) | oldCarry
}

// rotate right on reg A but keep the carry from the carry flag
func (c *CPU) rar() {
	var oldCarry uint8 = 0
	if c.carry {
		oldCarry = 1
	}

	carry := c.A & 1
	c.carry = carry != 0
	c.A = (c.A >> 1) | (oldCarry << 7)
}

// Decimal Adjust Accumulator: the eight-bit number in register A is adjusted
// to form two four-bit binary-coded-decimal digits.
// For example, if A=$2B and DAA is executed, A becomes $31.
func (c *CPU) daa() {
	// TODO actually skip this first
	c.unimplementedInstruction()
}

// swap DE and HL
func (c *CPU) xchg() {
	de := c.getDE()
	c.setDE(c.getHL())
	c.setHL(de)
}

// swap a word at (sp) and HL
func (c *CPU) xthl() {
	val := c.readWord(c.SP)
	c.writeWord(c.SP, c.getHL())
	c.setHL(val)
}

func (c *CPU) setZSP(val uint8) {
	c.zero = val == 0
	c.sign = (val >> 7) == 1
	c.parity = parity(val)
}
func (c *CPU) unimplementedInstruction() {}

// func (c *CPU) redirectPort(data, port uint8) {
// c.portMap[port] = data
// }

// TODO: make flag helper functions
// func (c *CPU) stateFlagsA() {}
/*END OF CPU HELPER FUNCTIONS*/

func (c *CPU) printState() string { return "lol" }
func TestCpu(c *CPU) uint8        { return c.A }
func TestString() string          { return "hi hello" }
