package screen

import (
	"emulator8080/cpu"
	"fmt"
	"image"
	"math"
	"image/color"

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
	breakpoint   string
	memoryAdr    string
	// running      bool

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
		// running:      false,
		gameImage: *image.NewGray(image.Rect(0, 0, gameWidth, gameHeight)),
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
	vram := g.c.GetMemory()[0x2400:0x4000]
	// g.gameImage
	// TODO Modify this so that for each byte, loop 8 times, then update the
	// x and y coordinates accordingly, using some modulo
	for i, b := range vram {
		for k := 0; k < 8; k++ {
			index := i*8 + k
			xPos := index % gameWidth
			yPos := index / gameWidth
			bit := (b >> k) & 0x01

			if bit != 0 {
				g.gameImage.SetGray(xPos, yPos, color.Gray{Y: 255})
			} else {
				g.gameImage.SetGray(xPos, yPos, color.Gray{Y: 25})
			}

		}
	}
	/*
		bitIndex := 0
		for y := 0; y < gameHeight; y++ {
			for x := 0; x < gameWidth; x++ {
				if bitIndex < (0x1c00)*8 { // check in bounds
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
	*/
}

func (g *Game) Update() error {
	// update player input
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown){
		// insert credit
		g.c.Ports[1] |= 0b1
	} else {
		g.c.Ports[1] &= ^uint8(0b1)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft){
		// move left
		g.c.Ports[1] |= 0b00100000
	} else {
		g.c.Ports[1] &= ^uint8(0b00100000)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight){
		// move right
		g.c.Ports[1] |= 0b01000000
	} else {
		g.c.Ports[1] &= ^uint8(0b01000000)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp){
		// shoot & 1P start
		g.c.Ports[1] |= 0b00010100
	} else {
		g.c.Ports[1] &= ^uint8(0b00010100)
	}
	
	// update cpu cycles
	// draw debug screen
	if g.c.Running {
		// if running, run for 1/60 s cycles
		// for 2.1 MHz, 1/60 s is 35,000 cycles
		g.c.RunCycles(17500)
		// half screen interrupt
		g.c.CallRST(0x0008)
		g.c.RunCycles(17500)
		g.c.CallRST(0x0010)
	}

	// update the screen
	g.genImage()

	_, err := g.debugUI.Update(func(ctx *debugui.Context) error {
		g.cpuWindow(ctx)
		g.logWindow(ctx)
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
	op.GeoM.Scale(2, 1.5)
	op.GeoM.Rotate(math.Pi* 3/2)
	op.GeoM.Translate(100,480)

	screen.DrawImage(ebiten.NewImageFromImage(&g.gameImage), op)

	// draw debug screen
	g.debugUI.Draw(screen)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f\nFPS: %0.2f", ebiten.ActualTPS(), ebiten.ActualFPS()))
}
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
