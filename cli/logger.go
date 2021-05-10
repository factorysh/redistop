package cli

import (
	"fmt"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type Logger struct {
	block *widgets.Paragraph
}

func (l *Logger) Printf(tpl string, args ...interface{}) {
	l.block.Text = fmt.Sprintf(tpl, args...)
	ui.Render(l.block)
}
