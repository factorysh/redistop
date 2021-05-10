package cli

import (
	"fmt"
	"sync"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type Logger struct {
	block *widgets.Paragraph
	wg    *sync.WaitGroup
}

func (l *Logger) Printf(tpl string, args ...interface{}) {
	l.block.Text = fmt.Sprintf(tpl, args...)
	ui.Render(l.block)
	if l.wg == nil {
		l.wg = &sync.WaitGroup{}
		l.wg.Add(1)
		go func() {
			l.wg.Wait()
			l.wg = nil
			l.block.Text = ""
			ui.Render(l.block)
		}()
	} else {
		l.wg.Add(1)
	}
	time.AfterFunc(5*time.Second, func() {
		l.wg.Done()
	})
}
