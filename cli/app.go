package cli

import (
	"fmt"
	"time"

	"github.com/factorysh/redistop/monitor"
	ui "github.com/gizak/termui/v3"

	_log "log"
)

type AppConfig struct {
	Host      string
	Password  string
	Frequency time.Duration // Stats per commands and per IPs, every freq seconds
}

type App struct {
	config *AppConfig
	redis  *monitor.RedisServer
	log    *Logger
	ui     *AppUI
}

func NewApp(cfg *AppConfig) *App {
	if cfg.Frequency == 0 {
		cfg.Frequency = 2 * time.Second
	}
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

	a.ui = NewAppUI()

	infos, err := a.redis.Info()
	if err != nil {
		return err
	}

	a.ui.header.Title = fmt.Sprintf("Redis Top -[ v%s/%s pid: %s port: %s hz: %s uptime: %sd ]",
		infos["redis_version"],
		infos["multiplexing_api"],
		infos["process_id"],
		infos["tcp_port"],
		infos["hz"],
		infos["uptime_in_days"],
	)
	ui.Render(a.ui.header)
	ui.Render(a.ui.graphBox)
	ui.Render(a.ui.errorPanel)

	a.log = &Logger{
		block: a.ui.errorPanel,
	}

	a.MonitorLoop()
	a.InfoLoop()
	if a.ui.myWidth > 80 {
		a.MemoryLoop()
	}

	for e := range ui.PollEvents() {
		if e.Type == ui.KeyboardEvent {
			break
		}
	}
	return nil
}
