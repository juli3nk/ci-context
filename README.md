# ci-context

Detects the git context in CI environments and exposes normalized refs.

## Installation

### From source

```bash
go install github.com/juli3nk/ci-context/cmd/ci-context@latest
```

### From releases

Download a prebuilt binary from the [releases page](https://github.com/juli3nk/ci-context/releases).

## Usage

```bash
# Default JSON output
ci-context

# GitHub Actions compatible output
ci-context --github-output

# Force the base ref
ci-context --base-ref=origin/main

# Set a minimum reference commit
ci-context --since=v1.0.0

# Inspect the detected CI context
ci-context --debug
```

## Flags

| Flag | Description |
|------|-------------|
| `--base-ref` | Force the base reference. |
| `--since` | Minimum reference commit. The detected base will never be older than this commit. |
| `--github-output` | Output in `GITHUB_OUTPUT` format. |
| `--debug` | Print the detected CI context and exit. |

## Outputs

### JSON (default)

```json
{
  "base_ref": "origin/main",
  "head_ref": "HEAD",
  "commit_count": 42
}
```

### GitHub Actions format

```text
base_ref=origin/main
head_ref=HEAD
commit_count=42
```

## Supported CI providers

- GitHub Actions
- GitLab CI
- Local / fallback

## License

See [LICENSE](./LICENSE).
