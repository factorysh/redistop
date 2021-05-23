package cli

import (
	"fmt"
	"time"

	"github.com/factorysh/redistop/monitor"
	ui "github.com/gizak/termui/v3"
)

func MemoryLoop(redis *monitor.RedisServer, app *AppUI, log *Logger) {
	go func() {
		for {
			m, err := redis.Memory()
			if err != nil {
				log.Printf("Memory Error : %s", err.Error())
			} else {
				app.header.Rows[0][4] = fmt.Sprintf("keys: %d", m.KeysCount)
				app.header.Rows[0][5] = fmt.Sprintf("mem: %s", DisplayUnit(float64(m.PeakAllocated)))
				app.memories.Rows = m.Table()
			}
			kv, err := redis.Info()
			if err != nil {
				log.Printf("Info Memory Error : %s", err.Error())
			} else {
				app.memories.Title = fmt.Sprintf("Memory [ %s ]", kv["maxmemory_policy"])
			}

			if len(app.memories.Rows) > 0 && len(app.memories.Rows[0]) > 0 {
				ui.Render(app.memories)
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
