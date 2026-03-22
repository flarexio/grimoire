# Transport Layer Review Guidelines

## HTTP
- Use appropriate HTTP status codes (400 for bad input, 404 for not found, 500 for server errors)
- Validate query parameters and path parameters before calling endpoints
- Do not expose raw error messages to clients; map domain errors to user-safe responses

## MCP
- Validate JSON-RPC request structure and required parameters
- Return proper JSON-RPC error codes per MCP specification
- Ensure MCP tool schemas match actual parameter handling in CallToolEndpoint
