package main

import (
	. "github.com/codepr/overseer"
)

func main() {
	agent, err := NewAgentFromConfig("./conf.yaml")
	if err != nil {
		panic(err)
	}
	agent.Run()
}
