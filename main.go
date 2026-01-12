package main

import (
	"bufio"
	"emulator8080/cpu"
	"fmt"
	"github.com/AllenDang/giu"
	"os"
)

func main() {
	fmt.Println("Hi " + cpu.TestString())

	// memory := make([]uint8, 0xffff)
	var memory [0xffff]uint8
	// read the ROM into the cpu memory
	if len(os.Args) > 1 {
		// then read the file
		filename := os.Args[1]
		file, err := os.Open(filename)
		if err != nil {
			fmt.Println("error reading file: " + filename)
			return
		}

		defer file.Close()

		bufr := bufio.NewReader(file)
		n, err := bufr.Read(memory[:])

		if err != nil {
			fmt.Println("error reading bytes of: " + filename)
			return
		}

		fmt.Printf("Read %d bytes into memory", n)
	}
	c := cpu.New(memory)

	wnd := giu.NewMasterWindow("Table sorting", 670, 380, 0)
	giu.Context.FontAtlas.SetDefaultFont(cpu.DefaultFont, 18)
	wnd.Run(func() { cpu.DebuggerLoop(c) })

}
