# Grimoire

Grimoire is a skill discovery server that exposes reusable prompt workflows as MCP tools. LLM agents can search for relevant skills, retrieve step-by-step instructions, and then execute the suggested tools via [MCPBlade](https://github.com/flarexio/mcpblade) or other MCP servers.

## How It Works

```
LangChain Agent
    │
    ├── grimoire (knowledge layer)
    │   └── search_skills("deploy to k8s") → skill with prompt + suggested tools
    │
    └── mcpblade (execution layer)
        └── call suggested tools (kubectl_apply, docker_build, ...)
```

A **Skill** is a structured prompt template that tells the LLM *what to do* and *which tools to use*:

```yaml
id: deploy_k8s
name: Deploy to Kubernetes
description: Guide the deployment process to a Kubernetes cluster
tags: [kubernetes, deploy, k8s, container]
suggestedTools:
  - kubectl_apply
  - kubectl_get
  - docker_build
prompt: |
  Please follow these steps to deploy to Kubernetes:
  1. Use docker_build to build the container image
  2. Use kubectl_apply to apply the deployment manifest
  3. Use kubectl_get to verify the deployment status
```

## MCP Tools

Grimoire exposes three tools via the MCP protocol:

| Tool | Description |
|---|---|
| `search_skills` | Semantic search for skills using natural language queries |
| `find_skill` | Find a specific skill by name, returns full prompt and suggested tools |
| `list_skills` | List all available skills |

## Installation

```bash
go install github.com/flarexio/grimoire/cmd/grimoire@latest
```

## Configuration

Create `~/.flarex/grimoire/config.yaml`:

```yaml
skillsDir: skills
vector:
  enabled: true
  persistent: true
  collection: skills
```

Place skill definitions as YAML files in the `skills/` directory.

## Usage

```bash
# Run with default configuration (~/.flarex/grimoire)
grimoire

# Run with custom path
grimoire --path /path/to/config

# Custom HTTP address
grimoire --http-addr :9090
```

## Integration with MCPBlade

Register Grimoire as a backend MCP server in MCPBlade's `config.yaml`:

```yaml
mcpServers:
  grimoire:
    transport: streamable-http
    url: http://localhost:8080/mcp/
```

MCPBlade aggregates Grimoire's tools alongside other MCP servers, providing a unified interface for LLM agents.

## Architecture

```
grimoire/
├── skill/                  # Domain model
│   ├── skill.go            # Skill struct and errors
│   └── repository.go       # Repository interface
├── service.go              # Core business logic
├── endpoint.go             # Go-Kit endpoints
├── logging.go              # Logging middleware
├── model.go                # Config and vector document helpers
├── mcp/
│   └── endpoint.go         # MCP protocol endpoints
├── transport/http/         # HTTP transport layer
├── persistence/
│   ├── chromem/            # Vector DB (ChromeM) implementation
│   └── yaml/              # YAML file-based skill repository
└── cmd/grimoire/           # Entry point
```

## API Endpoints

```
GET  /api/skills            # List all skills
GET  /api/skills/search     # Search skills (?query=...&k=...)
GET  /api/skills/:name      # Find skill by name
POST /mcp/                  # MCP streamable HTTP endpoint
```

## Dependencies

- [mcp-go](https://github.com/mark3labs/mcp-go) — MCP protocol
- [chromem-go](https://github.com/philippgille/chromem-go) — Vector database
- [go-kit](https://github.com/go-kit/kit) — Microservice toolkit
- [gin](https://github.com/gin-gonic/gin) — HTTP framework
- [zap](https://github.com/uber-go/zap) — Structured logging

## License

MIT License — see [LICENSE.md](LICENSE.md) for details.

This project is part of the [FlareX](https://github.com/flarexio) ecosystem.
