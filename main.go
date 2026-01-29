package main

import (
	// "bufio"
	"emulator8080/cpu"
	"emulator8080/screen"
	// "fmt"
	"log"
	// "os"
	_ "embed"
	// "github.com/AllenDang/giu"
	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed invaders_compiled
var rom []byte

func main() {
	ebiten.SetWindowSize(screen.ScreenWidth, screen.ScreenHeight)
	ebiten.SetWindowTitle("8080 emulator")

	//init cpu
	
	memory := make([]byte, 0xFFFF)
	copy(memory, rom)
	// Stuff for the test rom
	// https://web.archive.org/web/20240118230906/http://www.emulator101.com/full-8080-emulation.html
	// copy(memory[0x100:], rom)
	// memory[0] = 0xc3
	// memory[1] = 0x0
	// memory[2] = 0x01
	//
	// // stack pointer thing
	// memory[0x1ad] = 0x7
	//
	// // skip to AIMM test - jump to 0x022A
	// // memory[0x1b1] = 0x2A
	// // memory[0x1b2] = 0x02
	//
	// // skip DAA test
	// memory[0x59c] = 0xc3
	// memory[0x59d] = 0xc2
	// memory[0x59e] = 0x05

	c := cpu.New(memory)
	g := screen.NewGame(c)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}

}
