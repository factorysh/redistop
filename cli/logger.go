package cli

import (
	"fmt"
	"sync"
	"time"

	"github.com/rivo/tview"
)

type Logger struct {
	block *tview.TextView
	wg    *sync.WaitGroup
}

func (l *Logger) Printf(tpl string, args ...interface{}) {
	l.block.SetText(fmt.Sprintf(tpl, args...))
	if l.wg == nil {
		l.wg = &sync.WaitGroup{}
		l.wg.Add(1)
		go func() {
			l.wg.Wait()
			l.wg = nil
			l.block.SetText("")
		}()
	} else {
		l.wg.Add(1)
	}
	time.AfterFunc(5*time.Second, func() {
		l.wg.Done()
	})
}
