package circular

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCircular(t *testing.T) {
	c := NewCircular(5, 2.0)
	c.Add(12)
	c.Next()
	c.Add(10)
	assert.Equal(t, []float64{0, 0, 0, 6, 5}, c.Values())
	assert.Equal(t, []float64{0, 6, 5}, c.LastValues(3))
}
