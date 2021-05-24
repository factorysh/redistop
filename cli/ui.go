package cli

import (
	"bytes"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"

	"github.com/gizak/termui/v3/widgets"
	"github.com/rivo/tview"
)

type AppUI struct {
	app            *tview.Application
	header         *tview.Table
	graph          *tview.TextView
	splash         *widgets.Paragraph
	cmds           *widgets.Table
	ips            *widgets.Table
	memories       *widgets.Table
	pile           *Pile
	keyspaces      *widgets.Table
	clients        *widgets.Table
	persistence    *widgets.Table
	pubsub         *widgets.Table
	errorPanel     *widgets.Paragraph
	myWidth        int
	fatGraphY      int
	width          int
	height         int
	monitorIsReady bool
}

func NewAppUI() *AppUI {
	appUI := &AppUI{
		app:            tview.NewApplication(),
		monitorIsReady: false,
	}
	appUI.fundation()
	//appUI.WatchResize()
	return appUI
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

func (a *AppUI) resize() {
	a.width, a.height = ui.TerminalDimensions()
	if a.width >= 120 {
		a.myWidth = 120
	} else {
		a.myWidth = 80
	}
	a.fatGraphY = 8
	if a.height > 40 {
		a.fatGraphY = 16
	}
}

func (a *AppUI) draw() {
	a.resize()

	a.header.SetRect(0, 0, a.myWidth, 3)

	if !a.monitorIsReady {
		a.drawSplash()
	} else {
		blank := ui.NewBlock()
		blank.SetRect(0, a.fatGraphY, 80, a.height-3)
		blank.Border = false
		ui.Render(blank)
	}

	a.cmds.SetRect(0, a.fatGraphY, 40, a.height-3)
	a.ips.SetRect(41, a.fatGraphY, 80, a.height-3)

	blank := ui.NewBlock()
	blank.SetRect(80, 0, 120, a.height)
	blank.Border = false
	ui.Render(blank)

	a.pile.y = a.fatGraphY
	a.pile.ComputePosition()
	if a.myWidth > 80 {
		a.pile.Render()
	}

	a.errorPanel.SetRect(0, a.height-3, a.myWidth, a.height)
	ui.Render(a.errorPanel)
}

func (a *AppUI) WatchResize() {
	go func() {
		for {
			width, height := ui.TerminalDimensions()
			if a.width != width || a.height != height {
				a.draw()
			}
			time.Sleep(time.Second)
		}
	}()
}

func (a *AppUI) drawSplash() {
	b := &bytes.Buffer{}
	for i := 0; i < (a.height-a.fatGraphY-3-17)/2; i++ {
		b.WriteRune('\n')
	}
	for _, line := range strings.Split(art, "\n") {
		b.WriteString("                          ")
		b.WriteString(line)
		b.WriteRune('\n')
	}
	a.splash.Text = b.String()
	a.splash.SetRect(0, a.fatGraphY, 80, a.height-3)
	ui.Render(a.splash)
}

func (a *AppUI) fundation() {
	a.header = tview.NewTable().SetFixed(1, 4)
	a.header.SetBorder(true)
	for i := 0; i < 4; i++ {
		a.header.SetCell(0, i, tview.NewTableCell("*"))
	}
	a.graph = tview.NewTextView()
	a.graph.SetBorder(true)
	grid := tview.NewGrid().SetRows(3, 0, 3).SetColumns(0, 30, 30).
		AddItem(a.header, 0, 0, 1, 3, 0, 0, false).
		AddItem(a.graph, 1, 0, 1, 3, 0, 0, false)
	a.header.SetTitle("Redistop")
	a.app.SetRoot(grid, true).SetFocus(grid)

	a.splash = widgets.NewParagraph()

	a.cmds = widgets.NewTable()
	a.cmds.RowSeparator = false
	a.cmds.Title = "By command/s"
	a.cmds.ColumnWidths = []int{30, 10}

	a.ips = widgets.NewTable()
	a.ips.RowSeparator = false
	a.ips.Title = "By IP/s"

	a.errorPanel = widgets.NewParagraph()
	a.errorPanel.Title = "Error"

	a.pile = NewPile(81, a.fatGraphY, 39)

	a.keyspaces = widgets.NewTable()
	a.pile.Add(a.keyspaces)
	a.keyspaces.RowSeparator = false
	a.keyspaces.Title = "Keyspace"
	a.keyspaces.Rows = make([][]string, 1)

	a.pubsub = widgets.NewTable()
	a.pile.Add(a.pubsub)
	a.pubsub.RowSeparator = false
	a.pubsub.Title = "Pubsub"
	a.pubsub.Rows = make([][]string, 1)

	a.memories = widgets.NewTable()
	a.pile.Add(a.memories)
	a.memories.RowSeparator = false
	a.memories.Title = "Memory"
	a.memories.Rows = make([][]string, 4)

	a.clients = widgets.NewTable()
	a.pile.Add(a.clients)
	a.clients.RowSeparator = false
	a.clients.Title = "Clients"
	a.clients.Rows = make([][]string, 2)

	a.persistence = widgets.NewTable()
	a.pile.Add(a.persistence)
	a.persistence.RowSeparator = false
	a.persistence.Title = "Persistance"
	a.persistence.Rows = make([][]string, 3)

	a.pile.ComputePosition()
}

func (a *AppUI) Alert(msg string) {
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
