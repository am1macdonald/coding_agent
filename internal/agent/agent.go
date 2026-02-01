package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type Agent struct {
	client       anthropic.Client
	conversation []anthropic.MessageParam
}

func (a *Agent) call() *anthropic.Message {
	message, err := a.client.Messages.New(context.TODO(), anthropic.MessageNewParams{
		MaxTokens: 1024,
		Messages:  a.conversation,
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
	})

	if err != nil {
		panic(err.Error())
	}

	return message
}

func (a *Agent) Message(input string) string {
	a.conversation = append(a.conversation, anthropic.NewUserMessage(anthropic.NewTextBlock(input)))

	m := a.call()

	a.conversation = append(a.conversation, m.ToParam())

	return fmt.Sprintf("%+v\n", m.Content)
}

func NewAgent() *Agent {
	client := anthropic.NewClient(
		option.WithAPIKey(os.Getenv("ANTHROPIC_KEY")),
	)
	return &Agent{
		client: client,
	}
}
