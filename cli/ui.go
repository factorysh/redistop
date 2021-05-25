package cli

import (
	"bytes"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	ui "github.com/gizak/termui/v3"

	"github.com/gizak/termui/v3/widgets"
	"github.com/rivo/tview"
)

type AppUI struct {
	app            *tview.Application
	grid           *tview.Grid
	header         *tview.Table
	graph          *tview.TextView
	splash         *tview.TextView
	cmds           *tview.Table
	ips            *tview.Table
	memories       *tview.Table
	pile           *tview.Flex
	keyspaces      *tview.Table
	clients        *tview.Table
	persistence    *tview.Table
	pubsub         *tview.Table
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
	a.splash.SetText(b.String())
}

func (a *AppUI) fundation() {
	a.header = tview.NewTable().SetFixed(1, 4)
	a.header.SetBorder(true)
	a.header.SetTitle("Redistop")
	for i := 0; i < 4; i++ {
		a.header.SetCell(0, i, tview.NewTableCell("*"))
	}
	a.graph = tview.NewTextView()
	a.graph.SetBorder(true).SetMouseCapture(
		func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
			if action == tview.MouseLeftDoubleClick {
				_, _, _, h := a.graph.GetInnerRect()
				if h <= 5 {
					a.grid.SetRows(3, 12, 0)
				} else {
					a.grid.SetRows(3, 7, 0)
				}
			}
			return action, event
		})

	a.splash = tview.NewTextView()

	a.cmds = tview.NewTable()
	a.cmds.SetBorder(true)
	a.cmds.SetSeparator(tcell.RuneVLine)
	a.cmds.SetTitle("By command/s")

	a.ips = tview.NewTable()
	a.ips.SetBorder(true)
	a.ips.SetSeparator(tcell.RuneVLine)
	a.ips.SetTitle("By IP/s")

	a.errorPanel = widgets.NewParagraph()
	a.errorPanel.Title = "Error"

	a.pile = tview.NewFlex()
	a.pile.SetDirection(tview.FlexRow)

	a.grid = tview.NewGrid().SetRows(3, 7, 0).SetColumns(0, 0, 40).
		AddItem(a.header, 0, 0, 1, 3, 0, 0, false).
		AddItem(a.graph, 1, 0, 1, 3, 0, 0, false).
		AddItem(a.splash, 2, 0, 1, 2, 0, 0, false).
		AddItem(a.pile, 2, 2, 1, 1, 0, 0, false)

	a.app.SetRoot(a.grid, true).SetFocus(a.grid).EnableMouse(true)

	a.keyspaces = tview.NewTable()
	a.keyspaces.SetBorder(true)
	a.keyspaces.SetTitle("Keyspace")
	a.keyspaces.SetCellSimple(0, 0, "")
	a.keyspaces.SetCellSimple(0, 1, "")
	a.pile.AddItem(a.keyspaces, 3, 1, false)

	a.pubsub = tview.NewTable()
	a.pubsub.SetBorder(true)
	a.pubsub.SetTitle("Pubsub")
	a.pubsub.SetCellSimple(0, 0, "")
	a.pubsub.SetCellSimple(0, 1, "")
	a.pile.AddItem(a.pubsub, 3, 1, false)

	a.memories = tview.NewTable()
	a.memories.SetBorder(true)
	a.memories.SetTitle("Memory")
	for i := 0; i < 4; i++ {
		a.memories.SetCellSimple(i, 0, "")
		a.memories.SetCellSimple(i, 1, "")
	}
	a.pile.AddItem(a.memories, 6, 1, false)

	a.clients = tview.NewTable()
	a.clients.SetBorder(true)
	a.clients.SetTitle("Clients")
	for i := 0; i < 2; i++ {
		a.clients.SetCellSimple(i, 0, "")
		a.clients.SetCellSimple(i, 1, "")
	}
	a.pile.AddItem(a.clients, 4, 1, false)

	a.persistence = tview.NewTable()
	a.persistence.SetBorder(true)
	a.persistence.SetTitle("Persistance")
	for i := 0; i < 3; i++ {
		a.persistence.SetCellSimple(i, 0, "")
		a.persistence.SetCellSimple(i, 1, "")
	}
	a.pile.AddItem(a.persistence, 5, 1, false)
	a.drawSplash()

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
