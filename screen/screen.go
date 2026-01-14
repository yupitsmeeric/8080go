package screen

import (
	// "bytes"
	// _ "embed"
	"emulator8080/cpu"
	"fmt"
	// "image"
	// "image/color"
	// _ "image/jpeg"
	// "math/rand/v2"
	// "os"
	// "strings"

	"github.com/ebitengine/debugui"
	// "github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	screenWidth  int
	screenHeight int
	c *cpu.CPU

	debugUI debugui.DebugUI
}

const (
	ScreenWidth  = 700
	ScreenHeight = 700
)

func NewGame(c *cpu.CPU) *Game {
	return &Game{screenWidth: ScreenWidth, screenHeight: ScreenHeight, c: c}
}
func (g *Game) Update() error {
	// update cpu cycles
	// draw debug screen
	_, err := g.debugUI.Update(func(ctx *debugui.Context) error {
		g.cpuWindow(ctx)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
func (g *Game) Draw(screen *ebiten.Image) {
	// draw pixels from cpu memory

	// draw debug screen
	g.debugUI.Draw(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f\nFPS: %0.2f", ebiten.ActualTPS(), ebiten.ActualFPS()))
}
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
