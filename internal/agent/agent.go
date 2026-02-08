package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

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
			case "write_file":
				var input struct {
					Path     string `json:"path"`
					Contents string `json:"contents"`
				}

				err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input)
				if err != nil {
					panic(err)
				}

				response = WriteFile(input.Path, input.Contents)
			case "list_directory":
				var input struct {
					Path string `json:"path"`
				}

				err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input)
				if err != nil {
					panic(err)
				}

				response = ListDirectory(input.Path)

			case "search_files":
				var input struct {
					Pattern string `json:"pattern"`
					Path    string `json:"path"`
				}

				err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input)
				if err != nil {
					panic(err)
				}

				response = SearchFiles(input.Pattern, input.Path)
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
			{
				Name:        "write_file",
				Description: anthropic.String("Writes contents to a file at the specified path. Creates the file if it doesn't exist, overwrites if it does."),
				InputSchema: WriteFileInputSchema,
			},
			{
				Name:        "list_directory",
				Description: anthropic.String("Lists files and directories in the specified path. Returns detailed file information."),
				InputSchema: ListDirectoryInputSchema,
			},
			{
				Name:        "search_files",
				Description: anthropic.String("Searches for a pattern in files using grep. Returns matching lines with line numbers."),
				InputSchema: SearchFilesInputSchema,
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
	Path string `json:"path" jsonschema_description:"The file system path."`
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

type WriteFileInput struct {
	Path     string `json:"path" jsonschema_description:"The file system path where the file should be written."`
	Contents string `json:"contents" jsonschema_description:"The contents to write to the file."`
}

var WriteFileInputSchema = GenerateSchema[WriteFileInput]()

type WriteFileResponse string

func WriteFile(path string, contents string) WriteFileResponse {
	err := os.WriteFile(path, []byte(contents), 0644)
	if err != nil {
		return WriteFileResponse(fmt.Sprintf("Error writing file: %s", err.Error()))
	}

	return WriteFileResponse(fmt.Sprintf("Successfully wrote %d bytes to %s", len(contents), path))
}

type ListDirectoryInput struct {
	Path string `json:"path" jsonschema_description:"The directory path to list. Defaults to current directory."`
}

var ListDirectoryInputSchema = GenerateSchema[ListDirectoryInput]()

type ListDirectoryResponse string

func ListDirectory(path string) ListDirectoryResponse {
	if path == "" {
		path = "."
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return ListDirectoryResponse(err.Error())
	}

	result := ""
	for _, entry := range entries {
		info, _ := entry.Info()
		result += fmt.Sprintf("%s %10d %s\n", entry.Name(), info.Size(), info.ModTime().Format("2006-01-02 15:04:05"))
	}

	return ListDirectoryResponse(result)
}

type SearchFilesInput struct {
	Pattern string `json:"pattern" jsonschema_description:"The search pattern to look for."`
	Path    string `json:"path" jsonschema_description:"The path to search in. Defaults to current directory."`
}

var SearchFilesInputSchema = GenerateSchema[SearchFilesInput]()

type SearchFilesResponse string

func SearchFiles(pattern string, path string) SearchFilesResponse {
	if path == "" {
		path = "."
	}

	cmd := exec.Command("grep", "-r", "-n", "-I", pattern, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) == 0 {
			return SearchFilesResponse("No matches found")
		}
	}

	return SearchFilesResponse(string(output))
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
