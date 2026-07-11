package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ListFilesInput is the input for the list_files tool.
type ListFilesInput struct {
	Path string `json:"path,omitempty" jsonschema_description:"Optional relative path to list files from. Defaults to current directory if not provided."`
}

// ListFilesInputSchema is the JSON Schema for [ListFilesInput].
var ListFilesInputSchema = GenerateSchema[ListFilesInput]()

// ListFilesDefinition is the tool definition for listing files and directories.
var ListFilesDefinition = ToolDefinition{
	Name:        "list_files",
	Description: "List files and directories at a given path. If no path is provided, lists files in the current directory.",
	InputSchema: ListFilesInputSchema,
	Function:    ListFiles,
}

// ListFiles recursively lists files and directories at the given path,
// returning them as a JSON array of relative paths (directories suffixed with /).
func ListFiles(input json.RawMessage) (string, error) {
	var in ListFilesInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", err
	}

	dir := "."
	if in.Path != "" {
		dir = in.Path
	}

	files := make([]string, 0)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		if info.IsDir() {
			files = append(files, relPath+"/")
		} else {
			files = append(files, relPath)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	result, err := json.Marshal(files)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
