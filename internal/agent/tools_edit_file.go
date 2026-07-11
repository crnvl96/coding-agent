package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

// EditFileInput is the input for the edit_file tool.
type EditFileInput struct {
	Path   string `json:"path" jsonschema_description:"The path to the file to modify."`
	OldStr string `json:"old_str" jsonschema_description:"Text to search for — must match exactly and be unique in the file."`
	NewStr string `json:"new_str" jsonschema_description:"Text to replace old_str with."`
}

// EditFileInputSchema is the JSON Schema for [EditFileInput].
var EditFileInputSchema = GenerateSchema[EditFileInput]()

// EditFileDefinition is the tool definition for editing files.
var EditFileDefinition = ToolDefinition{
	Name: "edit_file",
	Description: `Make edits to a text file.

Replaces 'old_str' with 'new_str' in the given file. 'old_str' and 'new_str' MUST be different from each other.

If the file doesn't exist, it will be created.`,
	InputSchema: EditFileInputSchema,
	Function:    EditFile,
}

// EditFile replaces old_str with new_str in the file at path.
// old_str must appear exactly once; zero or multiple matches produce an error.
// When old_str is empty on an existing file, new_str is appended to the end.
// When the file doesn't exist and old_str is empty, the file is created.
func EditFile(input json.RawMessage) (string, error) {
	var in EditFileInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", err
	}

	if in.Path == "" || in.OldStr == in.NewStr {
		return "", fmt.Errorf("invalid input parameters")
	}

	content, err := os.ReadFile(in.Path)
	if err != nil {
		if os.IsNotExist(err) && in.OldStr == "" {
			return createFile(in.Path, in.NewStr)
		}
		return "", err
	}

	old := string(content)

	var newContent string
	if in.OldStr == "" {
		newContent = old + in.NewStr
	} else {
		count := strings.Count(old, in.OldStr)
		switch {
		case count == 0:
			return "", fmt.Errorf("old_str not found in file")
		case count > 1:
			return "", fmt.Errorf("old_str found %d times in file, must be unique", count)
		}
		newContent = strings.Replace(old, in.OldStr, in.NewStr, 1)
	}

	if err := os.WriteFile(in.Path, []byte(newContent), 0o644); err != nil {
		return "", err
	}

	return "OK", nil
}

// createFile writes content to a new file, creating parent directories as needed.
func createFile(filePath, content string) (string, error) {
	dir := path.Dir(filePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}
	if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	return fmt.Sprintf("Successfully created file %s", filePath), nil
}
