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
	c := cpu.New(memory)
	g := screen.NewGame(c)
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}

}
