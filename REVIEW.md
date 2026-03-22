# Code Review Guidelines

## General
- Follow Go conventions: effective Go, standard naming, error handling patterns
- Keep functions focused and small; avoid unnecessary abstractions
- No dead code, unused imports, or commented-out code
- All exported types and functions should have clear intent from naming

## Error Handling
- Always handle errors explicitly; never discard with `_`
- Use sentinel errors defined in the domain layer (`skill/` package)
- Wrap errors with context when propagating across layer boundaries

## Architecture
- Respect the layered architecture: domain → service → endpoint → transport
- Domain types (`skill/`, `vector/`) must not depend on infrastructure packages
- Persistence implementations must satisfy domain interfaces
- Do not leak transport-level concerns (HTTP status, JSON tags for request/response) into the service layer

## Security
- Validate all external input at the transport layer before passing to endpoints
- Do not expose internal error details to API consumers
- No hardcoded secrets or credentials
