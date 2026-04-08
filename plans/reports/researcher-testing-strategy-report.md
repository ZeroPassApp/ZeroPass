# Testing Strategy: Tauri v2 + React 19 + Rust

## Summary

Achieve >90% coverage with **Vitest v8** (React), **cargo test** (Rust), **Playwright** (E2E). Vitest v8: 50% faster than Istanbul, AST-aware accuracy, low memory. Playwright only maintained E2E solution (WebDriverIO deprecated for Tauri).

## React Frontend (Vitest + React Testing Library)

**Install:** `npm install -D vitest @vitest/coverage-v8 @testing-library/react happy-dom`

**vitest.config.ts:** Use `happy-dom` environment, v8 coverage provider, set `coverage.lines = 90`.

**Mock Tauri API (tests/setup.ts):**
```ts
vi.mock('@tauri-apps/api/core', () => ({
  invoke: vi.fn(), listen: vi.fn(), emit: vi.fn(),
}))
```
Mock pattern: `vi.mocked(invoke).mockResolvedValueOnce({...})`

**Store testing:** Direct state via `useVaultStore().getState()`. No special harness needed.

## Rust Backend (cargo test)

**Commands:**
```bash
cargo test --lib                 # unit tests
cargo test --test integration    # integration tests
cargo test --release -- --test-threads=1  # if FFI state conflicts
```

**Pattern (in #[cfg(test)]):**
```rust
#[test]
fn test_unlock_wrong_password() {
    assert!(unlock_vault("wrong".to_string()).is_err());
}
```

**FFI bridge:** Real C lib → linker resolves naturally in `cargo test`. For isolation: wrap FFI calls in `#[cfg(test)]` stubs. Document if Go state persists between tests.

## E2E Testing (Playwright)

**Why Playwright?** WebDriverIO Tauri support incomplete/deprecated. Playwright: native desktop CDP, headless automation.

**Setup:** `npm install -D @playwright/test`

**playwright.config.ts:** Single worker in CI (`workers: process.env.CI ? 1 : undefined`). WebServer: `cargo tauri dev` on port 1430.

**Test pattern:**
```ts
test('vault workflow', async ({ page }) => {
  await page.goto('http://localhost:1430')
  await page.fill('[data-testid=password]', 'pw123')
  await page.click('text=Create Vault')
  // Assert flow states...
})
```

## Coverage Reporting

**Vitest:** `npm run coverage` → v8 provider outputs HTML. Config: `coverage.include: ['src/**/*.ts']` to capture uncovered files.

**Rust:** `cargo-tarpaulin` or `cargo-llvm-cov` → separate Lcov report.

**CI/CD pipe:**
```yaml
- run: npm run coverage
- run: cargo tarpaulin --out Lcov
- uses: codecov/codecov-action@v4
  with: { files: './coverage/lcov.info,./cobertura.xml' }
```

## CI/CD Considerations

- **Vitest:** Uses happy-dom (no browser binary needed).
- **Playwright:** Headless by default; runs Chrome/WebKit.
- **Rust:** Ensure C library built before `cargo test`.
- **Database:** Use SQLite in-memory or fixture setup/teardown.
- **Parallelization:** Vitest parallel default; limit E2E/Rust to single worker if FFI state shared.

## Recommended Stack

| Layer | Tool | Rationale |
|-------|------|-----------|
| React Components | Vitest + RTL | Fast, Jest-compat, v8 coverage |
| Stores/Hooks | Vitest | Direct state access |
| Tauri Commands | cargo test | Isolated Rust units |
| FFI Bridge | cargo test (#[cfg(test)]) | Mock C layer in tests |
| E2E | Playwright | Only maintained solution |
| Coverage | Vitest v8 + tarpaulin | Both recommended |

## Implementation Steps

1. Install test deps (Vitest, RTL, coverage, Playwright, tarpaulin).
2. Create vitest.config + tests/setup.ts with Tauri API mocks.
3. Write 3-5 component tests + 1 store test.
4. Add `#[cfg(test)]` blocks to each Tauri command.
5. Bootstrap Playwright E2E suite (3-5 smoke tests).
6. Add npm scripts: `test`, `coverage`, `test:e2e`.
7. Configure CI: run Vitest, cargo test, Playwright in parallel.
8. Target: 90% frontend, 85% backend coverage, E2E happy path.

## Unresolved

- Can Vitest browser mode + Playwright coexist? (Yes, but separate them for clarity.)
- Does `cargo test` isolate Go/C library state per test? (Verify + document.)
