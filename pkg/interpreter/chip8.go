package interpreter

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/otaviohenrique/zamorak/pkg/engine"
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

type Chip8Instr struct {
	INSTR [2]byte
	X     [2]byte
	Y     [2]byte
	N     [2]byte
	NN    [2]byte // KK
	NNN   [2]byte
}

type Chip8 struct {
	stack         [32]uint16 // The stack offers a max depth of 32 with 2 bytes per stack frame
	stackFrame    int        // current stack frame. Starts at -1 and is set to 0 on first use
	indexRegister uint16     // represents Index register aka I
	registers     [16]byte   // represents the 16 1-byte registers
	pc            uint16     // Program counter, set it to the initial memory offset
	memory        [4096]byte //4kb internal memory
	delayTimer    byte
	soundTimer    byte
}

func NewChip8() *Chip8 {
	c := new(Chip8)

	c.stack = [32]uint16{}
	c.stackFrame = -1
	c.indexRegister = 0x0
	c.registers = [16]byte{}
	c.pc = uint16(MEMORY_OFFSET)
	c.memory = [4096]byte{}
	c.delayTimer = 0x0
	c.soundTimer = 0x0

	return c
}

func (c *Chip8) startDelayTimer() {
	var tick = 1000 / 60

	for {
		time.Sleep(time.Millisecond * time.Duration(tick))

		if c.delayTimer > 0 {
			c.delayTimer--
		}
	}
}

func (c *Chip8) Interpret(r *engine.Runtime, programData []byte) {
	go c.startDelayTimer()
	// Verifies if program size is greater than chip memory
	if len(programData) > CHIP_MEMORY {
		fmt.Print("Given program is larger than memory")
	}

	for i := range FONT_SET {
		c.memory[FONT_OFFSET+i] = FONT_SET[i]
	}

	for i := range programData {
		c.memory[MEMORY_OFFSET+i] = programData[i]
	}

	for {
		// SET program counter to the first byte of the software

		b0 := c.memory[c.pc]
		b1 := c.memory[c.pc+1]
		c.pc += 2

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
				c.pc = c.stack[c.stackFrame]
				c.stackFrame--
				fmt.Print("set stack pointer to the top\n")
			default:
				fmt.Printf("Unknown instruction. INSTR: %02x Y: %02x\n", instr, Y)
			}
		case 0x1:
			c.pc = NNN
			fmt.Printf("JUMP to NNN: %02x\n", NNN)
			fmt.Printf("SET PROGRAM COUNTER to NNN: %02x\n", NNN)
		case 0x2:
			c.stackFrame++
			c.stack[c.stackFrame] = c.pc
			c.pc = NNN
			fmt.Printf("CALL subroutine at NNN: %02x\n", NNN)
			fmt.Printf("Increment stack pointer\n")
			fmt.Printf("Put PC at top of the stack. PC: %d\n", c.pc)
			fmt.Printf("SET PC to NNN. PC: %d, NNN: %02x\n", c.pc, NNN)
		case 0x3:
			if c.registers[X] == NN {
				c.pc += 2
			}

			fmt.Printf("Skip next instruction if Vx = kk.\n")
		case 0x4:
			if c.registers[X] != NN {
				c.pc += 2
			}
			fmt.Printf("Skip next instruction if Vx != kk.\n")
		case 0x5:
			if c.registers[X] == c.registers[Y] {
				c.pc += 2
			}
			fmt.Printf("Skip next instruction if Vx = Vy.\n")
		case 0x6:
			c.registers[X] = NN

			fmt.Printf("Set Vx = KK\n")
		case 0x7:
			c.registers[X] = NN + c.registers[X]
			fmt.Printf("Set Vx = Vx + KK\n")
		case 0x8:
			switch N {
			case 0x0:
				c.registers[X] = c.registers[Y]
				fmt.Printf("Set Vx = Vy\n")
			case 0x1:
				c.registers[X] = c.registers[X] | c.registers[Y]
				fmt.Printf("Set Vx = Vx OR Vy\n")
			case 0x2:
				c.registers[X] = c.registers[X] & c.registers[Y]
				fmt.Printf("Set Vx = Vx AND Vy\n")
			case 0x3:
				c.registers[X] = c.registers[X] ^ c.registers[Y]
				fmt.Printf("Set Vx = Vx XOR Vy\n")
			case 0x4:
				sum := uint16(c.registers[X]) + uint16(c.registers[Y])

				if int(sum) > 255 {
					c.registers[0xF] = 0x1
				} else {
					c.registers[0xF] = 0x0
				}

				c.registers[X] = byte(sum)

				fmt.Printf("Set Vx = Vx + Vy, set VF = carry\n")
			case 0x5:

				if c.registers[X] > c.registers[Y] {
					c.registers[0xF] = 0x1
				} else {
					c.registers[0xF] = 0x0
				}
				count := uint16(uint16(c.registers[Y] - c.registers[X]))

				c.registers[X] = byte(count)

				fmt.Printf("Set Vx = Vx - Vy, set VF = carry\n")
			case 0x6:
				lastBit := c.registers[X] & 0x01

				if lastBit > 0 {
					c.registers[0xF] = 0x1
				} else {
					c.registers[0xF] = 0x0
				}
				//
				//In Go, the right shift operator (>>) is often used to perform division by powers of 2 for integers. Shifting a binary number to the right by one position is equivalent to dividing it by 2.
				//
				//Here's a simple explanation:
				//
				//Shifting a binary number to the right by 1 is the same as dividing it by 2.
				//Shifting a binary number to the right by 2 is the same as dividing it by 4.
				//Shifting a binary number to the right by n is the same as dividing it by 2^n.
				//In the context of your CHIP-8 emulator, registers[X] >>= 1 is a concise way of expressing "divide registers[X] by 2." It's a common idiom used in low-level programming, especially when dealing with bitwise operations and binary representations of numbers.
				//
				//For example, if registers[X] is a binary number like 11010010, then registers[X] >>= 1 would result in 01101001, which is the value of registers[X] divided by 2.
				//
				//Using >> for division by powers of 2 is efficient and works well when you're dealing with integers and you want to express division in terms of binary operations.
				//
				c.registers[X] = c.registers[X] >> 1

				fmt.Printf("Set Vx = Vx SHR 1.\n")
			case 0x7:
				if c.registers[Y] > c.registers[X] {
					c.registers[0xF] = 0x1
				} else {
					c.registers[0xF] = 0x0
				}

				result := c.registers[X] - c.registers[Y]

				c.registers[X] = result

				fmt.Printf("Set Vx = Vy - Vx, set VF = NOT borrow.\n")
			case 0xE:
				fmt.Print("Set Vx = Vx SHL 1.\n")
			}
		case 0x9:
			if c.registers[X] != c.registers[Y] {
				c.pc += 2 // SKIP INSTRUCTION (wrap on function)
			}
			fmt.Printf("Skip next instruction if Vx != Vy.\n")
		case 0xA:
			c.indexRegister = NNN
			fmt.Printf("Set I = nnn.")
		case 0xB:
			c.pc = NNN + uint16(c.registers[0x0])
			fmt.Printf("Jump to location nnn + V0.\n")
		case 0xC:
			rand := rand.Intn(256)

			c.registers[X] = byte(rand) & NN

			fmt.Printf("Set Vx = random byte AND kk.\n")
		case 0xD:
			xc := c.registers[X] % 64
			yc := c.registers[Y] % 32

			c.registers[0xF] = 0x0

			numLines := int(N)
			firstByteIndex := c.indexRegister

			for line := 0; line < numLines; line++ {
				// first byte index
				sprite := c.memory[firstByteIndex]

				row := int(yc) + line
				if row > 31 {
					continue
				}

				for bit := 0; bit < 8; bit++ {

					col := int(xc) + bit
					// ignore if outside of screen
					if col > 63 {
						continue
					}

					// check if bit is set, moving from left-most bit to the right
					if sprite&(1<<(7-bit)) > 0 {
						if r.IsPixelSet(col, row) {
							r.Set(col, row, false)
							// set register F to 1
							c.registers[0xF] = 0x1
						} else {
							r.Set(col, row, true)
						}
					}
				}
				firstByteIndex++
			}
		case 0xE:
			switch NN {
			case 0x9E:
				fmt.Printf("Skip next instruction if key with the value of Vx is pressed.\n")

				if ebiten.IsKeyPressed(engine.ByteToKey(c.registers[X])) {
					c.pc += 2
				}
			case 0xA1:
				fmt.Printf("Skip next instruction if key with the value of Vx is not pressed.\n")

				if !ebiten.IsKeyPressed(engine.ByteToKey(c.registers[X])) {
					c.pc += 2
				}
			default:
				fmt.Printf("Unknown Instruction\n")
			}
		case 0xF:
			fmt.Printf("Timer instruction.\n")

			switch NN {
			case 0x07:
				fmt.Printf("Set Vx = delayTimer\n")

				c.registers[X] = c.delayTimer
			case 0x0A:
				ch := make(chan byte)
				go r.WaitKeyPress(ch)

				key := <-ch

				c.registers[X] = key
			case 0x15:
				fmt.Printf("Set delayTimer = Vx\n")

				c.delayTimer = c.registers[X]
			case 0x18:
				fmt.Printf("Set soundTimer = Vx\n")

				c.soundTimer = c.registers[X]
			case 0x1E:
				fmt.Printf("Set I = I + Vx\n")

				c.indexRegister = c.indexRegister + uint16(c.registers[X])
			case 0x29:
				fmt.Printf("Set I = location of sprite for digit Vx.\n")

				b := c.registers[X] & 0x0F

				c.indexRegister = uint16(FONT_SET[b])
			case 0x33:
				fmt.Printf("Store BCD representation of Vx in memory locations I, I+1, and I+2.\n")

				//The interpreter takes the decimal value of Vx,``
				//and places the hundreds digit in memory at location in I,
				//the tens digit at location I+1, and the ones digit at location I+2.

				c.memory[c.indexRegister+0] = (c.registers[X] / 100) % 10
				c.memory[c.indexRegister+1] = (c.registers[X] / 10) % 10
				c.memory[c.indexRegister+2] = (c.registers[X] / 1) % 10
			case 0x55:
				fmt.Printf("Store registers V0 through Vx in memory starting at location I.\n")

				for i := 0; i <= int(X); i++ {
					index := c.indexRegister + uint16(i)
					c.memory[index] = c.registers[i]
				}
				c.indexRegister = c.indexRegister + uint16(X+1)
			}
		default:
			fmt.Printf("Unknown Instruction.\n")
		}

		time.Sleep(time.Microsecond * 1300) // corresponds to about 700 instructions per second...
	}
}
