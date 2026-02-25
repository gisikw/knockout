package main

import (
	"os"
	"testing"
	"time"
)

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
on_fail:
  - git checkout -- .
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
	if len(p.OnFail) != 1 || p.OnFail[0] != "git checkout -- ." {
		t.Errorf("OnFail = %v", p.OnFail)
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

func TestParsePipelinePerNodeAllowAll(t *testing.T) {
	config := `
allow_all_tool_calls: true
workflows:
  main:
    - name: triage
      type: decision
      prompt: triage.md
      allow_all_tool_calls: false
      routes: [feature]
  feature:
    allow_all_tool_calls: false
    - name: implement
      type: action
      prompt: implement.md
    - name: review
      type: action
      prompt: review.md
      allow_all_tool_calls: true
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}

	if !p.AllowAll {
		t.Error("pipeline AllowAll = false, want true")
	}

	// Node-level override: triage explicitly false
	triage := p.Workflows["main"].Nodes[0]
	if triage.AllowAll == nil || *triage.AllowAll != false {
		t.Errorf("triage.AllowAll = %v, want false", triage.AllowAll)
	}
	if resolveAllowAll(p, p.Workflows["main"], &triage) != false {
		t.Error("resolveAllowAll(triage) = true, want false")
	}

	// Workflow-level override: feature workflow false
	featureWF := p.Workflows["feature"]
	if featureWF.AllowAll == nil || *featureWF.AllowAll != false {
		t.Errorf("feature.AllowAll = %v, want false", featureWF.AllowAll)
	}

	// implement inherits from workflow (false)
	implement := featureWF.Nodes[0]
	if implement.AllowAll != nil {
		t.Errorf("implement.AllowAll = %v, want nil (inherit)", implement.AllowAll)
	}
	if resolveAllowAll(p, featureWF, &implement) != false {
		t.Error("resolveAllowAll(implement) = true, want false (inherits from workflow)")
	}

	// review has node-level override to true
	review := featureWF.Nodes[1]
	if review.AllowAll == nil || *review.AllowAll != true {
		t.Errorf("review.AllowAll = %v, want true", review.AllowAll)
	}
	if resolveAllowAll(p, featureWF, &review) != true {
		t.Error("resolveAllowAll(review) = false, want true (node overrides workflow)")
	}
}

func TestResolveAllowAllInheritsPipeline(t *testing.T) {
	config := `
allow_all_tool_calls: true
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

	wf := p.Workflows["main"]
	node := wf.Nodes[0]
	if resolveAllowAll(p, wf, &node) != true {
		t.Error("resolveAllowAll should inherit pipeline-level true")
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

func TestParsePipelineSkillsMultiline(t *testing.T) {
	config := `
workflows:
  main:
    - name: task
      type: action
      skill: feature-dev
      skills:
        - .claude/commands
        - ~/my-skills
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	node := p.Workflows["main"].Nodes[0]
	if node.Skill != "feature-dev" {
		t.Errorf("Skill = %q, want %q", node.Skill, "feature-dev")
	}
	if len(node.Skills) != 2 {
		t.Fatalf("len(Skills) = %d, want 2", len(node.Skills))
	}
	if node.Skills[0] != ".claude/commands" {
		t.Errorf("Skills[0] = %q, want %q", node.Skills[0], ".claude/commands")
	}
	if node.Skills[1] != "~/my-skills" {
		t.Errorf("Skills[1] = %q, want %q", node.Skills[1], "~/my-skills")
	}
}

func TestParsePipelineSkillsInline(t *testing.T) {
	config := `
workflows:
  main:
    - name: task
      type: action
      skill: feature-dev
      skills: [.claude/commands, ~/my-skills]
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	node := p.Workflows["main"].Nodes[0]
	if node.Skill != "feature-dev" {
		t.Errorf("Skill = %q, want %q", node.Skill, "feature-dev")
	}
	if len(node.Skills) != 2 {
		t.Fatalf("len(Skills) = %d, want 2", len(node.Skills))
	}
	if node.Skills[0] != ".claude/commands" {
		t.Errorf("Skills[0] = %q", node.Skills[0])
	}
	if node.Skills[1] != "~/my-skills" {
		t.Errorf("Skills[1] = %q", node.Skills[1])
	}
}

func TestValidateWorkflowsSkillExclusivity(t *testing.T) {
	tests := []struct {
		name   string
		config string
		errMsg string
	}{
		{
			name: "skill and prompt",
			config: `
workflows:
  main:
    - name: task
      type: action
      skill: feature-dev
      prompt: impl.md
`,
			errMsg: "both prompt and skill",
		},
		{
			name: "skill and run",
			config: `
workflows:
  main:
    - name: task
      type: action
      skill: feature-dev
      run: just test
`,
			errMsg: "both run and skill",
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

func TestParseTimeout(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     time.Duration
		wantErr  bool
	}{
		{
			name:  "empty string returns 15 minutes",
			input: "",
			want:  15 * time.Minute,
		},
		{
			name:  "5 minutes",
			input: "5m",
			want:  5 * time.Minute,
		},
		{
			name:  "1 hour 30 minutes",
			input: "1h30m",
			want:  90 * time.Minute,
		},
		{
			name:  "2 hours",
			input: "2h",
			want:  2 * time.Hour,
		},
		{
			name:  "30 seconds",
			input: "30s",
			want:  30 * time.Second,
		},
		{
			name:    "invalid format",
			input:   "not-a-duration",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeout(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseTimeout(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParsePipelineStepTimeout(t *testing.T) {
	config := `
step_timeout: 20m
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
	if p.StepTimeout != "20m" {
		t.Errorf("StepTimeout = %q, want %q", p.StepTimeout, "20m")
	}
}

func TestParsePipelineNodeTimeout(t *testing.T) {
	config := `
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
      timeout: 5m
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}
	node := p.Workflows["main"].Nodes[0]
	if node.Timeout != "5m" {
		t.Errorf("Node.Timeout = %q, want %q", node.Timeout, "5m")
	}
}

func TestParsePipelineAllowedTools(t *testing.T) {
	config := `
allowed_tools:
  - Read
  - Write
workflows:
  main:
    allowed_tools:
      - Read
      - Bash
    - name: task1
      type: action
      prompt: impl.md
      allowed_tools:
        - Read
        - Write
        - Edit
    - name: task2
      type: action
      prompt: test.md
      allowed_tools: [Bash, Grep]
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}

	// Pipeline-level
	if len(p.AllowedTools) != 2 {
		t.Fatalf("len(Pipeline.AllowedTools) = %d, want 2", len(p.AllowedTools))
	}
	if p.AllowedTools[0] != "Read" {
		t.Errorf("Pipeline.AllowedTools[0] = %q, want %q", p.AllowedTools[0], "Read")
	}
	if p.AllowedTools[1] != "Write" {
		t.Errorf("Pipeline.AllowedTools[1] = %q, want %q", p.AllowedTools[1], "Write")
	}

	// Workflow-level
	wf := p.Workflows["main"]
	if len(wf.AllowedTools) != 2 {
		t.Fatalf("len(Workflow.AllowedTools) = %d, want 2", len(wf.AllowedTools))
	}
	if wf.AllowedTools[0] != "Read" {
		t.Errorf("Workflow.AllowedTools[0] = %q, want %q", wf.AllowedTools[0], "Read")
	}
	if wf.AllowedTools[1] != "Bash" {
		t.Errorf("Workflow.AllowedTools[1] = %q, want %q", wf.AllowedTools[1], "Bash")
	}

	// Node-level multiline
	node1 := wf.Nodes[0]
	if len(node1.AllowedTools) != 3 {
		t.Fatalf("len(Node1.AllowedTools) = %d, want 3", len(node1.AllowedTools))
	}
	if node1.AllowedTools[0] != "Read" {
		t.Errorf("Node1.AllowedTools[0] = %q, want %q", node1.AllowedTools[0], "Read")
	}
	if node1.AllowedTools[1] != "Write" {
		t.Errorf("Node1.AllowedTools[1] = %q, want %q", node1.AllowedTools[1], "Write")
	}
	if node1.AllowedTools[2] != "Edit" {
		t.Errorf("Node1.AllowedTools[2] = %q, want %q", node1.AllowedTools[2], "Edit")
	}

	// Node-level inline
	node2 := wf.Nodes[1]
	if len(node2.AllowedTools) != 2 {
		t.Fatalf("len(Node2.AllowedTools) = %d, want 2", len(node2.AllowedTools))
	}
	if node2.AllowedTools[0] != "Bash" {
		t.Errorf("Node2.AllowedTools[0] = %q, want %q", node2.AllowedTools[0], "Bash")
	}
	if node2.AllowedTools[1] != "Grep" {
		t.Errorf("Node2.AllowedTools[1] = %q, want %q", node2.AllowedTools[1], "Grep")
	}
}

func TestResolveAllowedToolsOverride(t *testing.T) {
	config := `
allowed_tools:
  - Read
  - Write
workflows:
  main:
    allowed_tools:
      - Bash
      - Grep
    - name: inherit_workflow
      type: action
      prompt: impl.md
    - name: override_node
      type: action
      prompt: impl.md
      allowed_tools:
        - Edit
        - TodoWrite
  other:
    - name: inherit_pipeline
      type: action
      prompt: impl.md
    - name: empty_override
      type: action
      prompt: impl.md
      allowed_tools: []
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}

	// Test node > workflow > pipeline precedence
	mainWF := p.Workflows["main"]

	// Node inherits from workflow
	inheritWorkflowNode := mainWF.Nodes[0]
	resolved := resolveAllowedTools(p, mainWF, &inheritWorkflowNode)
	if len(resolved) != 2 || resolved[0] != "Bash" || resolved[1] != "Grep" {
		t.Errorf("node inheriting from workflow: got %v, want [Bash Grep]", resolved)
	}

	// Node overrides workflow
	overrideNode := mainWF.Nodes[1]
	resolved = resolveAllowedTools(p, mainWF, &overrideNode)
	if len(resolved) != 2 || resolved[0] != "Edit" || resolved[1] != "TodoWrite" {
		t.Errorf("node overriding workflow: got %v, want [Edit TodoWrite]", resolved)
	}

	// Workflow inherits from pipeline
	otherWF := p.Workflows["other"]
	inheritPipelineNode := otherWF.Nodes[0]
	resolved = resolveAllowedTools(p, otherWF, &inheritPipelineNode)
	if len(resolved) != 2 || resolved[0] != "Read" || resolved[1] != "Write" {
		t.Errorf("workflow inheriting from pipeline: got %v, want [Read Write]", resolved)
	}

	// Node with empty list overrides parent
	emptyNode := otherWF.Nodes[1]
	resolved = resolveAllowedTools(p, otherWF, &emptyNode)
	if len(resolved) != 0 {
		t.Errorf("node with empty list: got %v, want []", resolved)
	}
}

func TestParsePipelineWorkflowOnSuccess(t *testing.T) {
	config := `
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
  research:
    on_success: resolved
    - name: investigate
      type: action
      prompt: investigate.md
  task:
    on_success: closed
    - name: execute
      type: action
      prompt: execute.md
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}

	// main workflow should have empty on_success (default)
	main := p.Workflows["main"]
	if main.OnSuccess != "" {
		t.Errorf("main.OnSuccess = %q, want empty (default)", main.OnSuccess)
	}

	// research workflow should have on_success: resolved
	research := p.Workflows["research"]
	if research.OnSuccess != "resolved" {
		t.Errorf("research.OnSuccess = %q, want %q", research.OnSuccess, "resolved")
	}

	// task workflow should have on_success: closed
	task := p.Workflows["task"]
	if task.OnSuccess != "closed" {
		t.Errorf("task.OnSuccess = %q, want %q", task.OnSuccess, "closed")
	}
}

func TestParseConfigUnified(t *testing.T) {
	configYAML := `
project:
  prefix: myproj

pipeline:
  model: sonnet
  max_retries: 1
  max_depth: 3
  discretion: high
  workflows:
    main:
      - name: triage
        type: decision
        prompt: triage.md
        routes: [task]
    task:
      - name: implement
        type: action
        prompt: implement.md
  on_succeed:
    - git commit
`

	c, err := ParseConfig(configYAML)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	// Project settings
	if c.Project.Prefix != "myproj" {
		t.Errorf("Project.Prefix = %q, want %q", c.Project.Prefix, "myproj")
	}

	// Pipeline settings
	if c.Pipeline.Model != "sonnet" {
		t.Errorf("Pipeline.Model = %q, want %q", c.Pipeline.Model, "sonnet")
	}
	if c.Pipeline.MaxRetries != 1 {
		t.Errorf("Pipeline.MaxRetries = %d, want %d", c.Pipeline.MaxRetries, 1)
	}
	if c.Pipeline.MaxDepth != 3 {
		t.Errorf("Pipeline.MaxDepth = %d, want %d", c.Pipeline.MaxDepth, 3)
	}
	if c.Pipeline.Discretion != "high" {
		t.Errorf("Pipeline.Discretion = %q, want %q", c.Pipeline.Discretion, "high")
	}

	// Workflows
	if len(c.Pipeline.Workflows) != 2 {
		t.Fatalf("len(Workflows) = %d, want 2", len(c.Pipeline.Workflows))
	}
	if _, ok := c.Pipeline.Workflows["main"]; !ok {
		t.Error("main workflow missing")
	}
	if _, ok := c.Pipeline.Workflows["task"]; !ok {
		t.Error("task workflow missing")
	}

	// Hooks
	if len(c.Pipeline.OnSucceed) != 1 || c.Pipeline.OnSucceed[0] != "git commit" {
		t.Errorf("OnSucceed = %v, want [git commit]", c.Pipeline.OnSucceed)
	}
}

func TestParseConfigWithInlineComment(t *testing.T) {
	configYAML := `
project:
  prefix: ko  # This is the project prefix

pipeline:
  model: sonnet
  workflows:
    main:
      - name: impl
        type: action
        prompt: impl.md
`

	c, err := ParseConfig(configYAML)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	// Inline comment should be stripped from prefix
	if c.Project.Prefix != "ko" {
		t.Errorf("Project.Prefix = %q, want %q (inline comment should be stripped)", c.Project.Prefix, "ko")
	}
}

func TestLoadConfigUnified(t *testing.T) {
	// Create a temp file with unified config
	dir := t.TempDir()
	configPath := dir + "/config.yaml"
	configYAML := `project:
  prefix: testproj

pipeline:
  model: opus
  workflows:
    main:
      - name: impl
        type: action
        prompt: impl.md
`
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	c, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if c.Project.Prefix != "testproj" {
		t.Errorf("Project.Prefix = %q, want %q", c.Project.Prefix, "testproj")
	}
	if c.Pipeline.Model != "opus" {
		t.Errorf("Pipeline.Model = %q, want %q", c.Pipeline.Model, "opus")
	}
}

func TestLoadConfigLegacyPipeline(t *testing.T) {
	// Create a temp file with legacy pipeline.yml format
	dir := t.TempDir()
	configPath := dir + "/pipeline.yml"
	pipelineYAML := `model: haiku
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
`
	if err := os.WriteFile(configPath, []byte(pipelineYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	c, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Legacy format should have empty project.prefix
	if c.Project.Prefix != "" {
		t.Errorf("Project.Prefix = %q, want empty for legacy format", c.Project.Prefix)
	}
	if c.Pipeline.Model != "haiku" {
		t.Errorf("Pipeline.Model = %q, want %q", c.Pipeline.Model, "haiku")
	}
}

func TestLoadPipelineBackwardsCompat(t *testing.T) {
	// LoadPipeline should still work for backwards compatibility
	dir := t.TempDir()
	configPath := dir + "/pipeline.yml"
	pipelineYAML := `model: sonnet
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
`
	if err := os.WriteFile(configPath, []byte(pipelineYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	p, err := LoadPipeline(configPath)
	if err != nil {
		t.Fatalf("LoadPipeline failed: %v", err)
	}

	if p.Model != "sonnet" {
		t.Errorf("Model = %q, want %q", p.Model, "sonnet")
	}
}
