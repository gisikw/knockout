package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func cmdAgentTriage(args []string) int {
	ticketsDir, args, err := resolveProjectTicketsDir(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent triage: %v\n", err)
		return 1
	}
	if ticketsDir == "" {
		fmt.Fprintf(os.Stderr, "ko agent triage: no .ko/tickets directory found (use --project or run from a project dir)\n")
		return 1
	}

	args = reorderArgs(args, map[string]bool{})

	fs := flag.NewFlagSet("agent triage", flag.ContinueOnError)
	verbose := fs.Bool("verbose", false, "stream full agent output to stdout")
	fs.BoolVar(verbose, "v", false, "stream full agent output to stdout")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "ko agent triage: %v\n", err)
		return 1
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "ko agent triage: ticket ID required")
		return 1
	}

	// Resolve ticket ID
	ticketsDir, id, err := ResolveTicket(ticketsDir, fs.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent triage: %v\n", err)
		return 1
	}

	// Load ticket
	t, err := LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent triage: %v\n", err)
		return 1
	}

	if t.Triage == "" {
		fmt.Fprintf(os.Stderr, "ko agent triage: ticket %s has no triage value\n", id)
		return 1
	}

	// Load pipeline config (required)
	configPath, err := FindPipelineConfig(ticketsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent triage: %v\n", err)
		return 1
	}

	p, err := LoadPipeline(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent triage: %v\n", err)
		return 1
	}

	// Construct prompt: ticket content + triage as instructions
	var prompt strings.Builder
	prompt.WriteString("## Ticket\n\n")
	prompt.WriteString(fmt.Sprintf("# %s\n", t.Title))
	if t.Body != "" {
		prompt.WriteString(t.Body)
	}
	prompt.WriteString("\n\n")
	prompt.WriteString("## Instructions\n\n")
	prompt.WriteString(t.Triage)

	// Ensure artifact dir and workspace
	artifactDir, err := EnsureArtifactDir(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent triage: %v\n", err)
		return 1
	}

	wsDir, err := CreateWorkspace(artifactDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent triage: %v\n", err)
		return 1
	}

	// Force allowAll=true for triage operations regardless of pipeline config
	cmd := p.Adapter().BuildCommand(prompt.String(), p.Model, "", true, p.AllowedTools)

	// Apply timeout
	timeout, err := parseTimeout(p.StepTimeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent triage: invalid step_timeout: %v\n", err)
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmdCtx := exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
	cmdCtx.Stdin = cmd.Stdin

	// Build env: inherit from adapter or OS, then append ko usage tokens
	baseEnv := os.Environ()
	if cmd.Env != nil {
		baseEnv = cmd.Env
	}
	cmdCtx.Env = append(baseEnv,
		"TICKETS_DIR="+ticketsDir,
		"KO_TICKET_WORKSPACE="+wsDir,
		"KO_ARTIFACT_DIR="+artifactDir,
	)

	if *verbose {
		cmdCtx.Stdout = os.Stdout
		cmdCtx.Stderr = os.Stderr
		if err := cmdCtx.Run(); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				fmt.Fprintf(os.Stderr, "ko agent triage: timed out after %v\n", timeout)
			}
			return 1
		}
	} else {
		out, err := cmdCtx.Output()
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				fmt.Fprintf(os.Stderr, "ko agent triage: timed out after %v\n", timeout)
			} else {
				if exitErr, ok := err.(*exec.ExitError); ok {
					fmt.Fprintf(os.Stderr, "%s", string(exitErr.Stderr))
				}
				fmt.Fprintf(os.Stderr, "ko agent triage: agent command failed: %v\n", err)
			}
			return 1
		}
		_ = out
	}

	// On success: reload ticket (model may have modified it), clear triage, save
	t, err = LoadTicket(ticketsDir, id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ko agent triage: failed to reload ticket: %v\n", err)
		return 1
	}
	t.Triage = ""
	if err := SaveTicket(ticketsDir, t); err != nil {
		fmt.Fprintf(os.Stderr, "ko agent triage: failed to save ticket: %v\n", err)
		return 1
	}

	fmt.Printf("%s: triage cleared\n", id)
	return 0
}
