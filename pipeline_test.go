package main

import (
	"os"
	"path/filepath"
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

func TestParsePipelineNoteArtifact(t *testing.T) {
	config := `
workflows:
  main:
    - name: classify
      type: decision
      prompt: classify.md
      routes: [task, research]
  task:
    - name: implement
      type: action
      prompt: implement.md
    - name: review
      type: decision
      prompt: review.md
      note_artifact: summary.md
  research:
    - name: investigate
      type: action
      prompt: investigate.md
      note_artifact: findings.md
`
	p, err := ParsePipeline(config)
	if err != nil {
		t.Fatalf("ParsePipeline failed: %v", err)
	}

	// implement has no note_artifact
	implement := p.Workflows["task"].Nodes[0]
	if implement.NoteArtifact != "" {
		t.Errorf("implement.NoteArtifact = %q, want empty", implement.NoteArtifact)
	}

	// review has note_artifact
	review := p.Workflows["task"].Nodes[1]
	if review.NoteArtifact != "summary.md" {
		t.Errorf("review.NoteArtifact = %q, want %q", review.NoteArtifact, "summary.md")
	}

	// investigate has note_artifact
	investigate := p.Workflows["research"].Nodes[0]
	if investigate.NoteArtifact != "findings.md" {
		t.Errorf("investigate.NoteArtifact = %q, want %q", investigate.NoteArtifact, "findings.md")
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

func TestParseRequireCleanTree(t *testing.T) {
	dir := t.TempDir()

	// require_clean_tree: true sets the field
	configPath := filepath.Join(dir, "pipeline.yml")
	if err := os.WriteFile(configPath, []byte(`command: ./fake-llm
require_clean_tree: true
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
`), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	p, err := LoadPipeline(configPath)
	if err != nil {
		t.Fatalf("LoadPipeline failed: %v", err)
	}
	if !p.RequireCleanTree {
		t.Errorf("RequireCleanTree = false, want true")
	}

	// Omitting the field defaults to false
	if err := os.WriteFile(configPath, []byte(`command: ./fake-llm
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
`), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	p2, err := LoadPipeline(configPath)
	if err != nil {
		t.Fatalf("LoadPipeline failed: %v", err)
	}
	if p2.RequireCleanTree {
		t.Errorf("RequireCleanTree = true, want false when not specified")
	}
}

func TestParsePipelineAutoTriage(t *testing.T) {
	minimalWorkflow := `
workflows:
  main:
    - name: implement
      type: action
      run: echo done
`
	t.Run("auto_triage true", func(t *testing.T) {
		config := "command: fake-llm\nauto_triage: true\n" + minimalWorkflow
		p, err := ParsePipeline(config)
		if err != nil {
			t.Fatalf("ParsePipeline failed: %v", err)
		}
		if !p.AutoTriage {
			t.Errorf("AutoTriage = false, want true")
		}
	})

	t.Run("auto_triage false", func(t *testing.T) {
		config := "command: fake-llm\nauto_triage: false\n" + minimalWorkflow
		p, err := ParsePipeline(config)
		if err != nil {
			t.Fatalf("ParsePipeline failed: %v", err)
		}
		if p.AutoTriage {
			t.Errorf("AutoTriage = true, want false")
		}
	})

	t.Run("auto_triage absent defaults to false", func(t *testing.T) {
		config := "command: fake-llm\n" + minimalWorkflow
		p, err := ParsePipeline(config)
		if err != nil {
			t.Fatalf("ParsePipeline failed: %v", err)
		}
		if p.AutoTriage {
			t.Errorf("AutoTriage = true, want false when not specified")
		}
	})
}

func TestParsePipelineAutoAgent(t *testing.T) {
	minimalWorkflow := `
workflows:
  main:
    - name: implement
      type: action
      run: echo done
`
	t.Run("auto_agent true", func(t *testing.T) {
		config := "command: fake-llm\nauto_agent: true\n" + minimalWorkflow
		p, err := ParsePipeline(config)
		if err != nil {
			t.Fatalf("ParsePipeline failed: %v", err)
		}
		if !p.AutoAgent {
			t.Errorf("AutoAgent = false, want true")
		}
	})

	t.Run("auto_agent false", func(t *testing.T) {
		config := "command: fake-llm\nauto_agent: false\n" + minimalWorkflow
		p, err := ParsePipeline(config)
		if err != nil {
			t.Fatalf("ParsePipeline failed: %v", err)
		}
		if p.AutoAgent {
			t.Errorf("AutoAgent = true, want false")
		}
	})

	t.Run("auto_agent absent defaults to false", func(t *testing.T) {
		config := "command: fake-llm\n" + minimalWorkflow
		p, err := ParsePipeline(config)
		if err != nil {
			t.Fatalf("ParsePipeline failed: %v", err)
		}
		if p.AutoAgent {
			t.Errorf("AutoAgent = true, want false when not specified")
		}
	})
}

// --- from: directive tests ---

func TestExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"tilde only", "~", home},
		{"tilde slash path", "~/foo/bar", filepath.Join(home, "foo/bar")},
		{"absolute unchanged", "/abs/path", "/abs/path"},
		{"relative unchanged", "rel/path", "rel/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandTilde(tt.in)
			if err != nil {
				t.Fatalf("expandTilde(%q) error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("expandTilde(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParsePipelineFromRejected(t *testing.T) {
	config := `
from: /some/path
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
`
	_, err := ParsePipeline(config)
	if err == nil {
		t.Fatal("expected error for from: in legacy format")
	}
	if !containsStr(err.Error(), "only supported in unified config.yaml") {
		t.Errorf("error = %q, want mention of unified format", err.Error())
	}
}

// setupTemplateDir creates a template directory with pipeline.yaml and optional prompt files.
func setupTemplateDir(t *testing.T, pipelineYAML string, prompts map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pipeline.yaml"), []byte(pipelineYAML), 0644); err != nil {
		t.Fatalf("write pipeline.yaml: %v", err)
	}
	if len(prompts) > 0 {
		promptDir := filepath.Join(dir, "prompts")
		if err := os.MkdirAll(promptDir, 0755); err != nil {
			t.Fatalf("mkdir prompts: %v", err)
		}
		for name, content := range prompts {
			if err := os.WriteFile(filepath.Join(promptDir, name), []byte(content), 0644); err != nil {
				t.Fatalf("write prompt %s: %v", name, err)
			}
		}
	}
	return dir
}

func TestParseConfigWithFrom(t *testing.T) {
	templatePipeline := `
model: opus
max_retries: 5
discretion: high
workflows:
  main:
    - name: classify
      type: decision
      prompt: classify.md
      routes: [task]
  task:
    - name: implement
      type: action
      prompt: implement.md
`
	templateDir := setupTemplateDir(t, templatePipeline, nil)

	configYAML := "project:\n  prefix: myproj\n\npipeline:\n  from: " + templateDir + "\n  model: sonnet\n  on_succeed:\n    - git commit\n"

	c, err := ParseConfig(configYAML)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if c.Project.Prefix != "myproj" {
		t.Errorf("Project.Prefix = %q, want %q", c.Project.Prefix, "myproj")
	}

	// model should be overridden
	if c.Pipeline.Model != "sonnet" {
		t.Errorf("Pipeline.Model = %q, want %q (override)", c.Pipeline.Model, "sonnet")
	}

	// max_retries should be preserved from template (not clobbered by default 2)
	if c.Pipeline.MaxRetries != 5 {
		t.Errorf("Pipeline.MaxRetries = %d, want 5 (from template)", c.Pipeline.MaxRetries)
	}

	// discretion should be preserved from template
	if c.Pipeline.Discretion != "high" {
		t.Errorf("Pipeline.Discretion = %q, want %q (from template)", c.Pipeline.Discretion, "high")
	}

	// workflows should come from template
	if len(c.Pipeline.Workflows) != 2 {
		t.Fatalf("len(Workflows) = %d, want 2", len(c.Pipeline.Workflows))
	}
	if _, ok := c.Pipeline.Workflows["main"]; !ok {
		t.Error("main workflow missing")
	}
	if _, ok := c.Pipeline.Workflows["task"]; !ok {
		t.Error("task workflow missing")
	}

	// on_succeed should be from override
	if len(c.Pipeline.OnSucceed) != 1 || c.Pipeline.OnSucceed[0] != "git commit" {
		t.Errorf("OnSucceed = %v, want [git commit]", c.Pipeline.OnSucceed)
	}

	// TemplatePromptDir should be set
	want := filepath.Join(templateDir, "prompts")
	if c.Pipeline.TemplatePromptDir != want {
		t.Errorf("TemplatePromptDir = %q, want %q", c.Pipeline.TemplatePromptDir, want)
	}
}

func TestParseConfigFromNoOverrides(t *testing.T) {
	templatePipeline := `
model: opus
max_retries: 3
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
on_succeed:
  - echo done
`
	templateDir := setupTemplateDir(t, templatePipeline, nil)

	configYAML := "project:\n  prefix: bare\n\npipeline:\n  from: " + templateDir + "\n"

	c, err := ParseConfig(configYAML)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if c.Pipeline.Model != "opus" {
		t.Errorf("Pipeline.Model = %q, want %q", c.Pipeline.Model, "opus")
	}
	if c.Pipeline.MaxRetries != 3 {
		t.Errorf("Pipeline.MaxRetries = %d, want 3", c.Pipeline.MaxRetries)
	}
	if len(c.Pipeline.OnSucceed) != 1 || c.Pipeline.OnSucceed[0] != "echo done" {
		t.Errorf("OnSucceed = %v, want [echo done]", c.Pipeline.OnSucceed)
	}
}

func TestParseConfigFromListOverride(t *testing.T) {
	templatePipeline := `
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
on_succeed:
  - git add .
  - git commit
`
	templateDir := setupTemplateDir(t, templatePipeline, nil)

	configYAML := "pipeline:\n  from: " + templateDir + "\n  on_succeed:\n    - custom commit\n"

	c, err := ParseConfig(configYAML)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	// Override replaces, not appends
	if len(c.Pipeline.OnSucceed) != 1 || c.Pipeline.OnSucceed[0] != "custom commit" {
		t.Errorf("OnSucceed = %v, want [custom commit] (replace not append)", c.Pipeline.OnSucceed)
	}
}

func TestParseConfigFromWorkflowOverride(t *testing.T) {
	templatePipeline := `
workflows:
  main:
    - name: classify
      type: decision
      prompt: classify.md
      routes: [task, research]
  task:
    - name: implement
      type: action
      prompt: implement.md
  research:
    - name: investigate
      type: action
      prompt: investigate.md
`
	templateDir := setupTemplateDir(t, templatePipeline, nil)

	// Override with completely different workflows
	configYAML := "pipeline:\n  from: " + templateDir + "\n  workflows:\n    main:\n      - name: do_it\n        type: action\n        prompt: do_it.md\n"

	c, err := ParseConfig(configYAML)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	// Should have only the override's workflows, not template's
	if len(c.Pipeline.Workflows) != 1 {
		t.Fatalf("len(Workflows) = %d, want 1 (override replaces entirely)", len(c.Pipeline.Workflows))
	}
	if _, ok := c.Pipeline.Workflows["main"]; !ok {
		t.Error("main workflow missing")
	}
	if c.Pipeline.Workflows["main"].Nodes[0].Name != "do_it" {
		t.Errorf("main node name = %q, want %q", c.Pipeline.Workflows["main"].Nodes[0].Name, "do_it")
	}
}

func TestParseConfigFromRecursiveError(t *testing.T) {
	templatePipeline := `
from: /some/other/path
workflows:
  main:
    - name: impl
      type: action
      prompt: impl.md
`
	templateDir := setupTemplateDir(t, templatePipeline, nil)

	configYAML := "pipeline:\n  from: " + templateDir + "\n"

	_, err := ParseConfig(configYAML)
	if err == nil {
		t.Fatal("expected error for recursive from:")
	}
	if !containsStr(err.Error(), "cannot itself contain a from:") {
		t.Errorf("error = %q, want mention of recursive from:", err.Error())
	}
}

func TestParseConfigFromRelativePathError(t *testing.T) {
	configYAML := `
pipeline:
  from: ./relative/path
  workflows:
    main:
      - name: impl
        type: action
        prompt: impl.md
`
	_, err := ParseConfig(configYAML)
	if err == nil {
		t.Fatal("expected error for relative from: path")
	}
	if !containsStr(err.Error(), "must be absolute or tilde-prefixed") {
		t.Errorf("error = %q, want mention of absolute path", err.Error())
	}
}

func TestParseConfigFromMissingError(t *testing.T) {
	configYAML := `
pipeline:
  from: /nonexistent/path/to/template
`
	_, err := ParseConfig(configYAML)
	if err == nil {
		t.Fatal("expected error for missing template dir")
	}
	if !containsStr(err.Error(), "template pipeline not found") {
		t.Errorf("error = %q, want mention of template not found", err.Error())
	}
}

func TestMergePipeline(t *testing.T) {
	base := &Pipeline{
		Agent:      "claude",
		Model:      "opus",
		MaxRetries: 5,
		MaxDepth:   3,
		Discretion: "high",
		Workflows: map[string]*Workflow{
			"main": {Name: "main", Nodes: []Node{{Name: "impl", Type: NodeAction}}},
		},
		OnSucceed:         []string{"git add ."},
		OnFail:            []string{"echo fail"},
		TemplatePromptDir: "/template/prompts",
	}

	override := &Pipeline{
		Agent:      "claude", // default — should NOT override since setFields won't include it
		Model:      "sonnet",
		MaxRetries: 2, // default — should NOT override
		setFields:  map[string]bool{"model": true, "on_succeed": true},
		OnSucceed:  []string{"git commit -m 'done'"},
	}

	result := MergePipeline(base, override)

	if result.Model != "sonnet" {
		t.Errorf("Model = %q, want %q (overridden)", result.Model, "sonnet")
	}
	if result.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5 (base preserved, override was default)", result.MaxRetries)
	}
	if result.MaxDepth != 3 {
		t.Errorf("MaxDepth = %d, want 3 (base preserved)", result.MaxDepth)
	}
	if result.Discretion != "high" {
		t.Errorf("Discretion = %q, want %q (base preserved)", result.Discretion, "high")
	}
	if len(result.OnSucceed) != 1 || result.OnSucceed[0] != "git commit -m 'done'" {
		t.Errorf("OnSucceed = %v, want [git commit -m 'done'] (overridden)", result.OnSucceed)
	}
	if len(result.OnFail) != 1 || result.OnFail[0] != "echo fail" {
		t.Errorf("OnFail = %v, want [echo fail] (base preserved)", result.OnFail)
	}
	if result.TemplatePromptDir != "/template/prompts" {
		t.Errorf("TemplatePromptDir = %q, want %q (always from base)", result.TemplatePromptDir, "/template/prompts")
	}

	// Workflows should be from base (not overridden)
	if len(result.Workflows) != 1 {
		t.Fatalf("len(Workflows) = %d, want 1", len(result.Workflows))
	}
	if result.Workflows["main"].Nodes[0].Name != "impl" {
		t.Errorf("main node = %q, want %q", result.Workflows["main"].Nodes[0].Name, "impl")
	}

	// Mutating result's base-sourced slices shouldn't affect original
	result.OnFail[0] = "modified"
	if base.OnFail[0] != "echo fail" {
		t.Error("MergePipeline aliased base slices — mutation leaked")
	}
}

func TestLoadPromptFileFallback(t *testing.T) {
	// Set up project dir with .ko/prompts/ and a template prompts dir
	projectDir := t.TempDir()
	ticketsDir := filepath.Join(projectDir, ".ko", "tickets")
	localPrompts := filepath.Join(projectDir, ".ko", "prompts")
	if err := os.MkdirAll(ticketsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(localPrompts, 0755); err != nil {
		t.Fatal(err)
	}

	templatePrompts := filepath.Join(t.TempDir(), "prompts")
	if err := os.MkdirAll(templatePrompts, 0755); err != nil {
		t.Fatal(err)
	}

	// Write plan.md only in template
	if err := os.WriteFile(filepath.Join(templatePrompts, "plan.md"), []byte("template plan"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should find template version
	content, err := LoadPromptFile(ticketsDir, "plan.md", templatePrompts)
	if err != nil {
		t.Fatalf("LoadPromptFile fallback failed: %v", err)
	}
	if content != "template plan" {
		t.Errorf("content = %q, want %q", content, "template plan")
	}

	// Write local override
	if err := os.WriteFile(filepath.Join(localPrompts, "plan.md"), []byte("local plan"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should prefer local version
	content, err = LoadPromptFile(ticketsDir, "plan.md", templatePrompts)
	if err != nil {
		t.Fatalf("LoadPromptFile local override failed: %v", err)
	}
	if content != "local plan" {
		t.Errorf("content = %q, want %q (local should override template)", content, "local plan")
	}

	// Missing in both should error with both paths mentioned
	_, err = LoadPromptFile(ticketsDir, "missing.md", templatePrompts)
	if err == nil {
		t.Fatal("expected error for missing prompt")
	}
	if !containsStr(err.Error(), "template") {
		t.Errorf("error = %q, want mention of template path", err.Error())
	}

	// No template dir — should error with simple message
	_, err = LoadPromptFile(ticketsDir, "missing.md", "")
	if err == nil {
		t.Fatal("expected error for missing prompt without template")
	}
	if containsStr(err.Error(), "template") {
		t.Errorf("error = %q, should not mention template when no template dir", err.Error())
	}
}
