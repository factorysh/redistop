package cli

import (
	"bytes"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type App struct {
	header     *widgets.Table
	graph      *widgets.Sparkline
	graphBox   *widgets.SparklineGroup
	splash     *widgets.Paragraph
	cmds       *widgets.Table
	ips        *widgets.Table
	memories   *widgets.Table
	pile       *Pile
	keyspaces  *widgets.Table
	errorPanel *widgets.Paragraph
	myWidth    int
}

func NewApp() *App {
	app := &App{}
	width, height := ui.TerminalDimensions()
	app.fundation(width, height)
	return app
}

const art = `
               _._
          _.-''__ ''-._
     _.-''    '.  '_.  ''-._
 .-'' .-'''.  '''\/    _.,_ ''-._
(    '      ,       .-'  | ',    )
|'-._'-...-' __...-.''-._|'' _.-'|
|    '-._   '._    /     _.-'    |
 '-._    '-._  '-./  _.-'    _.-'
|'-._'-._    '-.__.-'    _.-'_.-'|
|    '-._'-._        _.-'_.-'    |
 '-._    '-._'-.__.-'_.-'    _.-'
|'-._'-._    '-.__.-'    _.-'_.-'|
|    '-._'-._        _.-'_.-'    |
 '-._    '-._'-.__.-'_.-'    _.-'
     '-._    '-.__.-'    _.-'
         '-._        _.-'
             '-.__.-'
`

func (a *App) fundation(width, height int) {
	if width >= 120 {
		a.myWidth = 120
	} else {
		a.myWidth = 80
	}

	a.header = widgets.NewTable()
	a.header.Rows = make([][]string, 1)
	if a.myWidth > 80 {
		a.header.Rows[0] = make([]string, 6)
	} else {
		a.header.Rows[0] = make([]string, 4)
	}
	a.header.Rows[0][0] = ""
	a.header.SetRect(0, 0, a.myWidth, 3)

	a.graph = widgets.NewSparkline()
	a.graphBox = widgets.NewSparklineGroup(a.graph)
	fatGraphY := 8
	if height > 40 {
		fatGraphY = 16
	}
	a.graphBox.SetRect(0, 3, a.myWidth, fatGraphY)

	a.splash = widgets.NewParagraph()
	b := &bytes.Buffer{}
	for i := 0; i < (height-fatGraphY-3-17)/2; i++ {
		b.WriteRune('\n')
	}
	for _, line := range strings.Split(art, "\n") {
		b.WriteString("                          ")
		b.WriteString(line)
		b.WriteRune('\n')
	}
	a.splash.Text = b.String()
	a.splash.SetRect(0, fatGraphY, 80, height-3)
	ui.Render(a.splash)

	a.cmds = widgets.NewTable()
	a.cmds.RowSeparator = false
	a.cmds.Title = "By command/s"
	a.cmds.ColumnWidths = []int{30, 10}
	a.cmds.SetRect(0, fatGraphY, 40, height-3)

	a.ips = widgets.NewTable()
	a.ips.RowSeparator = false
	a.ips.Title = "By IP/s"
	a.ips.SetRect(41, fatGraphY, 80, height-3)

	a.pile = NewPile(81, fatGraphY, 39)

	a.keyspaces = widgets.NewTable()
	a.pile.Add(a.keyspaces)
	a.keyspaces.RowSeparator = false
	a.keyspaces.Title = "Keyspace"
	a.keyspaces.Rows = make([][]string, 2)

	a.errorPanel = widgets.NewParagraph()
	a.errorPanel.Title = "Error"
	a.errorPanel.SetRect(0, height-3, a.myWidth, height)

	if a.myWidth > 80 {
		a.memories = widgets.NewTable()
		a.pile.Add(a.memories)
		a.memories.RowSeparator = false
		a.memories.Title = "Memory"
		a.memories.Rows = make([][]string, 4)
	}

	a.pile.ComputePosition()
}

func (a *App) Alert(msg string) {
	argh := widgets.NewParagraph()
	argh.SetRect(20, 6, a.myWidth-20, 11)
	buff := &bytes.Buffer{}
	buff.WriteRune('\n')
	for i := 0; i < (a.myWidth-40-len(msg))/2; i++ {
		buff.WriteRune(' ')
	}
	buff.WriteString(msg)
	argh.Text = buff.String()
	argh.TextStyle.Fg = ui.ColorRed
	argh.Block.BorderStyle.Fg = ui.ColorRed
	ui.Render(argh)
}
