package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

type Agent struct {
	client       *anthropic.Client
	tools        []anthropic.ToolUnionParam
	conversation []anthropic.MessageParam
}

var readTool = anthropic.ToolParam{}

func (a *Agent) Message(input string) string {
	a.conversation = append(a.conversation, anthropic.NewUserMessage(anthropic.NewTextBlock(input)))

	message, err := a.client.Messages.New(context.TODO(), anthropic.MessageNewParams{
		MaxTokens: 2048,
		Messages:  a.conversation,
		Model:     anthropic.ModelClaudeSonnet4_5_20250929,
		Tools:     a.tools,
	})

	if err != nil {
		panic(err.Error())
	}

	print(color("[assistant]: "))
	for _, block := range message.Content {
		switch block := block.AsAny().(type) {
		case anthropic.TextBlock:
			println(block.Text)
			println()
		case anthropic.ToolUseBlock:
			inputJSON, _ := json.Marshal(block.Input)
			println(block.Name + ": " + string(inputJSON))
			println()
		}
	}

	a.conversation = append(a.conversation, message.ToParam())
	toolResults := []anthropic.ContentBlockParamUnion{}

	for _, block := range message.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.ToolUseBlock:
			print(color("[user (" + block.Name + ")]: "))

			var response any
			switch block.Name {
			case "get_file":
				var input struct {
					Path string `json:"path"`
				}

				err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input)
				if err != nil {
					panic(err)
				}

				response = GetFile(input.Path)
			}

			b, err := json.Marshal(response)
			if err != nil {
				panic(err)
			}

			println(string(b))

			toolResults = append(toolResults, anthropic.NewToolResultBlock(block.ID, string(b), false))
		}

	}
	if len(toolResults) != 0 {
		a.conversation = append(a.conversation, anthropic.NewUserMessage(toolResults...))
	}

	return fmt.Sprintf("%+v\n", message.Content)
}

func NewAgent(client *anthropic.Client) *Agent {
	return &Agent{
		client: client,
		tools: newTools([]anthropic.ToolParam{
			{
				Name:        "get_file",
				Description: anthropic.String("Accepts a file path, then returns the file contents."),
				InputSchema: GetFileInputSchema,
			},
		}),
	}
}

func newTools(tools []anthropic.ToolParam) []anthropic.ToolUnionParam {
	t := make([]anthropic.ToolUnionParam, len(tools))
	for i, toolParam := range tools {
		t[i] = anthropic.ToolUnionParam{OfTool: &toolParam}
	}

	return t
}

type GetFileInput struct {
	Path string `json:"" jsonschema_description:"The file system path."`
}

var GetFileInputSchema = GenerateSchema[GetFileInput]()

type GetFileResponse string

func GetFile(path string) GetFileResponse {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return GetFileResponse(err.Error())
	}

	return GetFileResponse(bytes)
}

func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}

func color(s string) string {
	return fmt.Sprintf("\033[1;%sm%s\033[0m", "33", s)
}
