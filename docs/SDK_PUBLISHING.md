# SDK Publishing Guide

All official SDK publishing is managed **only** through GitHub Actions. Do not publish packages from a laptop except for emergencies.

## GitHub Actions Workflow

Workflow file: [`.github/workflows/publish-sdks.yml`](../.github/workflows/publish-sdks.yml)

| SDK | Registry | Secret | Package |
|-----|----------|--------|---------|
| JavaScript | npm | `NPM_TOKEN` | `pipeops-rexec` |
| Python | PyPI | `PYPI_TOKEN` | `pipeops-rexec` |
| Rust | crates.io | `CRATES_IO_TOKEN` | `pipeops-rexec` |
| Ruby | RubyGems | `RUBYGEMS_API_KEY` | `pipeops-rexec` |
| .NET | NuGet | `NUGET_API_KEY` | `PipeOps.Rexec` |
| Java | Maven Central | `CENTRAL_*` + GPG (or legacy `OSSRH_*`) | `io.pipeops:rexec` |
| PHP | Packagist | `SDK_PUSH_TOKEN` (sync) + Packagist webhook | `pipeopshq/rexec` |
| Go | GitHub module | `SDK_PUSH_TOKEN` (sync) | `github.com/PipeOpsHQ/rexec-go` |

### Triggers

1. **On Release** — when a GitHub Release is published (`release.yml` or manual release UI)
2. **Manual dispatch** — Actions → **Publish SDKs** → Run workflow
3. **workflow_call** — reusable from other workflows

### Manual trigger options

- **version**: Version to publish (e.g., `1.0.0`)
- **sdks**: Comma-separated list (e.g., `js,python,rust`) or `all`
- **dry_run**: Validate only, no upload

Re-runs are **idempotent**: if a version already exists on the registry, the job skips or treats that as success.

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
4. Ensure the token can publish `pipeops-rexec`

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

### Maven Central (Java) — Sonatype Central Publisher Portal

Publishing uses the **Central Publisher Portal** Maven plugin
(`central-publishing-maven-plugin` in `sdk/java/pom.xml` profile `release`).

| Secret | Description |
|--------|-------------|
| `CENTRAL_USERNAME` | Portal **User Token** username (preferred) |
| `CENTRAL_TOKEN` | Portal **User Token** password/token (preferred) |
| `OSSRH_USERNAME` | Legacy alias — still accepted if `CENTRAL_USERNAME` unset |
| `OSSRH_TOKEN` | Legacy alias — still accepted if `CENTRAL_TOKEN` unset |
| `GPG_PRIVATE_KEY` | ASCII-armored private key used to sign artifacts |
| `GPG_PASSPHRASE` | Passphrase for that GPG key |

#### One-time setup (you do this in the browser)

1. **Create account** at https://central.sonatype.com/ and sign in (GitHub SSO is fine).
2. **Generate a User Token**  
   Account → **View Account** / **User Tokens** → **Generate User Token**.  
   Save the username + password pair — these become `CENTRAL_USERNAME` / `CENTRAL_TOKEN`.
3. **Register namespace `io.pipeops`**  
   - Namespaces → **Add Namespace** → `io.pipeops`  
   - Verify ownership (DNS TXT on `pipeops.io`, or GitHub org verification when offered).  
   - Wait until the namespace shows as **verified**. You cannot publish until this is done.
4. **Create a GPG signing key** (on your laptop):

   ```bash
   gpg --full-generate-key
   # RSA 4096, no expiry (or long expiry), real name/email matching publisher identity
   gpg --list-secret-keys --keyid-format LONG
   # Note the KEY_ID (e.g. ABCD1234EFGH5678)

   gpg --export-secret-keys --armor YOUR_KEY_ID > private-key.asc
   gpg --export --armor YOUR_KEY_ID > public-key.asc
   gpg --keyserver keys.openpgp.org --send-keys YOUR_KEY_ID
   # also try: keyserver.ubuntu.com
   ```

5. **Add GitHub Actions secrets** on `PipeOpsHQ/Rexec`  
   Settings → Secrets and variables → Actions:

   | Name | Value |
   |------|--------|
   | `CENTRAL_USERNAME` | User token username from step 2 |
   | `CENTRAL_TOKEN` | User token password from step 2 |
   | `GPG_PRIVATE_KEY` | Full contents of `private-key.asc` (including BEGIN/END lines) |
   | `GPG_PASSPHRASE` | Key passphrase |

6. **Publish**  
   Actions → **Publish SDKs** → Run workflow:

   - version: `1.0.1` (or next)
   - sdks: `java`
   - dry_run: `false`

   Or cut a GitHub Release (tags all SDKs). The Java job runs `mvn deploy -P release` and auto-publishes via the portal.

7. **Verify** after ~10–30 minutes:  
   https://central.sonatype.com/artifact/io.pipeops/rexec  
   or https://search.maven.org/artifact/io.pipeops/rexec

Local emergency publish (only if Actions is broken):

```bash
cd sdk/java
# ~/.m2/settings.xml server id "central" with your portal token
mvn clean deploy -P release
```

### NuGet (.NET)

| Secret | Description |
|--------|-------------|
| `NUGET_API_KEY` | NuGet.org API key |

**How to get**:
1. Create account at https://www.nuget.org/
2. Create API key at https://www.nuget.org/account/apikeys with push for **`PipeOps.Rexec`**

### Packagist (PHP)

Packagist **cannot** install from a monorepo subdirectory cleanly. We publish from a
dedicated repo that always has `composer.json` at the root:

| Piece | Value |
|-------|--------|
| Dedicated repo | https://github.com/PipeOpsHQ/rexec-php |
| Package name | `pipeopshq/rexec` |
| Sync workflow | [`.github/workflows/sync-php-sdk.yml`](../.github/workflows/sync-php-sdk.yml) |
| Install | `composer require pipeopshq/rexec` |

#### One-time setup (you do this once)

1. **Confirm the sync repo exists** — `PipeOpsHQ/rexec-php` (already created; re-run
   **Sync PHP SDK** if empty).
2. **Cross-repo PAT** (same secret as Go sync):
   - GitHub → Settings → Developer settings → Personal access tokens  
   - Classic PAT with `repo` scope (or fine-grained: write to `rexec-php` + `rexec-go`)  
   - Add as Actions secret **`SDK_PUSH_TOKEN`** on `PipeOpsHQ/Rexec`  
   - Legacy names still work: `PHP_SDK_PUSH_TOKEN`, `GO_SDK_PUSH_TOKEN`
3. **Packagist account** at https://packagist.org/ (GitHub login recommended).
4. **Submit the package**:
   - https://packagist.org/packages/submit  
   - Repository URL: **`https://github.com/PipeOpsHQ/rexec-php`**  
   - **Not** the monorepo URL
5. Enable **GitHub Service Hook / Auto-Update** on the Packagist package page  
   (so new tags on `rexec-php` appear on Packagist within minutes).
6. Optional: GitHub App install for Packagist if prompted.

#### Ongoing releases

| Trigger | What happens |
|---------|----------------|
| Push to `main` touching `sdk/php/**` | Syncs files → `rexec-php` main |
| **Publish SDKs** with `php` / `all` | Syncs + tags `vX.Y.Z` on `rexec-php` |
| Manual **Sync PHP SDK** + version input | Sync + optional tag |

After Packagist is linked, tags become installable as:

```bash
composer require pipeopshq/rexec:1.0.1
```

### Go / PHP SDK sync PAT

| Secret | Description |
|--------|-------------|
| `SDK_PUSH_TOKEN` | **Preferred** GitHub PAT with `repo` write to `rexec-go` and `rexec-php` |
| `GO_SDK_PUSH_TOKEN` | Legacy alias for Go only |
| `PHP_SDK_PUSH_TOKEN` | Legacy alias for PHP only |

Used by:

- [`.github/workflows/sync-go-sdk.yml`](../.github/workflows/sync-go-sdk.yml) → `PipeOpsHQ/rexec-go`
- [`.github/workflows/sync-php-sdk.yml`](../.github/workflows/sync-php-sdk.yml) → `PipeOpsHQ/rexec-php`
- Publish SDKs → PHP job (tag release on `rexec-php`)

## SDK Package Registry URLs

| SDK | Registry | Package Name |
|-----|----------|--------------|
| JavaScript | [npm](https://www.npmjs.com/package/pipeops-rexec) | `pipeops-rexec` |
| Python | [PyPI](https://pypi.org/project/pipeops-rexec/) | `pipeops-rexec` |
| Rust | [crates.io](https://crates.io/crates/pipeops-rexec) | `pipeops-rexec` |
| Ruby | [RubyGems](https://rubygems.org/gems/pipeops-rexec) | `pipeops-rexec` |
| Java | [Maven Central](https://central.sonatype.com/artifact/io.pipeops/rexec) | `io.pipeops:rexec` |
| .NET | [NuGet](https://www.nuget.org/packages/PipeOps.Rexec) | `PipeOps.Rexec` |
| PHP | [Packagist](https://packagist.org/packages/pipeopshq/rexec) · [source](https://github.com/PipeOpsHQ/rexec-php) | `pipeopshq/rexec` |
| Go | [GitHub](https://github.com/PipeOpsHQ/rexec-go) | `github.com/PipeOpsHQ/rexec-go` |

## Checklist: publish PHP + Java for the first time

### PHP

- [x] Dedicated repo `PipeOpsHQ/rexec-php` exists and has tagged `v1.0.1`
- [ ] Secret `SDK_PUSH_TOKEN` set on monorepo (PAT can push to `rexec-php`)
- [ ] Packagist package submitted with URL `https://github.com/PipeOpsHQ/rexec-php`
- [ ] Packagist auto-update / GitHub webhook enabled
- [ ] `composer require pipeopshq/rexec` works

### Java

- [ ] Sonatype Central account + namespace `io.pipeops` **verified**
- [ ] Secrets: `CENTRAL_USERNAME`, `CENTRAL_TOKEN`, `GPG_PRIVATE_KEY`, `GPG_PASSPHRASE`
- [ ] Actions → Publish SDKs → version `1.0.1`, sdks `java`
- [ ] Artifact visible on Maven Central search

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
