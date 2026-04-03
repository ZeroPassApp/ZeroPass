# Phase 1: CLI Import — Missing Safari, LastPass, KeePass, 1PUX Sources

## Context

- **Priority:** HIGH | **Effort:** 30m | **Risk:** Low
- Core library already implements `ImportSafari()`, `ImportLastPass()`, `ImportKeePass()`, `Import1PUX()` in `core/vault/importexport/import.go`
- CLI `import` command only exposes 5 sources: `chrome|firefox|1password|bitwarden|csv`
- Bridge layer already exports all 9 sources (`ZPImportSafari`, etc.)

## Related Files

| Action | File |
|--------|------|
| MODIFY | `packages/cli/cmd/import_cmd.go` |

## Implementation Steps

### 1. Add 4 cases to switch statement

In `import_cmd.go`, the `switch importFrom` block (around line 53) currently handles: `chrome`, `firefox`, `1password`, `bitwarden`, `csv`.

Add 4 new cases before the `csv` case:

```go
case "safari":
    imported, err := importexport.ImportSafari(f)
    if err != nil {
        return fmt.Errorf("import Safari: %w", err)
    }
    return addImportedItems(cmd, mgr, imported)
case "lastpass":
    imported, err := importexport.ImportLastPass(f)
    if err != nil {
        return fmt.Errorf("import LastPass: %w", err)
    }
    return addImportedItems(cmd, mgr, imported)
case "keepass":
    imported, err := importexport.ImportKeePass(f)
    if err != nil {
        return fmt.Errorf("import KeePass: %w", err)
    }
    return addImportedItems(cmd, mgr, imported)
case "1pux":
    imported, err := importexport.Import1PUX(f)
    if err != nil {
        return fmt.Errorf("import 1PUX: %w", err)
    }
    return addImportedItems(cmd, mgr, imported)
```

### 2. Update flag description and error message

- Update `--from` flag description to include all sources
- Update `default:` error message to:
  `"unknown import source: %s. Use chrome, firefox, safari, 1password, 1pux, bitwarden, lastpass, keepass, or csv"`

## Todo List

- [x] Add safari case to switch
- [x] Add lastpass case to switch
- [x] Add keepass case to switch
- [x] Add 1pux case to switch
- [x] Update --from flag description
- [x] Update default error message
- [x] Run `go test ./packages/cli/cmd/ -v`

## Success Criteria

- `zeropass import --from=safari --file=test.csv` succeeds
- `zeropass import --from=lastpass --file=test.csv` succeeds
- `zeropass import --from=keepass --file=test.csv` succeeds
- `zeropass import --from=1pux --file=test.1pux` succeeds
- `--help` shows all 9 sources
- Existing imports unaffected
