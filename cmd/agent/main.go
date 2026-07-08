package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/adran/coding-agent/internal/agent"
)

func main() {
	verbose := flag.Bool("verbose", false, "enable verbose logging")
	flag.Parse()

	if *verbose {
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Verbose logging enabled")
	} else {
		log.SetOutput(os.Stdout)
		log.SetFlags(0)
		log.SetPrefix("")
	}

	authOpts, authSource := loadAuth("auth.json", *verbose)
	if *verbose {
		log.Printf("Client initialized (%s)", authSource)
	}

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

	agt := agent.NewAgent(&client, getUserMessage, *verbose)

	if err := agt.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

type authFile struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

func loadAuth(path string, verbose bool) ([]option.RequestOption, string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if verbose {
			log.Printf("No %s found, falling back to env vars (%s)", path, err)
		}
		return nil, "env vars"
	}

	var a authFile
	if err := json.Unmarshal(data, &a); err != nil {
		if verbose {
			log.Printf("%s is invalid JSON, falling back to env vars: %v", path, err)
		}
		return nil, "env vars"
	}

	if a.APIKey == "" {
		if verbose {
			log.Printf("%s has no api_key field, falling back to env vars", path)
		}
		return nil, "env vars"
	}

	var opts []option.RequestOption
	opts = append(opts, option.WithAPIKey(a.APIKey))

	if a.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(a.BaseURL))
	}

	if verbose {
		log.Printf("Loaded credentials from %s", absPath)
	}
	return opts, fmt.Sprintf("%s", absPath)
}
