package screen


import (
	"fmt"
	"image"
	// "image/color"
	"github.com/ebitengine/debugui"
	// "github.com/hajimehoshi/ebiten/v2"
	// "github.com/hajimehoshi/ebiten/v2/text/v2"
	// "github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) cpuWindow(ctx *debugui.Context) {
	ctx.Window("CPU State", image.Rect(40, 40, 600, 470), func(layout debugui.ContainerLayout) {
		ctx.SetGridLayout(nil, []int{25, -1})
		ctx.GridCell(func(bounds image.Rectangle) {
			// TODO buttons
			ctx.SetGridLayout([]int{50, 50, 50, 150, -1}, nil)
			ctx.Button("Run 1").On(func() {g.c.Run()})
			ctx.Button("Run").On(func() {g.running = true})
			ctx.Button("Pause").On(func() {g.running = false})
			ctx.Text(fmt.Sprintf("Running: %t", g.running))
			ctx.Text(fmt.Sprintf("Cycle # %v", g.c.Cycles))
		})
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.SetGridLayout([]int{100, -1}, nil)
			ctx.GridCell(func(bounds image.Rectangle) {
				ctx.SetGridLayout(nil, []int{70, -1})
				ctx.GridCell(func(bounds image.Rectangle) {
					// TODO Flags
					ctx.Text("Flags")
					ctx.SetGridLayout([]int{-1, -1, -1, -1, -1}, nil)
					ctx.Text("S")
					ctx.Text("Z")
					ctx.Text("AC")
					ctx.Text("P")
					ctx.Text("C")
					
					flags := g.c.GetFlags()
					oneZero := map[bool]int{true: 1, false: 0}
					ctx.Text(fmt.Sprintf("%d", oneZero[flags[0]]))
					ctx.Text(fmt.Sprintf("%d", oneZero[flags[1]]))
					ctx.Text(fmt.Sprintf("%d", oneZero[flags[2]]))
					ctx.Text(fmt.Sprintf("%d", oneZero[flags[3]]))
					ctx.Text(fmt.Sprintf("%d", oneZero[flags[4]]))
				})
				ctx.GridCell(func(bounds image.Rectangle) {
					// TODO Registers
					regs, regs2 := g.c.GetRegs()
					ctx.Text("Registers")
					ctx.SetGridLayout([]int{-1, -2}, nil)
					ctx.Text("A")
					ctx.Text(fmt.Sprintf("  %02X", regs[0]))
					ctx.Text("BC")
					ctx.Text(fmt.Sprintf("%02X%02X", regs[1], regs[2]))
					ctx.Text("DE")
					ctx.Text(fmt.Sprintf("%02X%02X", regs[3], regs[4]))
					ctx.Text("HL")
					ctx.Text(fmt.Sprintf("%02X%02X", regs[5], regs[6]))
					ctx.Text("SP")
					ctx.Text(fmt.Sprintf("%04X", regs2[0]))
					ctx.Text("PC")
					ctx.Text(fmt.Sprintf("%04X", regs2[1]))
				})
			})
			ctx.GridCell(func(bounds image.Rectangle) {
				// TODO Hex table
				memory := g.c.GetMemory()
				pc := g.c.PC
				memStart := int(pc & 0xfff0)
				ctx.SetGridLayout([]int{-2, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, nil)
				ctx.Text("")
				ctx.Loop(16, func(i int) {ctx.Text(fmt.Sprintf("%X", i))})
				ctx.Loop(16, func(i int) {
					// memory headers
					ctx.Text(fmt.Sprintf("0x%04X", i*0x10 + memStart))
					ctx.Loop(16, func(j int) {
						ctx.Text(fmt.Sprintf("%02X", memory[memStart + i*0x10 + j]))
					})
				})
			})
		})
	})
}
