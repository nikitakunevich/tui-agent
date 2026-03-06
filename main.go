package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/nikitakunevich/tui-agent/agent"
	"github.com/nikitakunevich/tui-agent/llm"
	anthropicprov "github.com/nikitakunevich/tui-agent/llm/anthropic"
	openaiprov "github.com/nikitakunevich/tui-agent/llm/openai"
	"github.com/nikitakunevich/tui-agent/logging"
	"github.com/nikitakunevich/tui-agent/tools"
	"github.com/nikitakunevich/tui-agent/ui"
)

func main() {
	// Load .env file (ignore if missing)
	_ = godotenv.Load()

	provider := flag.String("provider", "openai", "LLM provider: openai or anthropic")
	model := flag.String("model", "", "Model name (default: gpt-4o for openai, claude-sonnet-4-20250514 for anthropic)")
	debug := flag.Bool("debug", false, "Enable debug logging")
	baseURL := flag.String("base-url", "", "Base URL for OpenAI-compatible API")
	prompt := flag.String("prompt", "", "Run non-interactively with this prompt (skip TUI)")
	flag.Parse()

	cleanup, err := logging.Setup("tui-agent.log", *debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup logging: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	slog.Info("tui-agent starting", "provider", *provider, "debug", *debug)

	// Initialize LLM provider
	var llmProvider llm.Provider
	switch *provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "OPENAI_API_KEY is required (set in .env or environment)")
			os.Exit(1)
		}
		modelName := *model
		if modelName == "" {
			modelName = envOrDefault("MODEL_NAME", "gpt-4o")
		}
		bURL := *baseURL
		if bURL == "" {
			bURL = os.Getenv("OPENAI_BASE_URL")
		}
		llmProvider = openaiprov.New(openaiprov.Config{
			APIKey:  apiKey,
			BaseURL: bURL,
			Model:   modelName,
		})
	case "anthropic":
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			fmt.Fprintln(os.Stderr, "ANTHROPIC_API_KEY is required (set in .env or environment)")
			os.Exit(1)
		}
		modelName := *model
		if modelName == "" {
			modelName = envOrDefault("MODEL_NAME", "claude-sonnet-4-20250514")
		}
		llmProvider = anthropicprov.New(anthropicprov.Config{
			APIKey: apiKey,
			Model:  modelName,
		})
	default:
		fmt.Fprintf(os.Stderr, "unknown provider: %s (use 'openai' or 'anthropic')\n", *provider)
		os.Exit(1)
	}

	// Initialize tools
	registry := tools.NewRegistry()
	registry.Register(tools.NewBashTool())

	// Initialize agent
	systemPrompt := "You are a helpful terminal assistant. You can execute bash commands using the bash tool. Be concise in your responses."
	a := agent.New(llmProvider, registry, systemPrompt)

	// Non-interactive mode
	if *prompt != "" {
		result, err := a.Run(context.Background(), *prompt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
		return
	}

	// Run TUI
	p := tea.NewProgram(ui.New(a), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		slog.Error("TUI error", "error", err)
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
