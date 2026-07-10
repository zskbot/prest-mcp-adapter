# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - Unreleased

### Added

- Initial `prest-mcp` stdio adapter that forwards MCP JSON-RPC to pREST’s HTTP `/_mcp` endpoint
- Config via `PREST_MCP_URL`, optional `PREST_MCP_TOKEN`, optional `PREST_MCP_TIMEOUT_MS`
- Dockerfile with MCP Registry ownership label
- `server.json` for MCP Registry metadata
- GitHub Actions test and release workflows
