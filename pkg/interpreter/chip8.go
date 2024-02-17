package interpreter

import (
	"fmt"
	"log/slog"
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

	TIMER_TICK = 000 / 60

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
	logger        *slog.Logger
}

func NewChip8(log *slog.Logger) *Chip8 {
	c := new(Chip8)

	c.stack = [32]uint16{}
	c.stackFrame = -1
	c.indexRegister = 0x0
	c.registers = [16]byte{}
	c.pc = uint16(MEMORY_OFFSET)
	c.memory = [4096]byte{}
	c.delayTimer = 0x0
	c.soundTimer = 0x0
	c.logger = log

	return c
}

func (c *Chip8) startDelayTimer() {
	for {
		time.Sleep(time.Millisecond * time.Duration(TIMER_TICK))

		if c.delayTimer > 0 {
			c.delayTimer--
		}
	}
}

func (c *Chip8) startSoundTimer(r *engine.Runtime) {
	for {
		time.Sleep(time.Millisecond * time.Duration(TIMER_TICK))

		if c.soundTimer > 0 {
			c.soundTimer--

			r.PlayAudio()
		} else {
			r.StopAudio()
		}

	}
}

func (c *Chip8) Interpret(r *engine.Runtime, programData []byte) {
	go c.startDelayTimer()
	go c.startSoundTimer(r)

	// Verifies if program size is greater than chip memory
	if len := len(programData); len > CHIP_MEMORY {
		c.logger.Error("Given program is larger than memory", "program_data", len, "chip_memory", CHIP_MEMORY)
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

		c.logger.Debug(
			"Instruction decoded",
			"instruction", fmt.Sprintf("%02x", instr),
			"X", fmt.Sprintf("%02x", X),
			"Y", fmt.Sprintf("%02x", Y),
			"N", fmt.Sprintf("%02x", N),
			"NN", fmt.Sprintf("%02x", NN),
			"NNN", fmt.Sprintf("%02x", NNN),
		)

		switch instr {
		case 0x00:
			switch Y {
			case 0x0E:
				switch N {
				case 0x0: // clear screen
					r.ClearScreen()

					c.logger.Debug("Clear Screen Instruction", "INSTR", fmt.Sprintf("%02x", instr))
				case 0xE:
					c.pc = c.stack[c.stackFrame]
					c.stackFrame--

					c.logger.Debug("Set stack pointer to the top Instruction")
				default:
					c.logger.Warn("Unknown instruction", "INSTR", fmt.Sprintf("%02x", instr), "Y", fmt.Sprintf("%02x", Y))
				}
			}
		case 0x1:
			c.pc = NNN

			c.logger.Debug("Jump to NNN Instruction. Set Program counter", "NNN", fmt.Sprintf("%02x", NNN), "INSTR", fmt.Sprintf("%02x", instr))
		case 0x2:
			c.stackFrame++
			c.stack[c.stackFrame] = c.pc
			c.pc = NNN

			c.logger.Debug("Increment stack pointer", "INSTR", fmt.Sprintf("%02x", instr))
			c.logger.Debug("CALL subroutine at NNN", "NNN", fmt.Sprintf("%02x", NNN), "INSTR", fmt.Sprintf("%02x", instr))
			c.logger.Debug("Put PC at top of the stack", "PC", fmt.Sprintf("%02x", NNN), "INSTR", fmt.Sprintf("%02x", instr))
			c.logger.Debug("Set PC to NNN", "PC", fmt.Sprintf("%02x", c.pc), "NNN", fmt.Sprintf("%02x", NNN), "INSTR", fmt.Sprintf("%02x", instr))
		case 0x3:
			VX := c.registers[X]
			c.logger.Debug("Skip next instruction if Vx = kk (NN)", "VX", fmt.Sprintf("%02x", VX), "NN", fmt.Sprintf("%02x", NN), "INSTR", fmt.Sprintf("%02x", instr))

			if VX == NN {
				c.pc += 2
				c.logger.Debug("Skiping next instruction, VX == KK", "VX", fmt.Sprintf("%02x", VX), "NN", fmt.Sprintf("%02x", NN), "INSTR", fmt.Sprintf("%02x", instr))
			}
		case 0x4:
			VX := c.registers[X]
			c.logger.Debug("Skip next instruction if Vx != kk (NN)", "VX", fmt.Sprintf("%02x", VX), "NN", fmt.Sprintf("%02x", NN), "INSTR", fmt.Sprintf("%02x", instr))

			if VX != NN {
				c.pc += 2
				c.logger.Debug("Skiping next instruction, VX != KK", "VX", fmt.Sprintf("%02x", VX), "NN", fmt.Sprintf("%02x", NN), "INSTR", fmt.Sprintf("%02x", instr))
			}
		case 0x5:
			VX := c.registers[X]
			VY := c.registers[Y]

			c.logger.Debug("Skip next instruction if Vx = Vy", "VX", fmt.Sprintf("%02x", VX), "VY", fmt.Sprintf("%02x", VY), "INSTR", fmt.Sprintf("%02x", instr))

			if N == 0x0 && VX == VY {
				c.logger.Debug("Skiping next instruction Vx != Vy", "VX", fmt.Sprintf("%02x", VX), "VY", fmt.Sprintf("%02x", VY), "INSTR", fmt.Sprintf("%02x", instr))
				c.pc += 2
			}
		case 0x6:
			VX := c.registers[X]
			c.logger.Debug("SET Vx = KK", "VX", fmt.Sprintf("%02x", VX), "KK (NN)", fmt.Sprintf("%02x", NN), "INSTR", fmt.Sprintf("%02x", instr))

			c.registers[X] = NN
		case 0x7:
			c.logger.Debug("Set Vx = Vx + KK", "VX", fmt.Sprintf("%02x", c.registers[X]), "KK (NN)", fmt.Sprintf("%02x", NN), "INSTR", fmt.Sprintf("%02x", instr))

			c.registers[X] = NN + c.registers[X]
		case 0x8:
			switch N {
			case 0x0:
				c.logger.Debug("Set Vx = Vy", "VX", fmt.Sprintf("%02x", c.registers[X]), "VY", fmt.Sprintf("%02x", c.registers[Y]), "INSTR", fmt.Sprintf("%02x", instr))

				c.registers[X] = c.registers[Y]
			case 0x1:
				c.logger.Debug("Set Vx = Vx OR Vy", "VX", fmt.Sprintf("%02x", c.registers[X]), "VY", fmt.Sprintf("%02x", c.registers[Y]), "INSTR", fmt.Sprintf("%02x", instr))
				c.registers[X] = c.registers[X] | c.registers[Y]
			case 0x2:
				c.logger.Debug("Set Vx = Vx AND Vy", "VX", fmt.Sprintf("%02x", c.registers[X]), "VY", fmt.Sprintf("%02x", c.registers[Y]), "INSTR", fmt.Sprintf("%02x", instr))

				c.registers[X] = c.registers[X] & c.registers[Y]
			case 0x3:
				c.logger.Debug("Set Vx = Vx XOR Vy", "VX", fmt.Sprintf("%02x", c.registers[X]), "VY", fmt.Sprintf("%02x", c.registers[Y]), "INSTR", fmt.Sprintf("%02x", instr))

				c.registers[X] = c.registers[X] ^ c.registers[Y]
			case 0x4:
				c.logger.Debug("Set Vx = Vx + Vy, set VF = carry", "VX", fmt.Sprintf("%02x", c.registers[X]), "VY", fmt.Sprintf("%02x", c.registers[Y]), "VF", fmt.Sprintf("%02x", c.registers[0xF]), "INSTR", fmt.Sprintf("%02x", instr))

				sum := uint16(c.registers[X]) + uint16(c.registers[Y])

				if int(sum) > 255 {
					c.registers[0xF] = 0x1
				} else {
					c.registers[0xF] = 0x0
				}

				c.registers[X] = byte(sum)
			case 0x5:
				c.logger.Debug("Set Vx = Vx - Vy, set VF = carry", "VX", fmt.Sprintf("%02x", c.registers[X]), "VY", fmt.Sprintf("%02x", c.registers[Y]), "VF", fmt.Sprintf("%02x", c.registers[0xF]), "INSTR", fmt.Sprintf("%02x", instr))

				if c.registers[X] > c.registers[Y] {
					c.registers[0xF] = 0x1
				} else {
					c.registers[0xF] = 0x0
				}

				c.registers[X] = c.registers[X] - c.registers[Y]
			case 0x6:
				c.logger.Debug("Set Vx = Vx SHR 1", "VX", fmt.Sprintf("%02x", c.registers[X]), "INSTR", fmt.Sprintf("%02x", instr))
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
			case 0x7:
				c.logger.Debug("Set Vx = Vy - Vx, set VF = NOT borrow", "VX", fmt.Sprintf("%02x", c.registers[X]), "VY", fmt.Sprintf("%02x", c.registers[Y]), "VF", fmt.Sprintf("%02x", c.registers[0xF]), "INSTR", fmt.Sprintf("%02x", instr))

				if c.registers[Y] > c.registers[X] {
					c.registers[0xF] = 0x1
				} else {
					c.registers[0xF] = 0x0
				}

				result := c.registers[X] - c.registers[Y]

				c.registers[X] = result
			case 0xE:
				c.logger.Debug("Set Vx = Vx SHL 1", "VX", fmt.Sprintf("%02x", c.registers[X]), "INSTR", fmt.Sprintf("%02x", instr))

				c.registers[X] = c.registers[Y]

				// check if leftmost bit is set (and shifted out)
				if c.registers[X]&(1<<7) > 0 {
					c.registers[0xF] = 0x1
				} else {
					c.registers[0xF] = 0x0
				}
				c.registers[X] = c.registers[X] << 1
			}
		case 0x9:
			c.logger.Debug("Skip next instruction if Vx != Vy", "VX", fmt.Sprintf("%02x", c.registers[X]), "VY", fmt.Sprintf("%02x", c.registers[Y]), "INSTR", fmt.Sprintf("%02x", instr))

			if c.registers[X] != c.registers[Y] {
				c.pc += 2 // SKIP INSTRUCTION (wrap on function)
			}
		case 0xA:
			c.logger.Debug("Set I = nnn", "I", fmt.Sprintf("%02x", c.indexRegister), "NNN", fmt.Sprintf("%02x", NNN), "INSTR", fmt.Sprintf("%02x", instr))

			c.indexRegister = NNN
		case 0xB:
			c.logger.Debug("Jump to location nnn + V0", "V0", fmt.Sprintf("%02x", c.registers[0x0]), "NNN", fmt.Sprintf("%02x", NNN), "INSTR", fmt.Sprintf("%02x", instr))

			c.pc = NNN + uint16(c.registers[0x0])
		case 0xC:
			rand := rand.Intn(256)

			c.registers[X] = byte(rand) & NN

			c.logger.Debug("Set Vx = random byte AND kk", "RANDOM BYTE", fmt.Sprintf("%02x", rand), "VX", fmt.Sprintf("%02x", c.registers[X]), "KK (NN)", fmt.Sprintf("%02x", NN), "INSTR", fmt.Sprintf("%02x", instr))
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
			case 0x65:
				fmt.Printf("Read registers V0 through Vx from memory starting at location I.")

				for i := uint8(0); i <= X; i++ {
					c.registers[i] = c.memory[c.indexRegister]
					c.indexRegister = c.indexRegister + 1
				}
			}
		default:
			c.logger.Warn("Unknown instruction", "INSTR", fmt.Sprintf("%02x", instr), "Y", fmt.Sprintf("%02x", Y))
		}

		time.Sleep(time.Microsecond * 1300) // corresponds to about 700 instructions per second...
	}
}
