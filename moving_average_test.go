package main

import (
	"math"
	"testing"
)

const tolerance = .01

func TestMovingAveragePut(t *testing.T) {
	ma := MovingAverage{3, []float64{}}
	ma.Put(45.2)
	ma.Put(46.2)
	ma.Put(62.2)
	ma.Put(61.7)
	if ma.Size != 3 {
		t.Errorf("put failed: expected 3 got %d\n", ma.Size)
	}
}

func TestMovingAverageMean(t *testing.T) {
	ma := MovingAverage{3, []float64{}}
	ma.Put(45.3)
	ma.Put(46.4)
	ma.Put(62.3)
	ma.Put(61.8)
	if ma.Size != 3 {
		t.Errorf("mean failed: expected 3 got %d\n", ma.Size)
	}
	mean := ma.Mean()
	if math.Abs(mean-56.83) > tolerance {
		t.Errorf("mean failed: expected 56.83 got %f\n", mean)
	}
}

func TestMovingAverageMax(t *testing.T) {
	ma := MovingAverage{3, []float64{}}
	ma.Put(45.2)
	ma.Put(46.2)
	ma.Put(62.2)
	ma.Put(61.7)
	if ma.Size != 3 {
		t.Errorf("max failed: expected 3 got %d\n", ma.Size)
	}
	max := ma.Max()
	if max != 62.2 {
		t.Errorf("max failed: expected 62.2 got %f\n", max)
	}
}

func TestMovingAverageMin(t *testing.T) {
	ma := MovingAverage{3, []float64{}}
	ma.Put(45.2)
	ma.Put(46.2)
	ma.Put(62.2)
	ma.Put(61.7)
	if ma.Size != 3 {
		t.Errorf("min failed: expected 3 got %d\n", ma.Size)
	}
	min := ma.Min()
	if min != 46.2 {
		t.Errorf("min failed: expected 46.2 got %f\n", min)
	}
}
