package main

import "testing"

func TestValidateWorkflows(t *testing.T) {
	tests := []struct {
		name      string
		workflows map[string]*Workflow
		wantErr   string // substring of error, "" for no error
	}{
		{
			name:      "empty workflows",
			workflows: map[string]*Workflow{},
			wantErr:   "no workflows",
		},
		{
			name: "missing main",
			workflows: map[string]*Workflow{
				"feature": {Name: "feature", Nodes: []Node{
					{Name: "impl", Type: NodeAction, Prompt: "impl.md", MaxVisits: 1},
				}},
			},
			wantErr: "must have a 'main' workflow",
		},
		{
			name: "empty workflow",
			workflows: map[string]*Workflow{
				"main": {Name: "main", Nodes: []Node{}},
			},
			wantErr: "has no nodes",
		},
		{
			name: "duplicate node within workflow",
			workflows: map[string]*Workflow{
				"main": {Name: "main", Nodes: []Node{
					{Name: "triage", Type: NodeAction, Prompt: "a.md", MaxVisits: 1},
					{Name: "triage", Type: NodeAction, Prompt: "b.md", MaxVisits: 1},
				}},
			},
			wantErr: "duplicate node 'triage'",
		},
		{
			name: "node in multiple workflows",
			workflows: map[string]*Workflow{
				"main": {Name: "main", Nodes: []Node{
					{Name: "triage", Type: NodeAction, Prompt: "a.md", MaxVisits: 1},
				}},
				"feature": {Name: "feature", Nodes: []Node{
					{Name: "triage", Type: NodeAction, Prompt: "b.md", MaxVisits: 1},
				}},
			},
			wantErr: "appears in both",
		},
		{
			name: "node without prompt or run",
			workflows: map[string]*Workflow{
				"main": {Name: "main", Nodes: []Node{
					{Name: "empty", Type: NodeAction, MaxVisits: 1},
				}},
			},
			wantErr: "neither prompt nor run",
		},
		{
			name: "node with both prompt and run",
			workflows: map[string]*Workflow{
				"main": {Name: "main", Nodes: []Node{
					{Name: "both", Type: NodeAction, Prompt: "a.md", Run: "echo", MaxVisits: 1},
				}},
			},
			wantErr: "both prompt and run",
		},
		{
			name: "invalid node type",
			workflows: map[string]*Workflow{
				"main": {Name: "main", Nodes: []Node{
					{Name: "bad", Type: "magic", Prompt: "a.md", MaxVisits: 1},
				}},
			},
			wantErr: "invalid type",
		},
		{
			name: "routes on action node",
			workflows: map[string]*Workflow{
				"main": {Name: "main", Nodes: []Node{
					{Name: "impl", Type: NodeAction, Prompt: "a.md", Routes: []string{"feature"}, MaxVisits: 1},
				}},
				"feature": {Name: "feature", Nodes: []Node{
					{Name: "feat", Type: NodeAction, Prompt: "b.md", MaxVisits: 1},
				}},
			},
			wantErr: "not a decision node",
		},
		{
			name: "route to unknown workflow",
			workflows: map[string]*Workflow{
				"main": {Name: "main", Nodes: []Node{
					{Name: "triage", Type: NodeDecision, Prompt: "a.md", Routes: []string{"nonexistent"}, MaxVisits: 1},
				}},
			},
			wantErr: "unknown workflow 'nonexistent'",
		},
		{
			name: "zero max_visits",
			workflows: map[string]*Workflow{
				"main": {Name: "main", Nodes: []Node{
					{Name: "triage", Type: NodeAction, Prompt: "a.md", MaxVisits: 0},
				}},
			},
			wantErr: "invalid max_visits",
		},
		{
			name: "valid simple pipeline",
			workflows: map[string]*Workflow{
				"main": {Name: "main", Nodes: []Node{
					{Name: "triage", Type: NodeDecision, Prompt: "triage.md", Routes: []string{"feature"}, MaxVisits: 1},
				}},
				"feature": {Name: "feature", Nodes: []Node{
					{Name: "impl", Type: NodeAction, Prompt: "impl.md", MaxVisits: 1},
				}},
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkflows(tt.workflows)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !containsSubstring(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestResolveModel(t *testing.T) {
	p := &Pipeline{Model: "pipeline-model"}
	wfDefault := &Workflow{Name: "main"}
	wfOverride := &Workflow{Name: "breakdown", Model: "workflow-model"}

	nodeDefault := &Node{Name: "n1"}
	nodeOverride := &Node{Name: "n2", Model: "node-model"}

	// node > workflow > pipeline
	if got := resolveModel(p, wfDefault, nodeDefault); got != "pipeline-model" {
		t.Errorf("pipeline fallback: got %q", got)
	}
	if got := resolveModel(p, wfOverride, nodeDefault); got != "workflow-model" {
		t.Errorf("workflow override: got %q", got)
	}
	if got := resolveModel(p, wfOverride, nodeOverride); got != "node-model" {
		t.Errorf("node override: got %q", got)
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
