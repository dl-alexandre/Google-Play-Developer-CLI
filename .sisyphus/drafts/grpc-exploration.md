# Draft: gRPC Platform Architecture - Critical Decisions & Gaps

## Metis Gap Analysis Results (Completed)

### Critical Unanswered Questions
1. **Backward Compatibility Contract**: What must remain compatible and for how long?
   - CLI flags, JSON envelope, exit codes, stderr/stdout behavior
2. **Migration Mode**: Big bang vs dual-path with feature flags?
3. **SLOs**: Cold start, p95 latency, daemon memory, CPU idle targets
4. **Daemon Lifecycle**: Autostart, manual, per-user, per-project, system service?
5. **Auth Boundaries**: Credential ownership, process isolation, least privilege
6. **Offline/Degraded Behaviors**: When daemon dies, socket stale, plugin fails
7. **Protobuf Compatibility Policy**: Field evolution, versioning, deprecation
8. **Support Matrix**: macOS/Linux/Windows, UDS vs TCP security
9. **Streaming Needs**: Which ops truly need watchers vs unary RPC?
10. **Consumer Contracts**: MCP agents, CI runners, plugins - what do they need?

### Guardrails to Define (From Metis)
- Non-regression contract: preserve JSON/exit codes for N releases
- Performance: measurable targets (p95, RSS, reconnect)
- Security: no unscoped credentials in plugins
- Operational: daemon crash must not block; guaranteed fallback
- Versioning: semantic protobuf versioning + CI compatibility tests
- Scope: Phase 1 = parity only, no net-new features
- Rollout: canary, kill switch, rollback per phase
- Testing: contract tests for all 70+ ops before switch

### Scope Creep Risks
- Watchers becoming full event framework
- Over-designing plugin SDK before daemon stability
- "Cleaning up" UX/output while doing architecture
- Building MCP intelligence vs transport first
- Cross-platform service management abstractions
- Too many plugin languages early

### Assumptions Needing Validation
- Daemon improves latency (may regress one-shot commands)
- Connection pooling useful for all patterns
- Credential caching safe enough (threat model change)
- All 70+ ops map cleanly to RPC
- MCP needs bidirectional streaming immediately
- go-plugin manageable in user envs
- 3-6 months feasible (depends on testing burden)

### Unmade Technical Decisions
- Transport: UDS vs localhost TCP, security model
- Process model: singleton vs per-workspace, autostart
- API shape: monolith vs bounded domains, unary/streaming
- State model: cache/edit-lock location, survival/rebuild
- Failure model: retry at client/daemon/plugin boundaries
- Plugin boundary: capability model, sandboxing, stability
- Observability: tracing IDs, log redaction
- Migration: shadow mode, diffing outputs, switch criteria

### Risk Factors (Breaking Change)
- Silent contract drift breaks CI scripts
- Daemon introduces new failure classes (stale sockets, zombies)
- Security regression via long-lived creds
- Debuggability worsens (multi-process stack)
- Plugin ecosystem stalls due to unstable protocol
- Team throughput collapse
- "Almost done" trap: 90% parity hides long-tail effort

## Strategic Response: Phase 0 Required

Based on Metis analysis, this plan needs a **Phase 0** (2-4 weeks) before implementation:
- RFCs for daemon lifecycle, auth model, protobuf versioning, rollback
- Prototypes for transport, process model, state management
- Compatibility Charter definition
- Parity scorecard design
- Dual execution mode architecture

## Auto-Resolved Decisions

**Transport Default**: Unix Domain Socket (UDS) for local daemon
- Reason: Better security than TCP, no port conflicts, native on Unix
- Windows: Named pipes (fallback)

**Process Model**: Per-user singleton daemon
- Reason: Balance between isolation and resource sharing
- Autostart: On first CLI invocation (transparent to user)

**Migration Strategy**: Dual-path with feature flags
- Old path: Direct Kong CLI (preserved)
- New path: CLI → gRPC daemon (opt-in initially)
- Gradual default switch after parity proven

**Protobuf Organization**: Bounded service domains
- 14 namespaces → 14 proto packages
- Prevents monolithic proto from becoming unmaintainable

**Plugin Language Priority**: Go first, then Python
- Reason: Go for core plugins, Python for ML/image processing
- Delay: Rust, Node.js until ecosystem stabilizes

## Decisions Needing User Input

1. **Compatibility Window**: How many releases must maintain backward compatibility?
   - Options: 2 releases (~3 months) / 4 releases (~6 months) / 6 releases (~1 year)

2. **Default Switch Criteria**: When does new gRPC path become default?
   - Options: 100% parity / 95% parity + escape hatch / Never (parallel paths forever)

3. **Plugin Distribution**: How are plugins distributed/signed?
   - Options: GitHub releases + checksums / Built-in registry / Manual install only

4. **MCP Priority**: Is MCP/AI integration critical path or nice-to-have?
   - Options: Phase 2 critical / Phase 3 (after plugins) / Future roadmap only

## Work Plan Structure

**Phase 0** (Foundation & RFCs) - 2-4 weeks
**Phase 1** (Daemon Core) - 6-8 weeks  
**Phase 2** (MCP/AI) - 4-6 weeks
**Phase 3** (Plugins) - 6-8 weeks
**Phase 4** (Final Migration) - 2-4 weeks

Total: 20-30 weeks (5-7.5 months) with parallel workstreams

---

*This draft documents decisions made and open questions for the work plan.*
