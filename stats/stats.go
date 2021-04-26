package stats

import (
	"sort"

	"github.com/factorysh/redistop/monitor"
)

type Line struct {
	data map[string][]int
	max  int
	poz  int
}

func NewLine(max int) *Line {
	return &Line{
		data: make(map[string][]int),
		max:  max,
		poz:  0,
	}
}

func (l *Line) Incr(key string) {
	_, ok := l.data[key]
	if !ok {
		l.data[key] = make([]int, l.max)
	}
	l.data[key][l.poz] += 1
}

func (l *Line) Next() {
	l.poz += 1
	if l.poz >= l.max {
		l.poz = 0
	}
}

type SortedLine struct {
	K string
	V []float64
}

func (l Line) Values() []SortedLine {
	b := make(ByValue, len(l.data))
	i := 0
	for k, v := range l.data {
		b[i] = KV{
			K: k,
			V: v[l.poz],
		}
		i++
	}
	sort.Sort(b)
	r := make([]SortedLine, len(l.data))
	for i, kv := range b {
		r[i] = SortedLine{
			K: kv.K,
			V: make([]float64, l.max),
		}
		for j := 0; j < l.max; j++ {
			p := j + l.poz + 1
			if p <= 0 {
				p += l.max
			}
			if p >= l.max {
				p -= l.max
			}
			r[i].V[j] = float64(l.data[kv.K][p])
		}
	}
	return r
}

type Stats struct {
	Commands *Line
	Ips      *Line
}

func New(size int) *Stats {
	return &Stats{
		Commands: NewLine(size),
		Ips:      NewLine(size),
	}
}

type KV struct {
	K string
	V int
}

type ByValue []KV

func (a ByValue) Len() int      { return len(a) }
func (a ByValue) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByValue) Less(i, j int) bool {
	if a[i].V == a[j].V {
		return a[i].K < a[j].K
	}
	return a[i].V < a[j].V
}

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

func (s *Stats) Feed(line monitor.Line) {
	s.Commands.Incr(line.Command)
	s.Ips.Incr(line.IP)
}

func (s *Stats) Next() {
	s.Commands.Next()
	s.Ips.Next()
}
