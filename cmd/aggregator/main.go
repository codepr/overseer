package main

import (
	"github.com/codepr/overseer/backend"
)

func main() {
	aggregator := backend.NewAggregator()
	aggregator.Run()
}
