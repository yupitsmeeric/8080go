package screen

import (
	// "bytes"
	// _ "embed"
	"emulator8080/cpu"
	"fmt"
	"image"
	// "math"
	// "image"
	"image/color"
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
	gameImage    image.Gray
	c            *cpu.CPU
	running      bool

	debugUI debugui.DebugUI
}

const (
	ScreenWidth  = 700
	ScreenHeight = 700
	gameWidth    = 256
	gameHeight   = 224
)

func NewGame(c *cpu.CPU) *Game {
	return &Game{screenWidth: ScreenWidth,
		screenHeight: ScreenHeight,
		c:            c,
		running:      false,
		gameImage:    *image.NewGray(image.Rect(0, 0, gameWidth, gameHeight)),
	}
}

func (g *Game) genImage() {
	// take the bytes array and return an image
	/* draw pixels from cpu memory
			from https://computerarcheology.com/Arcade/SpaceInvaders/Hardware.html
			The raster resolution is 256x224 at 60Hz. The monitor is rotated in the cabinet 90 degrees counter-clockwise.
	The screens pixels are on/off (1 bit each). 256*224/8 = 7168 (7K) bytes.
			2400-3FFF 7K Video RAM
	*/
	vram := g.c.GetMemory()[0x2400:0x3FFF]
	// g.gameImage
	// TODO Modify this so that for each byte, loop 8 times, then update the 
	// x and y coordinates accordingly, using some modulo
	bitIndex := 0
	for y := 0; y < gameHeight; y++ {
		for x := 0; x < gameWidth; x++ {
			if bitIndex < (0x1c00 -1 )*8{ // check in bounds
				byteIndex := bitIndex / 8
				bitPosition := bitIndex % 8
				bit := (vram[byteIndex] >> (7 - bitPosition)) & 1

				if bit == 1 {
					g.gameImage.SetGray(x, y, color.Gray{Y: 255})
				} else {
					g.gameImage.SetGray(x, y, color.Gray{Y: 25})
				}
			bitIndex++
			}
		}
	}
}

func (g *Game) Update() error {
	// update cpu cycles
	// draw debug screen
	if g.running {
		// if running, run for 1/60 s cycles
		// for 2.1 MHz, 1/60 s is 35,000 cycles
		g.c.RunCycles(35000)
	}

	// update the screen
	g.genImage()

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
	// draw the gameImage
	op := &ebiten.DrawImageOptions{}
	// op.GeoM.Rotate(math.Pi* 3/2)
	// op.GeoM.Scale(2, 2)
	// op.GeoM.Translate(100, 400)

	screen.DrawImage(ebiten.NewImageFromImage(&g.gameImage), op)

	// draw debug screen
	g.debugUI.Draw(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f\nFPS: %0.2f", ebiten.ActualTPS(), ebiten.ActualFPS()))
}
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
