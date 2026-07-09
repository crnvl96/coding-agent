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

func main() {
	authOpts, _ := loadAuth("auth.json")

	defaultOpts := []option.RequestOption{
		option.WithBaseURL("https://api.deepseek.com/anthropic"),
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

	agt := agent.NewAgent(&client.Messages, getUserMessage)

	if err := agt.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

type authFile struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

// loadAuthFromReader parses auth options from any reader — usable in tests
// without touching the filesystem.
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

func loadAuth(path string) ([]option.RequestOption, string) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "env vars"
	}
	defer f.Close()
	return loadAuthFromReader(f, filepath.Base(path))
}
