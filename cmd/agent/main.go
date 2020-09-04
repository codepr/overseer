package main

import (
	"github.com/codepr/overseer/agent"
)

func main() {
	agent, err := agent.NewFromConfig("./conf.yaml")
	if err != nil {
		panic(err)
	}
	agent.Run()
}
