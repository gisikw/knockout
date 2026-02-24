package main

import "fmt"

// NodeType distinguishes decision nodes from action nodes.
type NodeType string

const (
	NodeDecision NodeType = "decision"
	NodeAction   NodeType = "action"
)

// Node represents a single step in a workflow.
type Node struct {
	Name         string   // node identifier (unique within workflow)
	Type         NodeType // decision or action
	Prompt       string   // prompt file reference (mutually exclusive with Run)
	Run          string   // shell command (mutually exclusive with Prompt)
	Model        string   // optional model override
	AllowAll     *bool    // per-node allow_all_tool_calls override (nil = inherit)
	AllowedTools []string // per-node allowed_tools override (nil = inherit)
	Routes       []string // workflows this decision node can route to
	MaxVisits    int      // max times this node can be entered per build (default: 1)
	Timeout      string   // optional timeout override (e.g., "5m", "1h30m")
	Skills       []string // skill directories to make available (future multi-agent harness support)
	Skill        string   // specific skill to invoke (future multi-agent harness support; mutually exclusive with Prompt/Run)
}

// IsPromptNode reports whether this node invokes an LLM.
func (n *Node) IsPromptNode() bool {
	return n.Prompt != ""
}

// IsRunNode reports whether this node runs a shell command.
func (n *Node) IsRunNode() bool {
	return n.Run != ""
}

// Workflow is a named sequence of nodes.
type Workflow struct {
	Name         string   // workflow identifier
	Model        string   // optional model override for all nodes in this workflow
	AllowAll     *bool    // per-workflow allow_all_tool_calls override (nil = inherit)
	AllowedTools []string // per-workflow allowed_tools override (nil = inherit)
	Nodes        []Node   // ordered list of nodes
}

// ValidateWorkflows checks the workflow graph for structural errors.
// Pure decision function — returns nil if valid.
func ValidateWorkflows(workflows map[string]*Workflow) error {
	if len(workflows) == 0 {
		return fmt.Errorf("pipeline has no workflows")
	}

	// Must have a "main" workflow — entry point for all tickets.
	if _, ok := workflows["main"]; !ok {
		return fmt.Errorf("pipeline must have a 'main' workflow")
	}

	// Collect all node names across all workflows for uniqueness check.
	nodeOwner := make(map[string]string) // node name -> workflow name

	for wfName, wf := range workflows {
		if len(wf.Nodes) == 0 {
			return fmt.Errorf("workflow '%s' has no nodes", wfName)
		}

		seen := make(map[string]bool)
		for _, node := range wf.Nodes {
			// Node name unique within workflow
			if seen[node.Name] {
				return fmt.Errorf("workflow '%s' has duplicate node '%s'", wfName, node.Name)
			}
			seen[node.Name] = true

			// Node name unique across all workflows
			if owner, exists := nodeOwner[node.Name]; exists {
				return fmt.Errorf("node '%s' appears in both workflow '%s' and '%s'", node.Name, owner, wfName)
			}
			nodeOwner[node.Name] = wfName

			// Must have prompt, run, or skill (exactly one)
			hasPrompt := node.Prompt != ""
			hasRun := node.Run != ""
			hasSkill := node.Skill != ""

			if !hasPrompt && !hasRun && !hasSkill {
				return fmt.Errorf("node '%s' in workflow '%s' has neither prompt, run, nor skill", node.Name, wfName)
			}
			if hasPrompt && hasRun {
				return fmt.Errorf("node '%s' in workflow '%s' has both prompt and run", node.Name, wfName)
			}
			if hasPrompt && hasSkill {
				return fmt.Errorf("node '%s' in workflow '%s' has both prompt and skill", node.Name, wfName)
			}
			if hasRun && hasSkill {
				return fmt.Errorf("node '%s' in workflow '%s' has both run and skill", node.Name, wfName)
			}

			// Valid node type
			if node.Type != NodeDecision && node.Type != NodeAction {
				return fmt.Errorf("node '%s' in workflow '%s' has invalid type '%s'", node.Name, wfName, node.Type)
			}

			// Routes only valid on decision nodes
			if len(node.Routes) > 0 && node.Type != NodeDecision {
				return fmt.Errorf("node '%s' in workflow '%s' declares routes but is not a decision node", node.Name, wfName)
			}

			// All route targets must exist
			for _, target := range node.Routes {
				if _, ok := workflows[target]; !ok {
					return fmt.Errorf("node '%s' in workflow '%s' routes to unknown workflow '%s'", node.Name, wfName, target)
				}
			}

			// max_visits must be positive
			if node.MaxVisits < 1 {
				return fmt.Errorf("node '%s' in workflow '%s' has invalid max_visits %d", node.Name, wfName, node.MaxVisits)
			}
		}
	}

	return nil
}
