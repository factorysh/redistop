package monitor

import (
	"time"

	"github.com/mediocregopher/radix/v3"
)

type CPU struct {
	sys  float64
	user float64
	ts   time.Time
}

func NewCPU(sys, user float64) *CPU {
	return &CPU{
		sys:  sys,
		user: user,
		ts:   time.Now(),
	}
}

func (c *CPU) Tick(sys, user float64) (float64, float64) {
	now := time.Now()
	delta := now.Sub(c.ts)
	s := (sys - c.sys) / delta.Seconds() * 100
	u := (user - c.user) / delta.Seconds() * 100
	c.sys = sys
	c.user = user
	c.ts = now
	return s, u
}

func (r *RedisServer) InfoCpu() (map[string]string, error) {
	var stats string
	err := r.pool.Do(radix.Cmd(&stats, "INFO", "CPU"))
	if err != nil {
		return nil, err
	}
	return BulkTable(stats)
}
