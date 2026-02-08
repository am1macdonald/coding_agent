package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/am1macdonald/coding_agent/internal/agent"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func ReadFile(filePath string) (string, bool) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", false
	}
	return string(bytes), true
}

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
	client := anthropic.NewClient(
		option.WithAPIKey(os.Getenv("ANTHROPIC_KEY")),
	)
	agent := agent.NewAgent(&client)

	for {
		input := getUserInput()
		response := agent.Message(input)
		fmt.Printf("%+v", response)
	}
}
