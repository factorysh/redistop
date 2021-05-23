package cli

import (
	"fmt"

	"github.com/factorysh/redistop/monitor"
	ui "github.com/gizak/termui/v3"

	_log "log"
)

type AppConfig struct {
	Host     string
	Password string
}

type App struct {
	config *AppConfig
	redis  *monitor.RedisServer
}

func NewApp(cfg *AppConfig) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Serve() error {
	_log.Printf("Connecting to redis://%s\n", a.config.Host)
	var err error
	a.redis, err = monitor.Redis(a.config.Host, a.config.Password)
	if err != nil {
		return err
	}

	if err := ui.Init(); err != nil {
		return fmt.Errorf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	appUI := NewAppUI()

	infos, err := a.redis.Info()
	if err != nil {
		return err
	}

	appUI.header.Title = fmt.Sprintf("Redis Top -[ v%s/%s pid: %s port: %s hz: %s uptime: %sd ]",
		infos["redis_version"],
		infos["multiplexing_api"],
		infos["process_id"],
		infos["tcp_port"],
		infos["hz"],
		infos["uptime_in_days"],
	)
	ui.Render(appUI.header)
	ui.Render(appUI.graphBox)
	ui.Render(appUI.errorPanel)

	log := &Logger{
		block: appUI.errorPanel,
	}

	MonitorLoop(a.redis, appUI, log)
	InfoLoop(a.redis, appUI, log)
	if appUI.myWidth > 80 {
		MemoryLoop(a.redis, appUI, log)
	}

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
	return nil
}
