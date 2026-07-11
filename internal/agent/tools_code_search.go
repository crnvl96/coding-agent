package agent

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// CodeSearchInput is the input for the code_search tool.
type CodeSearchInput struct {
	Pattern       string `json:"pattern" jsonschema_description:"The search pattern or regex to look for."`
	Path          string `json:"path,omitempty" jsonschema_description:"Optional path to search in (file or directory)."`
	FileType      string `json:"file_type,omitempty" jsonschema_description:"Optional file extension to limit search to (e.g. 'go', 'js', 'py')."`
	CaseSensitive bool   `json:"case_sensitive,omitempty" jsonschema_description:"Whether the search should be case sensitive (default: false)."`
}

// CodeSearchInputSchema is the JSON Schema for [CodeSearchInput].
var CodeSearchInputSchema = GenerateSchema[CodeSearchInput]()

// CodeSearchDefinition is the tool definition for searching code with ripgrep.
var CodeSearchDefinition = ToolDefinition{
	Name: "code_search",
	Description: `Search for code patterns using ripgrep (rg).

Use this to find code patterns, function definitions, variable usage, or any text in the codebase.
You can search by pattern, file type, or directory.`,
	InputSchema: CodeSearchInputSchema,
	Function:    CodeSearch,
}

// CodeSearch runs ripgrep with the given search parameters and returns matching lines.
// Exit code 1 (no matches) is treated as a successful empty result, not an error.
// Output is truncated to 50 lines if there are more matches.
func CodeSearch(input json.RawMessage) (string, error) {
	var in CodeSearchInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", err
	}

	if in.Pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	args := []string{"--line-number", "--with-filename", "--color=never"}

	if !in.CaseSensitive {
		args = append(args, "--ignore-case")
	}

	if in.FileType != "" {
		args = append(args, "--type", in.FileType)
	}

	args = append(args, in.Pattern)

	searchPath := "."
	if in.Path != "" {
		searchPath = in.Path
	}
	args = append(args, searchPath)

	cmd := exec.Command("rg", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "No matches found", nil
		}
		return "", fmt.Errorf("search failed: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "No matches found", nil
	}

	lines := strings.Split(result, "\n")
	if len(lines) > 50 {
		result = strings.Join(lines[:50], "\n") + fmt.Sprintf("\n... (showing first 50 of %d matches)", len(lines))
	}

	return result, nil
}
