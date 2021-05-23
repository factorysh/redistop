package circular

type Circular struct {
	points []int
	scale  float64
	poz    int
}

func NewCircular(size int, scale float64) *Circular {
	return &Circular{
		points: make([]int, size),
		scale:  scale,
		poz:    0,
	}
}

// Add value to the current one
func (c *Circular) Add(value int) {
	c.points[c.poz] += value
}

// Next move current value to the next slot
func (c *Circular) Next() {
	c.poz++
	if c.poz >= len(c.points) {
		c.poz = 0
	}
	c.points[c.poz] = 0
}

// Values return values, scaled, with right padding
func (c *Circular) Values() []float64 {
	v := make([]float64, len(c.points))
	l := len(c.points)
	for i := 0; i < len(c.points); i++ {
		j := l - c.poz + i - 1
		if j >= l {
			j -= l
		}
		if j < 0 {
			j += l
		}
		v[j] = float64(c.points[i]) / c.scale
		j++
	}
	return v
}

func (c *Circular) LastValues(size int) []float64 {
	return c.Values()[len(c.points)-size:]
}
