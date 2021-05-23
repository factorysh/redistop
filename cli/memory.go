package cli

import (
	"fmt"
	"time"

	ui "github.com/gizak/termui/v3"
)

func (a *App) MemoryLoop() {
	go func() {
		for {
			m, err := a.redis.Memory()
			if err != nil {
				a.log.Printf("Memory Error : %s", err.Error())
			} else {
				if len(a.ui.header.Rows[0]) > 4 {
					a.ui.header.Rows[0][4] = fmt.Sprintf("keys: %d", m.KeysCount)
					a.ui.header.Rows[0][5] = fmt.Sprintf("mem: %s", DisplayUnit(float64(m.PeakAllocated)))
				}
				a.ui.memories.Rows = m.Table()
			}
			kv, err := a.redis.Info()
			if err != nil {
				a.log.Printf("Info Memory Error : %s", err.Error())
			} else {
				a.ui.memories.Title = fmt.Sprintf("Memory [ %s ]", kv["maxmemory_policy"])
			}

			if a.ui.myWidth > 80 {
				if len(a.ui.memories.Rows) > 0 && len(a.ui.memories.Rows[0]) > 0 {
					ui.Render(a.ui.memories)
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
