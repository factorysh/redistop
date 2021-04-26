package stats

import (
	"sort"

	"github.com/factorysh/redistop/monitor"
)

type Stats struct {
	Commands map[string]int
	Ips      map[string]int
}

type KV struct {
	K string
	V int
}

type ByValue []KV

func (a ByValue) Len() int           { return len(a) }
func (a ByValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByValue) Less(i, j int) bool { return a[i].V < a[j].V }

func Count(data map[string]int) ByValue {
	r := make(ByValue, len(data))
	i := 0
	for k, v := range data {
		r[i] = KV{k, v}
		i++
	}
	sort.Sort(r)
	return r
}

func New() *Stats {
	return &Stats{
		Commands: make(map[string]int),
		Ips:      make(map[string]int),
	}
}

func (s *Stats) Feed(line monitor.Line) {
	_, ok := s.Commands[line.Command]
	if ok {
		s.Commands[line.Command] += 1
	} else {
		s.Commands[line.Command] = 1
	}
	_, ok = s.Ips[line.IP]
	if ok {
		s.Ips[line.IP] += 1
	} else {
		s.Ips[line.IP] = 1
	}
}

func (s *Stats) Reset() {
	s.Commands = make(map[string]int)
	s.Ips = make(map[string]int)
}
