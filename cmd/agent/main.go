package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/adran/coding-agent/internal/agent"
)

const (
	defaultBaseURL      = "https://api.deepseek.com/anthropic"
	defaultAuthFilePath = "auth.json"
)

func main() {
	authOpts, _ := loadAuth(defaultAuthFilePath)

	defaultOpts := []option.RequestOption{
		option.WithBaseURL(defaultBaseURL),
	}
	allOpts := append(defaultOpts, authOpts...)
	client := anthropic.NewClient(allOpts...)

	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	agt := agent.NewAgent(&client.Messages, getUserMessage, []agent.ToolDefinition{agent.ReadFileDefinition})

	if err := agt.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func loadAuth(path string) ([]option.RequestOption, string) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "env vars"
	}
	defer f.Close()
	return loadAuthFromReader(f, filepath.Base(path))
}

type authFile struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

func loadAuthFromReader(r io.Reader, source string) ([]option.RequestOption, string) {
	var a authFile
	if err := json.NewDecoder(r).Decode(&a); err != nil {
		return nil, "env vars"
	}

	if a.APIKey == "" {
		return nil, "env vars"
	}

	var opts []option.RequestOption
	opts = append(opts, option.WithAPIKey(a.APIKey))

	if a.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(a.BaseURL))
	}

	return opts, source
}
