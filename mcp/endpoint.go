package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/flarexio/grimoire"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      mcp.RequestId   `json:"id"`
	Method  mcp.MCPMethod   `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func errorResponse(id any, code int, message string) mcp.JSONRPCError {
	return mcp.JSONRPCError{
		JSONRPC: mcp.JSONRPC_VERSION,
		ID:      mcp.NewRequestId(id),
		Error: struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    any    `json:"data,omitempty"`
		}{
			Code:    code,
			Message: message,
		},
	}
}

type MCPEndpoint func(ctx context.Context, req JSONRPCRequest) mcp.JSONRPCMessage

const MCPSERVER_INSTRUCTIONS string = `Grimoire is a skill discovery server that helps you find the right skills (prompt workflows) for any task.

Available tools:
- search_skills: Find skills using natural language queries (semantic search)
- find_skill: Get the full details of a specific skill including its prompt and suggested tools
- list_skills: Browse all available skills, optionally filtered by category

Workflow:
1. Use search_skills to find relevant skills for your task
2. Use find_skill to retrieve the full prompt and suggested tools
3. Use the suggested tools (available via MCPBlade or other MCP servers) to execute the skill`

func InitializeEndpoint(svc grimoire.Service) MCPEndpoint {
	return func(ctx context.Context, req JSONRPCRequest) mcp.JSONRPCMessage {
		var params mcp.InitializeParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return errorResponse(req.ID, mcp.INVALID_PARAMS, err.Error())
		}

		protocolVersion := mcp.LATEST_PROTOCOL_VERSION
		if clientVersion := params.ProtocolVersion; clientVersion != "" {
			if slices.Contains(mcp.ValidProtocolVersions, clientVersion) {
				protocolVersion = clientVersion
			}
		}

		result := &mcp.InitializeResult{
			ProtocolVersion: protocolVersion,
			Capabilities: mcp.ServerCapabilities{
				Tools: &struct {
					ListChanged bool `json:"listChanged,omitempty"`
				}{},
			},
			ServerInfo: mcp.Implementation{
				Name:    "grimoire",
				Version: "1.0.0",
			},
			Instructions: MCPSERVER_INSTRUCTIONS,
		}

		return mcp.JSONRPCResponse{
			JSONRPC: mcp.JSONRPC_VERSION,
			ID:      req.ID,
			Result:  result,
		}
	}
}

func PingEndpoint(svc grimoire.Service) MCPEndpoint {
	return func(ctx context.Context, req JSONRPCRequest) mcp.JSONRPCMessage {
		return mcp.JSONRPCResponse{
			JSONRPC: mcp.JSONRPC_VERSION,
			ID:      req.ID,
			Result:  struct{}{},
		}
	}
}

func ListToolsEndpoint(svc grimoire.Service) MCPEndpoint {
	return func(ctx context.Context, req JSONRPCRequest) mcp.JSONRPCMessage {
		tools := []mcp.Tool{
			{
				Name:        "search_skills",
				Description: "Search for skills using natural language queries. Returns matching skills with their descriptions and suggested tools.",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]any{
						"query": map[string]any{
							"type":        "string",
							"description": "Natural language query to search for relevant skills",
						},
						"k": map[string]any{
							"type":        "integer",
							"description": "Maximum number of results to return (default: 5)",
						},
					},
					Required: []string{"query"},
				},
			},
			{
				Name:        "find_skill",
				Description: "Get the full details of a specific skill by its ID, including the prompt template and suggested tools.",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]any{
						"id": map[string]any{
							"type":        "string",
							"description": "The unique identifier of the skill",
						},
					},
					Required: []string{"id"},
				},
			},
			{
				Name:        "list_skills",
				Description: "List all available skills, optionally filtered by category.",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]any{
						"category": map[string]any{
							"type":        "string",
							"description": "Filter skills by category (optional)",
						},
					},
				},
			},
		}

		result := &mcp.ListToolsResult{
			Tools: tools,
		}

		return mcp.JSONRPCResponse{
			JSONRPC: mcp.JSONRPC_VERSION,
			ID:      req.ID,
			Result:  result,
		}
	}
}

func CallToolEndpoint(svc grimoire.Service) MCPEndpoint {
	return func(ctx context.Context, req JSONRPCRequest) mcp.JSONRPCMessage {
		var params mcp.CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return errorResponse(req.ID, mcp.INVALID_PARAMS, err.Error())
		}

		args, _ := params.Arguments.(map[string]any)
		if args == nil {
			args = make(map[string]any)
		}

		var result *mcp.CallToolResult

		switch params.Name {
		case "search_skills":
			result = handleSearchSkills(ctx, svc, args)

		case "find_skill":
			result = handleFindSkill(ctx, svc, args)

		case "list_skills":
			result = handleListSkills(ctx, svc, args)

		default:
			return errorResponse(req.ID, mcp.INVALID_PARAMS, "unknown tool: "+params.Name)
		}

		return mcp.JSONRPCResponse{
			JSONRPC: mcp.JSONRPC_VERSION,
			ID:      req.ID,
			Result:  result,
		}
	}
}

func handleSearchSkills(ctx context.Context, svc grimoire.Service, args map[string]any) *mcp.CallToolResult {
	query, _ := args["query"].(string)
	if query == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.NewTextContent("query parameter is required"),
			},
		}
	}

	var k int
	if v, ok := args["k"].(float64); ok {
		k = int(v)
	}

	skills, err := svc.SearchSkills(ctx, query, k)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.NewTextContent(err.Error()),
			},
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(formatSkillsList(skills)),
		},
	}
}

func handleFindSkill(ctx context.Context, svc grimoire.Service, args map[string]any) *mcp.CallToolResult {
	id, _ := args["id"].(string)
	if id == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.NewTextContent("id parameter is required"),
			},
		}
	}

	skill, err := svc.FindSkill(ctx, id)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.NewTextContent(err.Error()),
			},
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(formatSkillDetail(skill)),
		},
	}
}

func handleListSkills(ctx context.Context, svc grimoire.Service, args map[string]any) *mcp.CallToolResult {
	category, _ := args["category"].(string)

	skills, err := svc.ListSkills(ctx, category)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				mcp.NewTextContent(err.Error()),
			},
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewTextContent(formatSkillsList(skills)),
		},
	}
}

func formatSkillsList(skills []grimoire.Skill) string {
	var sb strings.Builder

	for i, skill := range skills {
		if i > 0 {
			sb.WriteString("\n---\n")
		}

		fmt.Fprintf(&sb, "ID: %s\n", skill.ID)
		fmt.Fprintf(&sb, "Name: %s\n", skill.Name)
		fmt.Fprintf(&sb, "Description: %s\n", skill.Description)
		fmt.Fprintf(&sb, "Category: %s\n", skill.Category)

		if len(skill.Tags) > 0 {
			fmt.Fprintf(&sb, "Tags: %s\n", strings.Join(skill.Tags, ", "))
		}

		if len(skill.SuggestedTools) > 0 {
			fmt.Fprintf(&sb, "Suggested Tools: %s\n", strings.Join(skill.SuggestedTools, ", "))
		}
	}

	return sb.String()
}

func formatSkillDetail(skill *grimoire.Skill) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "ID: %s\n", skill.ID)
	fmt.Fprintf(&sb, "Name: %s\n", skill.Name)
	fmt.Fprintf(&sb, "Description: %s\n", skill.Description)
	fmt.Fprintf(&sb, "Category: %s\n", skill.Category)

	if len(skill.Tags) > 0 {
		fmt.Fprintf(&sb, "Tags: %s\n", strings.Join(skill.Tags, ", "))
	}

	if len(skill.SuggestedTools) > 0 {
		fmt.Fprintf(&sb, "Suggested Tools: %s\n", strings.Join(skill.SuggestedTools, ", "))
	}

	fmt.Fprintf(&sb, "\n--- Prompt ---\n%s", skill.Prompt)

	return sb.String()
}
