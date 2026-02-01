package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/am1macdonald/coding_agent/internal/agent"
)

func getUserInput() string {
	text := ""
	for len(text) < 2 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("What do you want?\n")
		text, _ = reader.ReadString('\n')
	}
	return text
}

func main() {

	agent := agent.NewAgent()

	for {
		input := getUserInput()
		response := agent.Message(input)
		fmt.Printf("%+v", response)
	}
}
