package main

import (
	. "github.com/codepr/overseer"
)

func main() {
	urls := []string{"http://localhost:7892", "http://localhost:9898"}
	agent := NewAgent(urls, 12, 5000)
	agent.Run()
}
