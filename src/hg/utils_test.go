package hg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAngleShortestSweep(t *testing.T) {
	assert.Equal(t, float32(0), AngleShortestSweep(10, 10))
	assert.Equal(t, float32(-20), AngleShortestSweep(10, 350))
	assert.Equal(t, float32(20), AngleShortestSweep(350, 10))

	assert.Equal(t, float32(-30), AngleShortestSweep(200, 170))
	assert.Equal(t, float32(30), AngleShortestSweep(170, 200))
	assert.Equal(t, float32(80), AngleShortestSweep(356, 114))
}
