# SDK Publishing Guide

This document describes how to set up and use the SDK publishing workflow.

## GitHub Actions Workflow

The `publish-sdks.yml` workflow automates publishing all SDKs to their respective package registries.

### Triggers

1. **On Release**: Automatically triggered when a new GitHub release is published
2. **Manual Dispatch**: Can be triggered manually with custom options

### Manual Trigger Options

- **version**: Version to publish (e.g., `1.0.0`)
- **sdks**: Comma-separated list of SDKs to publish (e.g., `js,python`) or `all`
- **dry_run**: If true, validates packages without publishing

## Required Secrets

Configure these secrets in your repository settings (`Settings → Secrets and variables → Actions` on `PipeOpsHQ/rexec`):

### npm (JavaScript/TypeScript)

| Secret | Description |
|--------|-------------|
| `NPM_TOKEN` | npm automation token with publish access |

**How to get**:
1. Create/claim org scope at https://www.npmjs.com/org/create → `pipeopshq` (or use existing)
2. Go to https://www.npmjs.com/settings/tokens
3. Create new **Automation** token
4. Ensure the token can publish `@pipeopshq/rexec`

### PyPI (Python)

| Secret | Description |
|--------|-------------|
| `PYPI_TOKEN` | PyPI API token (`pypi-...`) |

**How to get**:
1. Create account at https://pypi.org/account/register/
2. Enable 2FA
3. Create API token at https://pypi.org/manage/account/token/ (scope: entire account, or project `rexec` after first upload)

### crates.io (Rust)

| Secret | Description |
|--------|-------------|
| `CRATES_IO_TOKEN` | crates.io API token |

**How to get**:
1. Log in with GitHub at https://crates.io/
2. Create token at https://crates.io/settings/tokens
3. Package name: **`pipeops-rexec`** (plain `rexec` is taken by another crate)

### RubyGems (Ruby)

| Secret | Description |
|--------|-------------|
| `RUBYGEMS_API_KEY` | RubyGems API key |

**How to get**:
1. Create account at https://rubygems.org/sign_up
2. Create API key at https://rubygems.org/profile/api_keys with **Push rubygem**
3. Package name: **`pipeops-rexec`** (plain `rexec` is taken)

### Maven Central (Java)

| Secret | Description |
|--------|-------------|
| `OSSRH_USERNAME` | Sonatype Central username |
| `OSSRH_TOKEN` | Sonatype Central token |
| `GPG_PRIVATE_KEY` | GPG private key for signing |
| `GPG_PASSPHRASE` | GPG key passphrase |

**How to get**:
1. Register at https://central.sonatype.com/
2. Claim namespace `io.pipeops` (verify domain or GitHub org)
3. Generate GPG key: `gpg --full-generate-key`
4. Export private key: `gpg --export-secret-keys --armor YOUR_KEY_ID`
5. Publish public key: `gpg --keyserver keyserver.ubuntu.com --send-keys YOUR_KEY_ID`

### NuGet (.NET)

| Secret | Description |
|--------|-------------|
| `NUGET_API_KEY` | NuGet.org API key |

**How to get**:
1. Create account at https://www.nuget.org/
2. Create API key at https://www.nuget.org/account/apikeys with push for **`PipeOps.Rexec`**

### Packagist (PHP)

PHP packages are updated via Packagist + GitHub. No Actions secrets required after submit.

**Setup**:
1. Create account at https://packagist.org/
2. Submit package pointing at the monorepo path or a dedicated repo
3. Composer package name: **`pipeopshq/rexec`**

### Go SDK sync (optional)

| Secret | Description |
|--------|-------------|
| `GO_SDK_PUSH_TOKEN` | GitHub PAT with `repo` scope that can push to `PipeOpsHQ/rexec-go` |

Used by `.github/workflows/sync-go-sdk.yml` so monorepo changes under `sdk/go/` are mirrored to the standalone Go module repo.

## SDK Package Registry URLs

| SDK | Registry | Package Name |
|-----|----------|--------------|
| JavaScript | [npm](https://www.npmjs.com/package/@pipeopshq/rexec) | `@pipeopshq/rexec` |
| Python | [PyPI](https://pypi.org/project/pipeops-rexec/) | `pipeops-rexec` |
| Rust | [crates.io](https://crates.io/crates/pipeops-rexec) | `pipeops-rexec` |
| Ruby | [RubyGems](https://rubygems.org/gems/pipeops-rexec) | `pipeops-rexec` |
| Java | [Maven Central](https://central.sonatype.com/artifact/io.pipeops/rexec) | `io.pipeops:rexec` |
| .NET | [NuGet](https://www.nuget.org/packages/PipeOps.Rexec) | `PipeOps.Rexec` |
| PHP | [Packagist](https://packagist.org/packages/pipeopshq/rexec) | `pipeopshq/rexec` |
| Go | GitHub | `github.com/PipeOpsHQ/rexec-go` |

## Version Management

The workflow automatically:
1. Extracts version from release tag (removes `v` prefix)
2. Updates version in each SDK's package manifest
3. Builds and publishes the package

### Manual Version Bump

To manually update versions across all SDKs:

```bash
# JavaScript
cd sdk/js && npm version 1.2.3 --no-git-tag-version

# Python
sed -i 's/version = ".*"/version = "1.2.3"/' sdk/python/pyproject.toml

# Rust
sed -i 's/^version = ".*"/version = "1.2.3"/' sdk/rust/Cargo.toml

# Ruby
sed -i 's/spec.version = ".*"/spec.version = "1.2.3"/' sdk/ruby/rexec.gemspec

# Java
cd sdk/java && mvn versions:set -DnewVersion=1.2.3

# .NET
sed -i 's/<Version>.*<\/Version>/<Version>1.2.3<\/Version>/' sdk/dotnet/Rexec.csproj

# PHP (version in composer.json is optional for libraries)
```

## Dry Run

To test publishing without actually releasing:

1. Go to Actions → Publish SDKs → Run workflow
2. Enter version (e.g., `1.0.0`)
3. Select SDKs to test
4. Check "Dry run" checkbox
5. Run workflow

The workflow will validate packages and show what would be published.

## Troubleshooting

### npm: 403 Forbidden
- Ensure `NPM_TOKEN` has publish access
- Check if package name is available or you own the scope

### PyPI: 400 Bad Request
- Version may already exist on PyPI
- Check package name conflicts

### crates.io: Unauthorized
- Regenerate token at crates.io
- Ensure crate name is available

### Maven Central: Signature Failed
- Verify GPG key is correctly exported
- Check passphrase is correct
- Ensure public key is published to keyserver

### NuGet: API Key Invalid
- Regenerate API key
- Ensure key has push permissions for package

## Go SDK Note

The Go SDK doesn't require registry publishing. Go modules are fetched directly from the Git repository. When a new version is tagged:

1. Users can reference it: `go get github.com/PipeOpsHQ/rexec-go@v1.2.3`
2. The Go module proxy caches it automatically
3. No additional publishing steps needed

To verify Go module availability:
```bash
go list -m github.com/PipeOpsHQ/rexec-go@v1.2.3
```
