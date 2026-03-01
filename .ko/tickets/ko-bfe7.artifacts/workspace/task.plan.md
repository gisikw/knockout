The plan is ready. The open question about auto-migration has been answered (silent auto-migration confirmed), so all 6 tasks are now fully actionable:

1. **`registry.go:ParseRegistry`** — extend parser to handle new nested format while keeping old flat format readable
2. **`registry.go:FormatRegistry`** — emit new nested format (no separate `prefixes:` section, no top-level `default:`)
3. **`registry.go:isOldFormat + LoadRegistry`** — detect old format and silently rewrite on first read
4. **`registry_test.go`** — update inline YAML to new format, add `TestParseRegistryNewFormat`
5. **`cmd_list_test.go`** — update ~8 inline `regContent` YAML strings to new format
6. **Full test suite** — `go test ./...` green
