# gt-telegram Release Steps

Step-by-step checklist for releasing a new version. Follow in order.

> **Key Rules**
>
> - NEVER push to main without `make ci` passing locally first
> - NEVER run `goreleaser release` locally — GitHub Actions handles it
> - ALWAYS do version bumps before tagging
> - Tag push triggers the release — only push the tag when ready

---

## Pre-Release

### 1. Version Bumps

Bump version in these files:

- [ ] `Makefile` line 7: `VERSION := X.Y.Z`
- [ ] `CHANGELOG.md`: add `## [X.Y.Z] - YYYY-MM-DD` section + link at bottom

**Note:** `main.go` uses `Version = "dev"` — set by ldflags at build time.
Do NOT hardcode.

### 2. CHANGELOG

- [ ] Add `## [X.Y.Z] - YYYY-MM-DD` section to `CHANGELOG.md`
- [ ] Include Added, Changed, Fixed sections as needed
- [ ] Add link at bottom: `[X.Y.Z]: https://github.com/leonletto/gt-telegram/releases/tag/vX.Y.Z`

### 3. Documentation Audit

Before committing, verify docs match code changes since the last release:

```bash
git log --oneline v<PREV>..HEAD | head -20
```

Check:
- [ ] `README.md` reflects any new commands, flags, or install methods
- [ ] `docs/setup.md` walkthrough matches current CLI behavior
- [ ] `docs/architecture.md` covers any new components or security changes
- [ ] `docs/troubleshooting.md` updated if new failure modes exist

### 4. Commit

- [ ] Commit all version bump and doc changes
- [ ] Do NOT push yet — quality gates first

---

## Quality Gates

**ALL must pass before pushing.** Fix failures, re-run, commit.

### 5. Full Local CI

```bash
make ci
```

Runs: `fmt` → `vet` → `test-race` → `build`

### 6. Manual Smoke Test (optional for patches, required for minor+)

```bash
GT_TOWN=/tmp/gt-test gt-telegram configure --token <test-token> --skip-pair
GT_TOWN=/tmp/gt-test gt-telegram status
GT_TOWN=/tmp/gt-test gt-telegram version
```

### 7. Fix Any Failures

If anything fails: fix, re-run the failing gate, commit.

---

## Release Sequence

### 8. Push Main

```bash
git push origin main
```

Wait for GitHub Actions CI to pass:

```bash
gh run list -R leonletto/gt-telegram --limit 1
```

**If CI fails: DO NOT tag.** Fix, push again.

### 9. Tag

```bash
git tag vX.Y.Z
```

Only tag on main, only after CI passes.

### 10. Push Tag

```bash
git push origin vX.Y.Z
```

This triggers `.github/workflows/release.yml`:

- GoReleaser builds linux + darwin, amd64 + arm64
- macOS binaries codesigned with Developer ID
- macOS binaries notarized with Apple
- Homebrew cask updated in `leonletto/homebrew-tap`
- Checksums generated, GitHub Release page created

### 11. Monitor Release

```bash
gh run list -R leonletto/gt-telegram --limit 1
```

Verify:

- [ ] GoReleaser builds all 4 platforms
- [ ] macOS binaries signed and notarized
- [ ] Checksums generated
- [ ] GitHub Release page shows correct version and artifacts

### 12. Update Release Notes

GoReleaser auto-generates notes from commits. Replace with cohesive notes:

```bash
gh release edit vX.Y.Z -R leonletto/gt-telegram --notes "$(cat <<'NOTES'
## [Summary — one line]

[1-3 sentence description of what this release adds/changes.]

### Added

- **Feature name** — description

### Changed

- **Area** — what changed

### Fixed

- **Bug name** — what was wrong and how it's fixed

**Full Changelog**: https://github.com/leonletto/gt-telegram/compare/vPREV...vX.Y.Z
NOTES
)"
```

---

## Post-Release

### 13. Verify Install Methods

- [ ] `curl -fsSL https://raw.githubusercontent.com/leonletto/gt-telegram/main/scripts/install.sh | sh`
- [ ] `gt-telegram version` shows new version
- [ ] `brew upgrade leonletto/tap/gt-telegram` or `brew install leonletto/tap/gt-telegram`
- [ ] `codesign -dv $(which gt-telegram)` shows `TeamIdentifier=XW56K59R9K`

### 14. Update Continuation Prompt

- [ ] Update `dev-docs/Continuation_Prompt.md` with new version, git state, and any changes

---

## Troubleshooting

### `goreleaser release` fails locally

Don't run it locally. The release is handled by GitHub Actions when the tag is
pushed. `GITHUB_TOKEN` and `HOMEBREW_TAP_GITHUB_TOKEN` are GitHub Actions
secrets.

### CI fails on main after push

Fix the issue, commit, push again. Do NOT tag until CI is green. Common causes:

- `gofmt` formatting: run `make fmt`
- Missing test updates for new code

### GitHub Actions secrets

These must be set on `leonletto/gt-telegram` repo (same values as thrum):

| Secret | Purpose |
|--------|---------|
| `APPLE_CERTIFICATE` | Base64-encoded .p12 Developer ID cert |
| `APPLE_CERTIFICATE_PASSWORD` | Password for the .p12 |
| `APPLE_ID` | Apple ID email |
| `APPLE_TEAM_ID` | `XW56K59R9K` |
| `APPLE_APP_SPECIFIC_PASSWORD` | App-specific password from appleid.apple.com |
| `HOMEBREW_TAP_GITHUB_TOKEN` | PAT with write access to `leonletto/homebrew-tap` |
