//go:build e2e

// Package examples_test provides E2E tests for all SDK examples.
// Run with: go test -tags=e2e -timeout=30m ./examples/...
package examples_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

// exampleConfig holds the configuration for running an example.
type exampleConfig struct {
	Name           string
	Dir            string
	Timeout        time.Duration
	Stdin          string                          // Optional stdin input
	Args           []string                        // Optional CLI arguments
	SetupFunc      func(t *testing.T, dir string)  // Optional setup
	CleanupFunc    func(t *testing.T, dir string)  // Optional cleanup
	ExpectedOutput []string                        // Regex patterns to match in stdout
	AllowedErrors  []string                        // Error patterns that are acceptable
}

// runExample executes a single example and validates its output.
func runExample(t *testing.T, cfg exampleConfig) {
	t.Helper()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Setup if needed
	if cfg.SetupFunc != nil {
		cfg.SetupFunc(t, cfg.Dir)
	}

	// Cleanup after test
	if cfg.CleanupFunc != nil {
		defer cfg.CleanupFunc(t, cfg.Dir)
	}

	// Build the example first (faster feedback on compile errors)
	binaryPath := filepath.Join(cfg.Dir, "test_binary")
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", "test_binary", ".")
	buildCmd.Dir = cfg.Dir
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build example: %v\nOutput: %s", err, out)
	}
	defer os.Remove(binaryPath)

	// Run the binary
	cmd := exec.CommandContext(ctx, "./test_binary", cfg.Args...)
	cmd.Dir = cfg.Dir

	// Setup stdin if needed
	var stdin io.WriteCloser
	if cfg.Stdin != "" {
		var err error
		stdin, err = cmd.StdinPipe()
		if err != nil {
			t.Fatalf("Failed to create stdin pipe: %v", err)
		}
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start example: %v", err)
	}

	// Write stdin if provided
	if cfg.Stdin != "" {
		go func() {
			// Small delay to allow process to initialize
			time.Sleep(500 * time.Millisecond)
			io.WriteString(stdin, cfg.Stdin)
			stdin.Close()
		}()
	}

	// Wait for completion
	err := cmd.Wait()
	exitCode := 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		// Context timeout or other error
		t.Logf("Command error: %v", err)
		exitCode = -1
	}

	// Log output for debugging
	t.Logf("Stdout:\n%s", stdout.String())
	if stderr.Len() > 0 {
		t.Logf("Stderr:\n%s", stderr.String())
	}

	// Check exit code
	if exitCode != 0 {
		// Check if this is an allowed/expected error
		for _, pattern := range cfg.AllowedErrors {
			if matched, _ := regexp.MatchString(pattern, stderr.String()); matched {
				t.Logf("Example %s exited with allowed error pattern: %s", cfg.Name, pattern)
				return
			}
			if matched, _ := regexp.MatchString(pattern, stdout.String()); matched {
				t.Logf("Example %s exited with allowed error pattern: %s", cfg.Name, pattern)
				return
			}
		}
		t.Errorf("Example %s exited with code %d\nStderr: %s", cfg.Name, exitCode, stderr.String())
		return
	}

	// Check expected output patterns
	for _, pattern := range cfg.ExpectedOutput {
		matched, err := regexp.MatchString(pattern, stdout.String())
		if err != nil {
			t.Errorf("Invalid regex pattern %q: %v", pattern, err)
			continue
		}
		if !matched {
			t.Errorf("Expected output pattern %q not found in stdout", pattern)
		}
	}
}

// getExampleConfigs returns configurations for all examples.
func getExampleConfigs(baseDir string) []exampleConfig {
	return []exampleConfig{
		// 01_quickstart - Simple one-shot query
		{
			Name:           "01_quickstart",
			Dir:            filepath.Join(baseDir, "01_quickstart"),
			Timeout:        2 * time.Minute,
			ExpectedOutput: []string{`Response:`, `Joke:`},
		},

		// 02_streaming - Streaming responses
		{
			Name:           "02_streaming",
			Dir:            filepath.Join(baseDir, "02_streaming"),
			Timeout:        2 * time.Minute,
			ExpectedOutput: []string{`Starting streaming query`, `Completed in \d+`},
		},

		// 03_mcp_tools - Custom MCP tools
		{
			Name:           "03_mcp_tools",
			Dir:            filepath.Join(baseDir, "03_mcp_tools"),
			Timeout:        2 * time.Minute,
			ExpectedOutput: []string{`Question:`, `Answer:`},
		},

		// 04_hooks - Hook system
		{
			Name:           "04_hooks",
			Dir:            filepath.Join(baseDir, "04_hooks"),
			Timeout:        2 * time.Minute,
			ExpectedOutput: []string{`Hooks configured`},
		},

		// 05_interactive - Interactive REPL (needs stdin)
		{
			Name:           "05_interactive",
			Dir:            filepath.Join(baseDir, "05_interactive"),
			Timeout:        2 * time.Minute,
			Stdin:          "What is 2+2?\nquit\n",
			ExpectedOutput: []string{`Interactive Claude Session`, `Goodbye!`},
		},

		// 06_human_in_loop - Permission callbacks (needs stdin)
		{
			Name:           "06_human_in_loop",
			Dir:            filepath.Join(baseDir, "06_human_in_loop"),
			Timeout:        2 * time.Minute,
			Stdin:          "y\ny\ny\ny\ny\n",
			ExpectedOutput: []string{`Human-in-the-Loop`},
		},

		// 07_events - Event handling
		{
			Name:           "07_events",
			Dir:            filepath.Join(baseDir, "07_events"),
			Timeout:        2 * time.Minute,
			ExpectedOutput: []string{`Event Types Demo`},
		},

		// 08_budget - Budget control
		{
			Name:           "08_budget",
			Dir:            filepath.Join(baseDir, "08_budget"),
			Timeout:        2 * time.Minute,
			ExpectedOutput: []string{`Budget Control Demo`},
		},

		// 09_structured_output - JSON parsing
		{
			Name:           "09_structured_output",
			Dir:            filepath.Join(baseDir, "09_structured_output"),
			Timeout:        2 * time.Minute,
			ExpectedOutput: []string{`Structured Output Demo`},
		},

		// 10_agents - Custom agents
		{
			Name:           "10_agents",
			Dir:            filepath.Join(baseDir, "10_agents"),
			Timeout:        5 * time.Minute,
			ExpectedOutput: []string{`Code Reviewer Agent Example`},
		},

		// 11_system_prompt - System prompt configurations
		{
			Name:           "11_system_prompt",
			Dir:            filepath.Join(baseDir, "11_system_prompt"),
			Timeout:        5 * time.Minute,
			ExpectedOutput: []string{`System Prompt`},
		},

		// 12_tools_option - Tools configuration
		{
			Name:           "12_tools_option",
			Dir:            filepath.Join(baseDir, "12_tools_option"),
			Timeout:        5 * time.Minute,
			ExpectedOutput: []string{`Tools Array Example`},
		},

		// 13_setting_sources - Settings sources
		{
			Name:           "13_setting_sources",
			Dir:            filepath.Join(baseDir, "13_setting_sources"),
			Timeout:        5 * time.Minute,
			ExpectedOutput: []string{`Setting Sources`},
		},

		// 14_plugins - Plugin loading (may fail gracefully)
		{
			Name:           "14_plugins",
			Dir:            filepath.Join(baseDir, "14_plugins"),
			Timeout:        2 * time.Minute,
			ExpectedOutput: []string{`Plugin Example`},
			AllowedErrors:  []string{`plugin`, `not found`, `no such file`},
		},

		// 15_stderr_callback - Stderr capture
		{
			Name:           "15_stderr_callback",
			Dir:            filepath.Join(baseDir, "15_stderr_callback"),
			Timeout:        2 * time.Minute,
			ExpectedOutput: []string{`Running query with stderr capture`},
		},

		// 16_partial_messages - Partial message streaming
		{
			Name:           "16_partial_messages",
			Dir:            filepath.Join(baseDir, "16_partial_messages"),
			Timeout:        5 * time.Minute,
			ExpectedOutput: []string{`Partial Message`},
		},

		// 17_hello_world_ts - Hello world with directory setup
		{
			Name:    "17_hello_world_ts",
			Dir:     filepath.Join(baseDir, "17_hello_world_ts"),
			Timeout: 5 * time.Minute,
			CleanupFunc: func(t *testing.T, dir string) {
				os.RemoveAll(filepath.Join(dir, "agent"))
			},
			ExpectedOutput: []string{`Hello World`},
		},

		// 18_session_api - Session API (basic mode)
		{
			Name:           "18_session_api_basic",
			Dir:            filepath.Join(baseDir, "18_session_api"),
			Timeout:        5 * time.Minute,
			Args:           []string{"basic"},
			ExpectedOutput: []string{`Basic Session`},
		},

		// 18_session_api - Session API (multi-turn mode)
		{
			Name:           "18_session_api_multi_turn",
			Dir:            filepath.Join(baseDir, "18_session_api"),
			Timeout:        5 * time.Minute,
			Args:           []string{"multi-turn"},
			ExpectedOutput: []string{`Multi-Turn`},
		},

		// 18_session_api - Session API (one-shot mode)
		{
			Name:           "18_session_api_one_shot",
			Dir:            filepath.Join(baseDir, "18_session_api"),
			Timeout:        5 * time.Minute,
			Args:           []string{"one-shot"},
			ExpectedOutput: []string{`One-Shot`},
		},

		// 18_session_api - Session API (resume mode)
		{
			Name:           "18_session_api_resume",
			Dir:            filepath.Join(baseDir, "18_session_api"),
			Timeout:        5 * time.Minute,
			Args:           []string{"resume"},
			ExpectedOutput: []string{`Session Resume`},
		},

		// 20_resume_generator - Resume generation
		{
			Name:    "20_resume_generator",
			Dir:     filepath.Join(baseDir, "20_resume_generator"),
			Timeout: 10 * time.Minute,
			Args:    []string{"Test Person"},
			CleanupFunc: func(t *testing.T, dir string) {
				os.RemoveAll(filepath.Join(dir, "agent"))
			},
			ExpectedOutput: []string{`Generating resume for:`},
		},

		// 21_research_agent - Research agent with prompts
		{
			Name:    "21_research_agent",
			Dir:     filepath.Join(baseDir, "21_research_agent"),
			Timeout: 3 * time.Minute,
			Stdin:   "What is Go?\nexit\n",
			SetupFunc: func(t *testing.T, dir string) {
				// Verify prompts directory exists
				promptsDir := filepath.Join(dir, "prompts")
				if _, err := os.Stat(promptsDir); os.IsNotExist(err) {
					t.Skipf("prompts directory not found: %s", promptsDir)
				}
			},
			CleanupFunc: func(t *testing.T, dir string) {
				os.RemoveAll(filepath.Join(dir, "logs"))
				os.RemoveAll(filepath.Join(dir, "files"))
			},
			ExpectedOutput: []string{`Research Agent`},
		},
	}
}

func TestExamples(t *testing.T) {
	// Verify claude CLI is available
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("Skipping E2E tests: claude CLI not found in PATH")
	}

	// Check if Claude CLI is authenticated by running --version
	// (if not authenticated, the CLI still works for version check)
	versionCmd := exec.Command("claude", "--version")
	if err := versionCmd.Run(); err != nil {
		t.Skip("Skipping E2E tests: claude CLI not working")
	}

	// Get base directory (examples/)
	baseDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// If we're not in the examples directory, try to find it
	if filepath.Base(baseDir) != "examples" {
		baseDir = filepath.Join(baseDir, "examples")
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			t.Fatalf("Could not find examples directory")
		}
	}

	configs := getExampleConfigs(baseDir)

	for _, cfg := range configs {
		cfg := cfg // capture for closure
		t.Run(cfg.Name, func(t *testing.T) {
			// Run sequentially to avoid API rate limits
			runExample(t, cfg)
		})
	}
}
