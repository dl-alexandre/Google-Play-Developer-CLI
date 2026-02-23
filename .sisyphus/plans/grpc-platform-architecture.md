# Work Plan: gRPC Platform Architecture for Google Play Developer CLI

## TL;DR

> **Complete architectural transformation** from stateless Kong CLI to gRPC-first platform with local daemon, MCP/AI integration, and polyglot plugin system.
>
> **Scope**: 4 phases over 20-30 weeks (5-7.5 months)
> - Phase 0: Foundation & RFCs (2-4 weeks)
> - Phase 1: gRPC Daemon Core (6-8 weeks)
> - Phase 2: MCP/AI Integration (4-6 weeks)
> - Phase 3: Plugin System (6-8 weeks)
> - Phase 4: Final Migration (2-4 weeks)
>
> **Deliverables**:
> - gRPC daemon with connection pooling + credential caching
> - Protobuf definitions for all 70+ operations (14 service domains)
> - MCP server for AI agent integration
> - Plugin architecture via HashiCorp go-plugin
> - Dual-path execution (old CLI + new gRPC) with gradual migration
>
> **Estimated Effort**: Large
> **Parallel Execution**: YES - 4 concurrent workstreams in Phase 1-3
> **Critical Path**: Proto definitions → Core services → Service implementations → Integration tests → Final migration

---

## Context

### Original Request
Transform the Google Play Developer CLI (gpd) from a stateless Kong-based REST CLI into a gRPC-first platform architecture. This includes:
- gRPC local daemon for connection pooling and credential caching
- Model Context Protocol (MCP) server for AI agent integration
- Polyglot plugin system using HashiCorp go-plugin
- Breaking change from current architecture to new gRPC-based system

### Metis Review Findings
**Identified Gaps** (addressed in this plan):
- **Phase 0 added**: RFCs and prototypes before implementation
- **Dual-path execution**: Old Kong CLI preserved during transition
- **Compatibility Charter**: Defined with 4-release window
- **Parity scorecards**: By namespace (14 scorecards)
- **Go/No-Go gates**: At phase boundaries
- **Technical decisions resolved**:
  - Transport: Unix Domain Socket (UDS) default, named pipes Windows
  - Process model: Per-user singleton, autostart on first CLI call
  - API shape: Bounded service domains (14 proto packages)
  - Migration: Dual-path with feature flags, gradual default switch

**Unmade Decisions** (awaiting user input):
1. Compatibility window: 2/4/6 releases?
2. Default switch criteria: 100%/95%+escape/never?
3. Plugin distribution: GitHub/built-in registry/manual?
4. MCP priority: Phase 2 critical/Phase 3/future only?

### Research Findings
- Google Play APIs are REST-only - gRPC is for local architecture only
- Current cold start <200ms, stateless design
- File-based edit locking works but has limitations at scale
- Kong CLI with 70+ commands across 14 namespaces
- JSON-first output with field projection already optimized

---

## Work Objectives

### Core Objective
Transform gpd from a stateless CLI into a gRPC-first platform with local daemon, AI/MCP integration, and plugin extensibility, while maintaining backward compatibility during transition.

### Concrete Deliverables
1. **Phase 0**: RFCs, prototypes, Compatibility Charter, parity scorecards
2. **Phase 1**: gRPC daemon, core services (auth, edits, publish), connection pooling, credential caching, file locking replacement
3. **Phase 2**: MCP server, bidirectional streaming, resource watchers, LLM documentation export
4. **Phase 3**: Plugin architecture, Go SDK, Python SDK, sample plugins
5. **Phase 4**: Default path switch, old path deprecation, final migration

### Definition of Done
- [ ] All 70+ CLI commands have gRPC equivalents with parity tests passing
- [ ] MCP server passes Model Context Protocol validation
- [ ] Plugin system supports Go and Python plugins with isolation
- [ ] Dual-path execution works with feature flags
- [ ] Performance meets SLOs (p95 < 100ms local, daemon RSS < 50MB idle)
- [ ] Security review passes (credential isolation, no unscoped plugin access)

### Must Have
- Non-regression: Preserve JSON envelope + exit codes for 4 releases
- Performance: Daemon cold start < 500ms, RPC latency < 10ms local
- Security: Credentials never in plugin memory without scoping
- Operational: Daemon crash must not block CLI (fallback mode)
- Testing: Contract tests for all 70+ operations before switch

### Must NOT Have (Guardrails)
- **No breaking changes to JSON schema during transition**
- **No removal of old Kong CLI path until Phase 4**
- **No net-new end-user features in Phase 1** (parity only)
- **No polyglot plugins beyond Go/Python in Phase 3**
- **No default switch without 95%+ parity + escape hatch**
- **No AI intelligence features** (transport/contract only in Phase 2)

---

## Verification Strategy

### Test Decision
- **Infrastructure exists**: YES (existing Go test setup)
- **Automated tests**: YES (TDD for new code, parity tests for migration)
- **Framework**: Go testing + gRPC reflection testing + contract tests
- **Test strategy**: 
  - Unit tests for individual services
  - Integration tests for daemon + CLI
  - Contract tests comparing old vs new output
  - Parity scorecards by namespace

### QA Policy
Every task MUST include agent-executed QA scenarios. Evidence saved to `.sisyphus/evidence/`.

- **gRPC services**: Use Go test client, verify proto contracts
- **Daemon lifecycle**: Use Bash scripts, verify start/stop/restart
- **CLI integration**: Use Bash, verify dual-path execution
- **MCP server**: Use Go MCP client, verify tool definitions
- **Plugins**: Use sample plugins, verify isolation + communication

---

## Execution Strategy

### Phase Overview

```
Phase 0 (Weeks 1-4): Foundation
├── RFCs for daemon, auth, state, migration
├── Prototypes for transport, process model
├── Compatibility Charter + parity scorecards
└── Go/No-Go gate 0

Phase 1 (Weeks 5-12): Daemon Core (MAX PARALLEL)
├── Proto definitions (14 service domains)
├── Core services implementation
├── Connection pooling + credential caching
├── Edit coordination (replacing file locks)
├── CLI → gRPC client thin wrapper
└── Go/No-Go gate 1

Phase 2 (Weeks 10-16): MCP/AI (parallel with Phase 1 late)
├── MCP server implementation
├── Bidirectional streaming for watchers
├── Resource watcher service
├── LLM documentation export
└── Go/No-Go gate 2

Phase 3 (Weeks 14-22): Plugin System (parallel with Phase 2)
├── go-plugin integration
├── Plugin SDK (Go)
├── Python plugin support
├── Sample plugins
└── Go/No-Go gate 3

Phase 4 (Weeks 20-24): Final Migration
├── Dual-path default switch
├── Old path deprecation warnings
├── Documentation + migration guide
└── Final Go/No-Go gate 4
```

### Parallel Execution Waves

**Wave 1 (Phase 0 - Foundation):**
- Tasks 1-3: RFCs and Charter
- Tasks 4-6: Prototypes
- Tasks 7-8: Scorecards + gate

**Wave 2 (Phase 1 - Proto & Core - MAX PARALLEL):**
- Tasks 9-12: Proto definitions (14 packages)
- Tasks 13-18: Core services (auth, edits, publish, reviews, etc.)
- Tasks 19-22: Infrastructure (pooling, caching, locking)
- Tasks 23-26: CLI client wrapper
- Tasks 27-32: Service implementations by namespace
- Tasks 33-40: Parity tests per namespace

**Wave 3 (Phase 2 & 3 - MCP & Plugins - parallel):**
- Tasks 41-46: MCP server + streaming
- Tasks 47-52: Plugin architecture

**Wave 4 (Phase 4 - Migration):**
- Tasks 53-56: Default switch
- Tasks 57-60: Deprecation + docs

**Wave FINAL (After ALL tasks):**
- Tasks F1-F4: Compliance, quality, QA, scope check

---

## TODOs

