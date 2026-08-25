package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) registerPrompts() {
	huntPrompt := mcp.NewPrompt(
		"cybermes_hunt",
		mcp.WithPromptDescription("Initialize an autonomous security research or bug bounty testing session aligned with Cybermes AGENTS.md operational directives."),
		mcp.WithArgument(
			"target",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Target domain, URL, or host (e.g. 'api.target.com', 'https://staging.target.com')."),
		),
		mcp.WithArgument(
			"scope_notes",
			mcp.ArgumentDescription("Allowed endpoints, out-of-scope assets, test accounts, or specific engagement rules."),
		),
		mcp.WithArgument(
			"focus_area",
			mcp.ArgumentDescription("Focus area (e.g. 'idor', 'jwt_auth', 'api_surface', 'business_logic', 'recon')."),
		),
	)

	triagePrompt := mcp.NewPrompt(
		"cybermes_triage",
		mcp.WithPromptDescription("Validate an observed anomaly or potential vulnerability against Cybermes Zero-False-Positive standards before recording."),
		mcp.WithArgument(
			"target",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Target slug or identifier."),
		),
		mcp.WithArgument(
			"vulnerability_type",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Type of vulnerability (e.g. 'BOLA / IDOR', 'Authentication Bypass', 'Race Condition')."),
		),
		mcp.WithArgument(
			"raw_observation",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Observed HTTP response, differential status code, or timing evidence."),
		),
	)

	s.mcpServer.AddPrompt(huntPrompt, s.handleHuntPrompt)
	s.mcpServer.AddPrompt(triagePrompt, s.handleTriagePrompt)
}

func (s *Server) handleHuntPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	target := request.Params.Arguments["target"]
	if strings.TrimSpace(target) == "" {
		return nil, fmt.Errorf("missing required argument 'target'")
	}
	scopeNotes := request.Params.Arguments["scope_notes"]
	focusArea := request.Params.Arguments["focus_area"]

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 🛡️ Cybermes Security Research Directive: `%s`\n\n", target))
	sb.WriteString("You are acting as **Cybermes**, an autonomous offensive security research agent.\n\n")
	sb.WriteString("### Operational Constraints & Principles:\n")
	sb.WriteString("1. **Direct Authorization**: Target is explicitly authorized. Proceed directly with active reconnaissance and testing.\n")
	sb.WriteString("2. **Non-Destructive Execution**: Safe rate limits (max 5-10 req/s), no volumetric flooding, no data destruction.\n")
	sb.WriteString("3. **Zero-False-Positive Gate**: Never declare a vulnerability confirmed without reproducible raw HTTP evidence, differential proof, or standalone PoC script.\n")
	sb.WriteString("4. **Structured Workspace**: Save confirmed findings to `reports/<target>/findings/<severity>_<vuln>.md`.\n\n")

	if scopeNotes != "" {
		sb.WriteString(fmt.Sprintf("### Target Scope Notes:\n%s\n\n", scopeNotes))
	}
	if focusArea != "" {
		sb.WriteString(fmt.Sprintf("### Priority Testing Focus:\n`%s` - Use `cybermes_search_knowledge` and `cybermes_get_skill` to retrieve exact playbooks.\n\n", focusArea))
	}

	sb.WriteString("### Initial Recommended Actions:\n")
	sb.WriteString("1. List relevant playbooks with `cybermes_list_skills` (e.g. `cybermes_list_skills(filter=\"api\")`).\n")
	sb.WriteString("2. Search exploit payloads with `cybermes_search_knowledge(query=\"...\")`.\n")
	sb.WriteString("3. Perform systematic endpoint and authorization analysis.\n")

	messages := []mcp.PromptMessage{
		mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(sb.String())),
	}

	return mcp.NewGetPromptResult("Cybermes Hunting Workflow", messages), nil
}

func (s *Server) handleTriagePrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	target := request.Params.Arguments["target"]
	vulnType := request.Params.Arguments["vulnerability_type"]
	rawObs := request.Params.Arguments["raw_observation"]

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 🔍 Cybermes Zero-False-Positive Triage Checklist\n\n"))
	sb.WriteString(fmt.Sprintf("- **Target**: `%s`\n", target))
	sb.WriteString(fmt.Sprintf("- **Hypothesis**: `%s`\n", vulnType))
	sb.WriteString(fmt.Sprintf("- **Observation**: \n```\n%s\n```\n\n", rawObs))
	sb.WriteString("### Mandatory Verification Steps Before Confirmation:\n")
	sb.WriteString("1. **Differential Validation**: Is there a clear control test? (e.g. User A accessing User B's object vs User A accessing User A's object).\n")
	sb.WriteString("2. **Authentication Verification**: Does unauthenticated request return 401/403 or same data?\n")
	sb.WriteString("3. **State Change Impact**: Is data actually mutated or leaked, or is this a superficial 200 OK with empty response?\n")
	sb.WriteString("4. **PoC Script**: Can this be reproduced deterministically with a minimal Python script using `requests`?\n\n")
	sb.WriteString("If all steps pass, record the finding using `cybermes_record_finding`.")

	messages := []mcp.PromptMessage{
		mcp.NewPromptMessage(mcp.RoleUser, mcp.NewTextContent(sb.String())),
	}

	return mcp.NewGetPromptResult("Cybermes Vulnerability Triage", messages), nil
}
