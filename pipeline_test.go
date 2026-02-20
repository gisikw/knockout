package main

import "testing"

func TestParsePipelineV2(t *testing.T) {
	config := `
command: test-llm
model: sonnet
max_retries: 1
max_depth: 3
discretion: high
workflows:
  main:
    - name: triage
      type: decision
      prompt: triage.md
      routes:
        - feature
        - investigation
  feature:
    - name: implement
      type: action
      prompt: implement.md
    - name: verify
      type: action
      run: just test
  investigation:
    - name: investigate
      type: action
      prompt: investigate.md
    - name: recommend
      type: decision
      prompt: recommend.md
on_succeed:
  - git add -A
on_close:
  - echo done
`

	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}

	// Top-level scalars
	if p.Command != "test-llm" {
		t.Errorf("Command = %q", p.Command)
	}
	if p.Model != "sonnet" {
		t.Errorf("Model = %q", p.Model)
	}
	if p.MaxRetries != 1 {
		t.Errorf("MaxRetries = %d", p.MaxRetries)
	}
	if p.MaxDepth != 3 {
		t.Errorf("MaxDepth = %d", p.MaxDepth)
	}
	if p.Discretion != "high" {
		t.Errorf("Discretion = %q", p.Discretion)
	}

	// Workflows
	if len(p.Workflows) != 3 {
		t.Fatalf("len(Workflows) = %d, want 3", len(p.Workflows))
	}

	// main workflow
	main := p.Workflows["main"]
	if main == nil {
		t.Fatal("main workflow missing")
	}
	if len(main.Nodes) != 1 {
		t.Fatalf("main.Nodes = %d, want 1", len(main.Nodes))
	}
	triage := main.Nodes[0]
	if triage.Name != "triage" {
		t.Errorf("triage.Name = %q", triage.Name)
	}
	if triage.Type != NodeDecision {
		t.Errorf("triage.Type = %q", triage.Type)
	}
	if triage.Prompt != "triage.md" {
		t.Errorf("triage.Prompt = %q", triage.Prompt)
	}
	if len(triage.Routes) != 2 || triage.Routes[0] != "feature" || triage.Routes[1] != "investigation" {
		t.Errorf("triage.Routes = %v", triage.Routes)
	}

	// feature workflow
	feature := p.Workflows["feature"]
	if feature == nil {
		t.Fatal("feature workflow missing")
	}
	if len(feature.Nodes) != 2 {
		t.Fatalf("feature.Nodes = %d, want 2", len(feature.Nodes))
	}
	if feature.Nodes[0].Type != NodeAction {
		t.Errorf("feature implement type = %q", feature.Nodes[0].Type)
	}
	if feature.Nodes[1].Run != "just test" {
		t.Errorf("feature verify run = %q", feature.Nodes[1].Run)
	}

	// Hooks
	if len(p.OnSucceed) != 1 || p.OnSucceed[0] != "git add -A" {
		t.Errorf("OnSucceed = %v", p.OnSucceed)
	}
	if len(p.OnClose) != 1 || p.OnClose[0] != "echo done" {
		t.Errorf("OnClose = %v", p.OnClose)
	}
}

func TestParsePipelineMaxVisits(t *testing.T) {
	config := `
workflows:
  main:
    - name: check
      type: decision
      prompt: check.md
      max_visits: 3
      routes:
        - main
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	node := p.Workflows["main"].Nodes[0]
	if node.MaxVisits != 3 {
		t.Errorf("MaxVisits = %d, want 3", node.MaxVisits)
	}
}

func TestParsePipelineDefaultMaxVisits(t *testing.T) {
	config := `
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	if p.Workflows["main"].Nodes[0].MaxVisits != 1 {
		t.Errorf("default MaxVisits = %d, want 1", p.Workflows["main"].Nodes[0].MaxVisits)
	}
}

func TestParsePipelineInlineRoutes(t *testing.T) {
	config := `
workflows:
  main:
    - name: triage
      type: decision
      prompt: triage.md
      routes: [feature, investigation]
  feature:
    - name: impl
      type: action
      prompt: impl.md
  investigation:
    - name: inv
      type: action
      prompt: inv.md
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	routes := p.Workflows["main"].Nodes[0].Routes
	if len(routes) != 2 || routes[0] != "feature" || routes[1] != "investigation" {
		t.Errorf("routes = %v", routes)
	}
}

func TestParsePipelineAgent(t *testing.T) {
	config := `
agent: cursor
allow_all_tool_calls: true
model: gpt-4
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	if p.Agent != "cursor" {
		t.Errorf("Agent = %q, want %q", p.Agent, "cursor")
	}
	if !p.AllowAll {
		t.Error("AllowAll = false, want true")
	}
	if p.Command != "" {
		t.Errorf("Command = %q, want empty", p.Command)
	}
}

func TestParsePipelineDefaultAgent(t *testing.T) {
	config := `
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	if p.Agent != "claude" {
		t.Errorf("Agent = %q, want %q", p.Agent, "claude")
	}
}

func TestParsePipelineCommandClearsDefaultAgent(t *testing.T) {
	config := `
command: ./fake-llm
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	if p.Agent != "" {
		t.Errorf("Agent = %q, want empty (command should clear default)", p.Agent)
	}
	if p.Command != "./fake-llm" {
		t.Errorf("Command = %q, want %q", p.Command, "./fake-llm")
	}
}

func TestParsePipelineAgentAndCommandConflict(t *testing.T) {
	config := `
agent: cursor
command: ./fake-llm
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
`
	_, err := ParsePipeline(config)
	if err == nil {
		t.Fatal("expected error for agent + command conflict")
	}
	if !containsStr(err.Error(), "both 'agent' and 'command'") {
		t.Errorf("error = %q, want substring about agent/command conflict", err.Error())
	}
}

func TestParsePipelineValidationErrors(t *testing.T) {
	tests := []struct {
		name   string
		config string
		errMsg string
	}{
		{
			name:   "no workflows",
			config: "command: test\n",
			errMsg: "no workflows",
		},
		{
			name: "no main",
			config: `
workflows:
  feature:
    - name: impl
      type: action
      prompt: impl.md
`,
			errMsg: "must have a 'main'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePipeline(tt.config)
			if err == nil {
				t.Fatal("expected error")
			}
			if !containsStr(err.Error(), tt.errMsg) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.errMsg)
			}
		})
	}
}
