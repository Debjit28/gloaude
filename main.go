package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
)

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(prompt),
	}

	tools := []openai.ChatCompletionToolUnionParam{
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Read",
			Description: openai.String("Read and return the contents of a file"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"file_path": map[string]any{
						"type":        "string",
						"description": "The path to the file to read",
					},
				},
				"required": []string{"file_path"},
			},
		}),
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Write",
			Description: openai.String("Write content to a file"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"file_path": map[string]any{
						"type":        "string",
						"description": "The path of the file to write to",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "The content to write to the file",
					},
				},
				"required": []string{"file_path", "content"},
			},
		}),
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Bash",
			Description: openai.String("Execute a shell command"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "The command to execute",
					},
				},
				"required": []string{"command"},
			},
		}),
	}

	for round := 0; round < 10; round++ {
		params := openai.ChatCompletionNewParams{
			Model:    shared.ChatModel("anthropic/claude-haiku-4.5"),
			Messages: messages,
			Tools:    tools,
		}

		resp, err := client.Chat.Completions.New(context.Background(), params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(resp.Choices) == 0 {
			panic("No choices in response")
		}

		choice := resp.Choices[0]
		messages = append(messages, choice.Message.ToParam())

		if len(choice.Message.ToolCalls) == 0 {
			fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")
			fmt.Print(choice.Message.Content)
			return
		}

		for _, toolCall := range choice.Message.ToolCalls {
			var result string

			switch toolCall.Function.Name {
			case "Read":
				var args struct {
					FilePath string `json:"file_path"`
				}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					result = fmt.Sprintf("Error parsing args: %v", err)
				} else {
					content, err := os.ReadFile(args.FilePath)
					if err != nil {
						result = fmt.Sprintf("Error reading file: %v", err)
					} else {
						result = string(content)
					}
				}

			case "Write":
				var args struct {
					FilePath string `json:"file_path"`
					Content  string `json:"content"`
				}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					result = fmt.Sprintf("Error parsing args: %v", err)
				} else {
					if err := os.WriteFile(args.FilePath, []byte(args.Content), 0644); err != nil {
						result = fmt.Sprintf("Error writing file: %v", err)
					} else {
						result = "OK"
					}
				}

			case "Bash":
				var args struct {
					Command string `json:"command"`
				}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					result = fmt.Sprintf("Error parsing args: %v", err)
				} else {
					cmd := exec.Command("sh", "-c", args.Command)
					output, err := cmd.CombinedOutput()
					if err != nil {
						result = fmt.Sprintf("Error: %v\n%s", err, string(output))
					} else {
						result = string(output)
					}
				}

			default:
				result = fmt.Sprintf("Unknown tool: %s", toolCall.Function.Name)
			}

			messages = append(messages, openai.ToolMessage(result, toolCall.ID))
		}
	}

	fmt.Println("Maximum tool call rounds reached")
}