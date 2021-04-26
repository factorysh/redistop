package stats

import (
	"fmt"
	"testing"

	"github.com/factorysh/redistop/monitor"
	"github.com/stretchr/testify/assert"
)

func TestStats(t *testing.T) {
	s := New(5)
	assert.Equal(t, 0, s.Commands.poz)
	s.Feed(monitor.Line{
		Command: "GET",
		IP:      "127.0.0.1",
	})
	s.Feed(monitor.Line{
		Command: "GET",
		IP:      "127.0.0.1",
	})
	s.Feed(monitor.Line{
		Command: "SET",
		IP:      "127.0.0.1",
	})
	s.Next()
	assert.Equal(t, 1, s.Commands.poz)
	s.Feed(monitor.Line{
		Command: "GET",
		IP:      "127.0.0.1",
	})
	s.Next()
	assert.Equal(t, 2, s.Commands.poz)
	s.Feed(monitor.Line{
		Command: "GET",
		IP:      "127.0.0.1",
	})
	values := s.Commands.Values()
	fmt.Println(values)
	fmt.Println(s.Commands.data)
	assert.Equal(t, []SortedLine{
		{
			K: "SET",
			V: []float64{0, 0, 1, 0, 0},
		},
		{
			K: "GET",
			V: []float64{0, 0, 2, 1, 1},
		},
	}, values)
}
