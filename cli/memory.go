package cli

import (
	"fmt"
	"time"
)

func (a *App) MemoryLoop() {
	go func() {
		for {
			m, err := a.redis.Memory()
			if err != nil {
				a.log.Printf("Memory Error : %s", err.Error())
			} else {
				if a.ui.myWidth > 80 {
					a.ui.header.GetCell(0, 4).Text = fmt.Sprintf("keys: %d", m.KeysCount)
					a.ui.header.GetCell(0, 5).Text = fmt.Sprintf("mem: %s", DisplayUnit(float64(m.PeakAllocated)))
				}
			}
			kv, err := a.redis.Info()
			if err != nil {
				a.log.Printf("Info Memory Error : %s", err.Error())
			} else {
				a.ui.app.QueueUpdate(func() {
					a.ui.memories.SetTitle(fmt.Sprintf("Memory [ %s ]", kv["maxmemory_policy"]))
					for i, line := range m.Table() {
						for j, col := range line {
							a.ui.memories.GetCell(i, j).Text = col
						}
					}
				})
			}

			time.Sleep(5 * time.Second)
		}
	}()
}
