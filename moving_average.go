package overseer

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
