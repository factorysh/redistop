package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit(t *testing.T) {
	assert.Equal(t, "42.00", DisplayUnit(42))
	assert.Equal(t, "4.81k", DisplayUnit(4807))
	assert.Equal(t, "42.00M", DisplayUnit(42000000))
	assert.Equal(t, "42.00T", DisplayUnit(42000000000))
}
