package agent

import (
	"encoding/json"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

// ToolDefinition describes a tool available to the agent: its name, description,
// JSON input schema, and the handler function that executes it.
type ToolDefinition struct {
	Name        string
	Description string
	InputSchema anthropic.ToolInputSchemaParam
	Function    func(json.RawMessage) (string, error)
}

// GenerateSchema derives an [anthropic.ToolInputSchemaParam] from the JSON Schema of type T.
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

// ReadFileInput is the input for the read_file tool.
type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

// ReadFileInputSchema is the JSON Schema for [ReadFileInput].
var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

// ReadFileDefinition is the tool definition for reading files.
var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

// ReadFile reads the file at the given path and returns its contents.
func ReadFile(input json.RawMessage) (string, error) {
	var in ReadFileInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", err
	}
	content, err := os.ReadFile(in.Path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
