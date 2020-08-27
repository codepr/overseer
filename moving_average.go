package main

type MovingAverage struct {
	Size  int
	items []float64
}

func (ma *MovingAverage) Put(item float64) {
	ma.items = append(ma.items, item)
	if len(ma.items) > ma.Size {
		_, ma.items = ma.items[0], ma.items[1:]
	}
}

func (ma *MovingAverage) Mean() float64 {
	var sum float64 = 0.0
	for _, value := range ma.items {
		sum += value
	}
	return sum / float64(len(ma.items))
}

func (ma *MovingAverage) Max() float64 {
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

func (ma *MovingAverage) Min() float64 {
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
