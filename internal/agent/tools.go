package agent

import (
	"encoding/json"

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
