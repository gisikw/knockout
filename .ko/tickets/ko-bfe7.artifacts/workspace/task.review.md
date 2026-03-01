The implementation is complete and correct. All 6 planned tasks were executed with no deviations:

1. **ParseRegistry** — extended to handle both old and new format via `currentProject` tracking
2. **FormatRegistry** — emits nested format, no `prefixes:` section, no top-level `default:`
3. **isOldFormat + auto-migration** — silently rewrites old files on first read
4. **registry_test.go** — inputs updated, `TestParseRegistryNewFormat` and `TestLoadRegistryAutoMigrates` added
5. **cmd_list_test.go** — all 8 inline YAML strings updated
6. **txtar** — stdout assertions updated to match new format

Backward compatibility is preserved. Tests pass. No security issues. Scope stayed tightly within the registry serialization layer — the `Registry` struct and all its callers are untouched.

```json
{"disposition": "continue"}
```
