# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.3] - 2026-07-09

### Changed

- MCP Registry / docs version bump to `0.1.3` for publish

## [0.1.2] - 2026-07-09

### Fixed

- errcheck findings on stderr writes and HTTP response body close

## [0.1.1] - 2026-07-09

### Fixed

- Release workflow only downloads `prest-mcp-*` artifacts (skips Buildx dockerbuild blobs)
- golangci-lint bumped to v2.12.2 for Go 1.26

## [0.1.0] - 2026-07-09

### Added

- Initial `prest-mcp` stdio adapter that forwards MCP JSON-RPC to pREST’s HTTP `/_mcp` endpoint
- Config via `PREST_MCP_URL`, optional `PREST_MCP_TOKEN`, optional `PREST_MCP_TIMEOUT_MS`
- Dockerfile with MCP Registry ownership label
- `server.json` for MCP Registry metadata
- GitHub Actions test and release workflows
