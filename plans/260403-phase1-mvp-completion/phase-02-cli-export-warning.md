# Phase 2: CLI Export — Unencrypted Export Warning + `--force` Flag

## Context

- **Priority:** HIGH | **Effort:** 30m | **Risk:** Low
- PRD §8.7: "unencrypted CSV (with warning)"
- CLI `export --format=csv|json` exports without any plaintext warning
- macOS app has confirmation ("type EXPORT"), CLI does not
- `promptYesNo()` helper already exists in `packages/cli/cmd/helpers.go`

## Related Files

| Action | File |
|--------|------|
| MODIFY | `packages/cli/cmd/export_cmd.go` |

## Implementation Steps

### 1. Add `--force` flag

```go
var (
    exportFormat     string
    exportOutputFile string
    exportForce      bool
)

func init() {
    exportCmd.Flags().StringVar(&exportFormat, "format", "json", "export format: json|csv|encrypted")
    exportCmd.Flags().StringVar(&exportOutputFile, "file", "", "output file path (default: stdout)")
    exportCmd.Flags().BoolVar(&exportForce, "force", false, "skip plaintext export warning")
    rootCmd.AddCommand(exportCmd)
}
```

### 2. Add warning before plaintext export

Insert before `switch exportFormat` block:

```go
if exportFormat != "encrypted" && !exportForce {
    fmt.Fprintln(cmd.ErrOrStderr(), "⚠ WARNING: This exports your vault in PLAINTEXT.")
    fmt.Fprintln(cmd.ErrOrStderr(), " Anyone with this file can read ALL your secrets.")
    if !promptYesNo("Continue?") {
        return fmt.Errorf("export cancelled")
    }
}
```

## Todo List

- [x] Add `exportForce` bool variable
- [x] Add `--force` flag registration
- [x] Add plaintext warning block before switch
- [x] Run `go test ./packages/cli/cmd/ -v`

## Success Criteria

- `zeropass export --format=csv` (no `--force`) → warning prompt appears
- `zeropass export --format=csv --force` → exports without prompt
- `zeropass export --format=json` (no `--force`) → warning prompt appears
- `zeropass export --format=encrypted` → no warning (already encrypted)
- Warning outputs to stderr (not stdout) so piping still works
