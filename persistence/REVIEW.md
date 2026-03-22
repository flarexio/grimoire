# Persistence Layer Review Guidelines

- Implementations must satisfy domain interfaces (`skill.Repository`, `vector.VectorDB`, `vector.Collection`)
- Handle file system errors gracefully (missing files, permission denied, malformed YAML)
- Do not let persistence-specific types leak into the domain layer
- Vector DB operations should handle edge cases (empty collections, duplicate documents)
