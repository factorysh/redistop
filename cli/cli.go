package cli

import (
	"fmt"
	_log "log"

	"github.com/factorysh/redistop/monitor"
	ui "github.com/gizak/termui/v3"
)

const freq = 2 // Stats per commands and per IPs, every freq seconds

func Top(host, password string) error {
	_log.Printf("Connecting to redis://%s\n", host)
	redis, err := monitor.Redis(host, password)
	if err != nil {
		return err
	}

	if err := ui.Init(); err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	app := NewApp()

	infos, err := redis.Info()
	if err != nil {
		return err
	}

	app.header.Title = fmt.Sprintf("Redis Top -[ v%s/%s pid: %s port: %s hz: %s uptime: %sd ]",
		infos["redis_version"],
		infos["multiplexing_api"],
		infos["process_id"],
		infos["tcp_port"],
		infos["hz"],
		infos["uptime_in_days"],
	)
	ui.Render(app.header)
	ui.Render(app.graphBox)
	ui.Render(app.errorPanel)

	log := &Logger{
		block: app.errorPanel,
	}

	MonitorLoop(redis, app, log)
	InfoLoop(redis, app, log)
	if app.myWidth > 80 {
		MemoryLoop(redis, app, log)
	}

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
	return nil
}
