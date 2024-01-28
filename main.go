package main

import (
	"fmt"
	"os"
)

var (
	// 4kb internal memory
	CHIP_MEMORY = 4096

	// The first CHIP-8 interpreter (on the COSMAC VIP computer) was also located in RAM,
	// from address 000 to 1FF. It would expect a CHIP-8 program to be loaded into memory
	// after it, starting at address 200 (512 in decimal).
	MEMORY_OFFSET = 0x200

	// For some reason, it’s become popular to put it at 050–09F,
	// so you can follow that convention if you want.
	FONT_OFFSET = 0x50

	// set of fonts
	FONT_SET = []uint8{
		0xF0, 0x90, 0x90, 0x90, 0xF0, //0
		0x20, 0x60, 0x20, 0x20, 0x70, //1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, //2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, //3
		0x90, 0x90, 0xF0, 0x10, 0x10, //4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, //5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, //6
		0xF0, 0x10, 0x20, 0x40, 0x40, //7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, //8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, //9
		0xF0, 0x90, 0xF0, 0x90, 0x90, //A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, //B
		0xF0, 0x80, 0x80, 0x80, 0xF0, //C
		0xE0, 0x90, 0x90, 0x90, 0xE0, //D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, //E
		0xF0, 0x80, 0xF0, 0x80, 0x80, //F
	}
)

type Instruction struct {
	INSTR [2]byte
	X     [2]byte
	Y     [2]byte
	N     [2]byte
	NN    [2]byte // KK
	NNN   [2]byte
}

func main() {
	var stack [32]uint16   // The stack offers a max depth of 32 with 2 bytes per stack frame
	var stackFrame int     // current stack frame. Starts at -1 and is set to 0 on first use
	var I uint16           // represents Index register
	var registers [16]byte // represents the 16 1-byte registers
	var pc uint16          // Program counter, set it to the initial memory offset
	var memory [4096]byte  //4kb internal memory

	stackFrame = -1

	programData, err := os.ReadFile("IBMLogo.ch8")
	if err != nil {
		fmt.Print(err)
	}

	// for _, b := range programData {
	// 	fmt.Printf("%02x ", b)
	// }

	// Verifies if program size is greater than chip memory
	if len(programData) > CHIP_MEMORY {
		fmt.Print("Given program is larger than memory")
	}

	for i := range FONT_SET {
		memory[FONT_OFFSET+i] = FONT_SET[i]
	}

	for i := range programData {
		memory[MEMORY_OFFSET+i] = programData[i]
	}

	// fmt.Printf("First instruction: %02x", instructions[0].INSTR)

	pc = uint16(MEMORY_OFFSET)
	fmt.Print(pc)
	for i := 0; i < len(programData); i++ {
		// SET program counter to the first byte of the software

		b0 := memory[pc]
		b1 := memory[pc+1]
		pc += 2

		// DECODE

		instr := (b0 & 0xF0) >> 4 // first nibble, the instruction
		X := b0 & 0x0F            // second nibble, register lookup!

		Y := (b1 & 0xF0) >> 4            // third nibble, register lookup!
		N := b1 & 0x0F                   // fourth nibble, 4 bit number
		NN := b1                         // NN = second byte
		NNN := uint16(X)<<8 | uint16(NN) // NNN = second, third and fourth nibbles

		fmt.Printf("First instruction: %02x", instr)

		fmt.Printf("X: %02x\n", X)
		fmt.Printf("Y: %02x\n", Y)
		fmt.Printf("N: %02x\n", N)
		fmt.Printf("NN: %02x\n", NN)
		fmt.Printf("NNN: %02x\n", NNN)

		switch instr {
		case 0x00:
			switch Y {
			case 0x0E:
				fmt.Print("clear scren\n")
			case 0xEE:
				pc = stack[stackFrame]
				stackFrame--
				fmt.Print("set stack pointer to the top\n")
			default:
				fmt.Printf("Unknown instruction. INSTR: %02x Y: %02x\n", instr, Y)
			}
		case 0x1:
			pc = NNN
			fmt.Printf("JUMP to NNN: %02x\n", NNN)
			fmt.Printf("SET PROGRAM COUNTER to NNN: %02x\n", NNN)
		case 0x2:
			stackFrame++
			stack[stackFrame] = pc
			pc = NNN
			fmt.Printf("CALL subroutine at NNN: %02x\n", NNN)
			fmt.Printf("Increment stack pointer\n")
			fmt.Printf("Put PC at top of the stack. PC: %d\n", pc)
			fmt.Printf("SET PC to NNN. PC: %d, NNN: %02x\n", pc, NNN)
		case 0x3:
			if registers[X] == NN {
				pc += 2
			}

			fmt.Printf("Skip next instruction if Vx = kk.\n")
		case 0x4:
			if registers[X] != NN {
				pc += 2
			}
			fmt.Printf("Skip next instruction if Vx != kk.\n")
		case 0x5:
			if registers[X] == registers[Y] {
				pc += 2
			}
			fmt.Printf("Skip next instruction if Vx = Vy.\n")
		case 0x6:
			registers[X] = NN

			fmt.Printf("Set Vx = KK\n")
		case 0x7:
			registers[X] = NN + registers[X]
			fmt.Printf("Set Vx = Vx + KK\n")
		case 0x8:
			switch N {
			case 0x0:
				registers[X] = registers[Y]
				fmt.Printf("Set Vx = Vy\n")
			case 0x1:
				registers[X] = registers[X] | registers[Y]
				fmt.Printf("Set Vx = Vx OR Vy\n")
			case 0x2:
				registers[X] = registers[X] & registers[Y]
				fmt.Printf("Set Vx = Vx AND Vy\n")
			case 0x3:
				registers[X] = registers[X] ^ registers[Y]
				fmt.Printf("Set Vx = Vx XOR Vy\n")
			case 0x4:
				fmt.Printf("Set Vx = Vx + Vy, set VF = carry\n")
			case 0x5:
				fmt.Printf("Set Vx = Vx - Vy, set VF = carry\n")
			case 0x6:
				fmt.Printf("Set Vx = Vx SHR 1.\n")
			case 0x7:
				fmt.Printf("Set Vx = Vy - Vx, set VF = NOT borrow.\n")
			case 0xE:
				fmt.Print("Set Vx = Vx SHL 1.\n")
			}
		case 0x9:
			if registers[X] != registers[Y] {
				pc += 2 // SKIP INSTRUCTION (wrap on function)
			}
			fmt.Printf("Skip next instruction if Vx != Vy.\n")
		case 0xA:
			I = NNN
			fmt.Printf("Set I = nnn.")
		case 0xB:
			fmt.Printf("Jump to location nnn + V0.\n")
		case 0xC:
			fmt.Printf("Set Vx = random byte AND kk.\n")
		case 0xD:
			fmt.Printf("Draw instruction\n") //TODO add more cases here
		case 0xE:
			switch NN {
			case 0x9E:
				fmt.Printf("Skip next instruction if key with the value of Vx is pressed.\n")
			case 0xA1:
				fmt.Printf("Skip next instruction if key with the value of Vx is not pressed.\n")
			default:
				fmt.Printf("Unknown Instruction\n")
			}
		case 0xF:
			fmt.Printf("Timer instruction.\n")
		default:
			fmt.Printf("Unknown Instruction.\n")
		}
	}
}
