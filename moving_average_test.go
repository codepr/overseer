package overseer

import (
	// "math"
	"testing"
	"time"
)

const tolerance = .01

func TestMovingAveragePut(t *testing.T) {
	ma := MovingAverage{3, []time.Duration{}}
	ma.Put(45 * time.Millisecond)
	ma.Put(46 * time.Millisecond)
	ma.Put(62 * time.Millisecond)
	ma.Put(61 * time.Millisecond)
	if ma.Size != 3 {
		t.Errorf("put failed: expected 3 got %d\n", ma.Size)
	}
}

func TestMovingAverageMean(t *testing.T) {
	ma := MovingAverage{3, []time.Duration{}}
	ma.Put(45 * time.Millisecond)
	ma.Put(46 * time.Millisecond)
	ma.Put(62 * time.Millisecond)
	ma.Put(61 * time.Millisecond)
	if ma.Size != 3 {
		t.Errorf("mean failed: expected 3 got %d\n", ma.Size)
	}
	mean := ma.Mean()
	expected, _ := time.ParseDuration("56.333333ms")
	if mean != expected {
		t.Errorf("mean failed: expected %v got %v\n", expected, mean)
	}
}

func TestMovingAverageMax(t *testing.T) {
	ma := MovingAverage{3, []time.Duration{}}
	ma.Put(45 * time.Millisecond)
	ma.Put(46 * time.Millisecond)
	ma.Put(62 * time.Millisecond)
	ma.Put(61 * time.Millisecond)
	if ma.Size != 3 {
		t.Errorf("max failed: expected 3 got %d\n", ma.Size)
	}
	max := ma.Max()
	if max != 62*time.Millisecond {
		t.Errorf("max failed: expected 62.2 got %v\n", max)
	}
}

func TestMovingAverageMin(t *testing.T) {
	ma := MovingAverage{3, []time.Duration{}}
	ma.Put(45 * time.Millisecond)
	ma.Put(46 * time.Millisecond)
	ma.Put(62 * time.Millisecond)
	ma.Put(61 * time.Millisecond)
	if ma.Size != 3 {
		t.Errorf("min failed: expected 3 got %d\n", ma.Size)
	}
	min := ma.Min()
	if min != 46*time.Millisecond {
		t.Errorf("min failed: expected 46.2 got %v\n", min)
	}
}
