// BSD 2-Clause License
//
// Copyright (c) 2020, Andrea Giacomo Baldan
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package backend contains all backend related modules and utilies to perform
// aggregations and analysis of incoming server statistics
package aggregator

import (
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
