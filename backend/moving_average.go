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
package backend

import "time"

type MovingAverage struct {
	Size  int
	items []time.Duration
}

func NewMovingAverage(size int) *MovingAverage {
	return &MovingAverage{size, []time.Duration{}}
}

func (ma *MovingAverage) Put(item time.Duration) {
	ma.items = append(ma.items, item)
	if len(ma.items) > ma.Size {
		_, ma.items = ma.items[0], ma.items[1:]
	}
}

func (ma *MovingAverage) Mean() time.Duration {
	if len(ma.items) == 0 {
		return 0.0
	}
	var sum time.Duration = 0.0
	for _, value := range ma.items {
		sum += value
	}
	return sum / time.Duration(len(ma.items))
}

func (ma *MovingAverage) Max() time.Duration {
	if len(ma.items) == 0 {
		return 0.0
	}
	max := ma.items[0]
	for _, value := range ma.items {
		if max < value {
			max = value
		}
	}
	return max
}

func (ma *MovingAverage) Min() time.Duration {
	if len(ma.items) == 0 {
		return 0.0
	}
	min := ma.items[0]
	for _, value := range ma.items {
		if min > value {
			min = value
		}
	}
	return min
}
