package cli

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type Pile struct {
	x      int
	y      int
	width  int
	tables []*widgets.Table
}

func NewPile(x, y, width int) *Pile {
	return &Pile{
		x:      x,
		y:      y,
		width:  width,
		tables: make([]*widgets.Table, 0),
	}
}

func (p *Pile) ComputePosition() {
	y := p.y
	for _, table := range p.tables {
		height := len(table.Rows) + 2
		table.SetRect(p.x, y, p.x+p.width, y+height)
		y += height
	}
}

func (p *Pile) Add(table *widgets.Table) {
	p.tables = append(p.tables, table)
}

func (p *Pile) Render() {
	for _, table := range p.tables {
		if len(table.Rows) == 0 {
			continue
		}
		empty := false
		for _, row := range table.Rows {
			empty = empty || len(row) == 0
		}
		if !empty {
			ui.Render(table)
		}
	}
}
