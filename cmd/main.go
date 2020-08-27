package main

import (
	"fmt"
	. "github.com/codepr/overseer"
)

func main() {
	urls := []string{"http://localhost:7892", "http://localhost:9898"}
	fmt.Printf("Monitoring websites: %v\n\n", urls)
	agent := NewAgent(urls, 12, 5000)
	agent.Run()
}
