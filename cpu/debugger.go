package cpu

/*
TODO:
	- Add PSW register
	- Highlight the PC cell
	- Flags table on the top row
	- Add IO port? maybe not
*/
import (
	"github.com/AllenDang/giu"
	"golang.org/x/image/colornames"
)

const DefaultFont = "IosevkaTermSlabNerdFontMono-Medium.ttf"

var (
	// registers = []string{"A", "BC", "DE", "HL", "PC", "SP"}
	rows   = []*giu.TableRowWidget{}
	regs   = [7]uint8{5, 55, 255, 127, 99, 88, 90}
	regs16 = [2]uint16{0xffff, 0x1010}
	memory = [0xffff]uint8{}
)

func rebuildRows(c *CPU) {
	// Rebuild register rows
	/*
	   A   0  x
	   BC  1  2
	   DE  3  4
	   HL  5  6
	   PC  0  x
	   SP  1  x
	*/
	rows = make([]*giu.TableRowWidget, 0)
	rows = append(rows, giu.TableRow(
		giu.Label("A"),
		giu.Labelf("%02x", c.A),
		giu.Labelf("%02x", c.PSW),
	))
	rows = append(rows, giu.TableRow(
		giu.Label("BC"),
		giu.Labelf("%02x", c.B),
		giu.Labelf("%02x", c.C),
	))
	rows = append(rows, giu.TableRow(
		giu.Label("DE"),
		giu.Labelf("%02x", c.D),
		giu.Labelf("%02x", c.E),
	))
	rows = append(rows, giu.TableRow(
		giu.Label("HL"),
		giu.Labelf("%02x", c.H),
		giu.Labelf("%02x", c.L),
	))
	rows = append(rows, giu.TableRow(
		giu.Label("PC"),
		giu.Labelf("%02x", c.PC>>8),
		giu.Labelf("%02x", c.PC&0xff),
	))
	rows = append(rows, giu.TableRow(
		giu.Label("SP"),
		giu.Labelf("%02x", c.SP>>8),
		giu.Labelf("%02x", c.SP&0xff),
	))
}

func bool2Str(a bool) string {
	if a {
		return "1"
	} else {
		return "0"
	}
}
func makeFlagsTable(c *CPU) *giu.TableWidget {
	rows := []*giu.TableRowWidget{}
	rows = append(rows,
		giu.TableRow(
			giu.Label("S"),
			giu.Label("Z"),
			giu.Label("AC"),
			giu.Label("P"),
			giu.Label("C"),
		))
	rows = append(rows,
		giu.TableRow(
			giu.Label(bool2Str(c.sign)),
			giu.Label(bool2Str(c.zero)),
			giu.Label(bool2Str(c.auxCarry)),
			giu.Label(bool2Str(c.parity)),
			giu.Label(bool2Str(c.carry)),
		))

	return giu.Table().Rows(rows...).Size(140, 45)
}
func rebuildHexTable(c *CPU, start uint16) *giu.TableWidget {
	// var row giu.TableRowWidget
	// var rowWidgets [17]giu.LabelWidget
	var rowWidgets []giu.Widget
	rows := []*giu.TableRowWidget{}
	roundedStart := int((start / 16) * 16)

	// ok now for every 16 bytes, make a whole row out of them, for 10 rows
	// 10 rows loop
	for i := 0; i < 10; i++ {
		rowWidgets = []giu.Widget{giu.Labelf("%04x", roundedStart+i*16)}
		// 16 bytes loop
		for j := 0; j < 16; j++ {
			if uint16(roundedStart+i*16+j) == c.PC {
				// label = giu.Label("9")
				label := giu.Style().
					SetColor(giu.StyleColorText, colornames.Lightgreen).
					To(
						giu.Labelf("%02x", c.memory[roundedStart+i*16+j]),
					)
				rowWidgets = append(rowWidgets, label)
			} else {
				label := giu.Labelf("%02x", c.memory[roundedStart+i*16+j])
				rowWidgets = append(rowWidgets, label)
			}

		}
		rows = append(rows, giu.TableRow(rowWidgets...))
	}

	return giu.Table().Rows(rows...)
}

func DebuggerLoop(c *CPU) {
	// run this function in the window
	rebuildRows(c)

	top := giu.Row(
		makeFlagsTable(c),
		giu.Button("Run").OnClick(
			func() { c.Run() },
		),
	)
	bottom := giu.Row(
		giu.Column(
			giu.Label("Registers"),
			giu.Table().Rows(
				rows...,
			).Size(90, 140),
		),
		giu.Column(
			giu.Label("Hex Table"),
			rebuildHexTable(c, c.PC).Size(600, 280),
		),
	)
	giu.SingleWindow().Layout(
		giu.Column(
			top,
			bottom,
		),
	)
}

func main() {
	// wnd := giu.NewMasterWindow("Table sorting", 640, 480, 0)
	// giu.Context.FontAtlas.SetDefaultFont("IosevkaTermSlabNerdFontMono-Medium.ttf", 18)
	// wnd.Run(loop)
}
