package monitor

import (
	"time"
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
