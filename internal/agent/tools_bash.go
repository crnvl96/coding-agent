package agent

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// BashInput is the input for the bash tool.
type BashInput struct {
	Command string `json:"command" jsonschema_description:"The bash command to execute."`
}

// BashInputSchema is the JSON Schema for [BashInput].
var BashInputSchema = GenerateSchema[BashInput]()

// BashDefinition is the tool definition for executing bash commands.
var BashDefinition = ToolDefinition{
	Name:        "bash",
	Description: "Execute a bash command and return its output. Use this to run shell commands.",
	InputSchema: BashInputSchema,
	Function:    Bash,
}

// Bash executes a bash command and returns its combined output (stdout + stderr).
// Command failures are returned as part of the output string, not as an error,
// so the model can see and understand failure details.
func Bash(input json.RawMessage) (string, error) {
	var in BashInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", err
	}

	cmd := exec.Command("bash", "-c", in.Command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Sprintf("Command failed with error: %s\nOutput: %s", err.Error(), string(output)), nil
	}

	return strings.TrimSpace(string(output)), nil
}
