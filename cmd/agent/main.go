package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
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

	agt := agent.NewAgent(&client, getUserMessage)

	if err := agt.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

type authFile struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

func loadAuth(path string) ([]option.RequestOption, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "env vars"
	}

	var a authFile
	if err := json.Unmarshal(data, &a); err != nil {
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

	return opts, filepath.Base(path)
}
