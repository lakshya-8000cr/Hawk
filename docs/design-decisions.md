# Hawk Design Decisions

## Overview

Hawk is intentionally designed as a lightweight, stateless dependency analysis tool for Kubernetes. Every architectural decision prioritizes simplicity, determinism, and compatibility with upstream Kubernetes while minimizing operational overhead.

This document explains the major engineering decisions behind Hawk, the alternatives that were considered, and the trade-offs accepted during implementation. For the resulting system design, see [architecture.md](architecture.md); for implementation details, see [internals.md](internals.md).

## Table of Contents

- [Why a `kubectl` Plugin?](#why-a-kubectl-plugin)
- [Why a Stateless Architecture?](#why-a-stateless-architecture)
- [Why Use `client-go`?](#why-use-client-go)
- [Why Not Use an Operator?](#why-not-use-an-operator)
- [Why No Custom Resource Definitions (CRDs)?](#why-no-custom-resource-definitions-crds)
- [Why OwnerReferences?](#why-ownerreferences)
- [Why Build an In-Memory Dependency Graph?](#why-build-an-in-memory-dependency-graph)
- [Why Separate Collectors?](#why-separate-collectors)
- [Why Read-Only API Operations?](#why-read-only-api-operations)
- [Why Terminal Output First?](#why-terminal-output-first)
- [Engineering Trade-offs](#engineering-trade-offs)
- [Future Architectural Evolution](#future-architectural-evolution)

---

## Why a `kubectl` Plugin?

Hawk is implemented as a native `kubectl` plugin rather than a standalone application.

### Rationale

- Integrates naturally with existing Kubernetes workflows.
- Reuses the user's existing kubeconfig and authentication.
- Requires no additional installation beyond the plugin itself.
- Behaves consistently with the Kubernetes CLI ecosystem.

### Alternatives Considered

- Standalone CLI
- Web dashboard
- Kubernetes Operator

### Trade-offs

**Advantages**

- Minimal operational overhead.
- Familiar user experience.
- No long-running processes.

**Limitations**

- Analysis is performed only when explicitly invoked.
- No continuous monitoring or background reconciliation.

## Why a Stateless Architecture?

Every Hawk execution is treated as an independent analysis session. No execution state, dependency graph, or cache is retained after the analysis completes.

### Rationale

Stateless execution guarantees that every analysis reflects the current state of the Kubernetes cluster without requiring cache invalidation or synchronization logic.

### Benefits

- Fresh results on every execution.
- No cache consistency issues.
- Simpler implementation.
- Lower memory footprint.

### Trade-offs

Repeated executions perform fresh API discovery instead of reusing cached information.

## Why Use `client-go`?

Hawk communicates exclusively through the official Kubernetes `client-go` library.

### Rationale

Using the upstream Kubernetes SDK ensures compatibility with Kubernetes API semantics, authentication mechanisms, resource versioning, and future Kubernetes releases.

### Alternatives Considered

- Raw REST requests
- Third-party Kubernetes SDKs

### Decision

The official SDK provides stronger type safety, better long-term compatibility, and reduced maintenance overhead.

## Why Not Use an Operator?

One of the earliest design decisions was to avoid implementing Hawk as a Kubernetes Operator.

### Rationale

Operators are designed for continuous reconciliation and automated resource management. Hawk performs neither of these responsibilities — instead, it executes short-lived dependency analysis on demand. Introducing an Operator would increase operational complexity without improving the core analysis workflow.

### Trade-offs

Operator-based implementations could support continuous monitoring but would require:

- Controllers
- CRDs
- RBAC configuration
- Persistent runtime components

These requirements conflict with Hawk's goal of remaining lightweight.

## Why No Custom Resource Definitions (CRDs)?

Hawk intentionally avoids introducing Custom Resource Definitions.

### Rationale

Dependency analysis does not require persistent cluster-side state — all required information already exists within standard Kubernetes resources. Avoiding CRDs keeps installation simple while maintaining compatibility with managed Kubernetes platforms.

## Why OwnerReferences?

Workload discovery begins by traversing Kubernetes `OwnerReferences`.

### Rationale

`OwnerReferences` represent authoritative ownership relationships maintained by Kubernetes itself. Unlike naming conventions or manually assigned labels, they accurately describe resource lineage.

### Benefits

- Deterministic traversal.
- Stable across clusters.
- Native Kubernetes metadata.
- No heuristic assumptions.

### Trade-offs

Resources without ownership metadata require alternative discovery strategies.

## Why Build an In-Memory Dependency Graph?

After resource discovery completes, Hawk constructs an in-memory directed dependency graph.

### Rationale

The graph provides a canonical representation of workload relationships that can be reused by multiple analysis stages. Instead of repeatedly querying Kubernetes resources, later stages consume the graph directly.

### Benefits

- Eliminates redundant traversal.
- Simplifies blast radius evaluation.
- Enables future graph-based analysis.
- Separates discovery from analysis.

### Trade-offs

Graph construction introduces a small memory overhead proportional to the number of discovered resources. Since the graph exists only for the duration of execution, this overhead remains predictable and bounded.

## Why Separate Collectors?

Each Kubernetes resource type is discovered by an independent collector.

### Rationale

Separating collectors follows the Single Responsibility Principle and keeps resource-specific logic isolated. Adding support for a new Kubernetes resource typically requires implementing only a new collector, without modifying existing discovery logic.

### Benefits

- Easier testing.
- Better maintainability.
- Lower coupling.
- Simpler future extensions.

## Why Read-Only API Operations?

Hawk never creates, modifies, patches, or deletes Kubernetes resources.

### Rationale

Dependency analysis should never alter production infrastructure. Restricting the tool to read-only API operations guarantees that executing Hawk cannot unintentionally affect cluster state.

### Benefits

- Safe for production clusters.
- Minimal RBAC permissions.
- Predictable execution.

## Why Terminal Output First?

The initial implementation focuses on human-readable terminal output.

### Rationale

Most operational workflows begin with engineers investigating production systems directly from the command line. Optimizing the CLI experience provides immediate value while establishing a foundation for future export formats.

### Future Evolution

The formatter is intentionally isolated so additional output formats — JSON, YAML, Graphviz, HTML — can be introduced without modifying the dependency analysis pipeline.

## Engineering Trade-offs

Like every software system, Hawk makes deliberate trade-offs.

| Decision | Benefit | Trade-off |
|---|---|---|
| Stateless execution | Always reflects live cluster state | No cached analysis |
| `kubectl` plugin | Simple installation | Manual execution |
| Read-only API | Safe for production | Cannot automate remediation |
| `OwnerReference` traversal | Deterministic discovery | Depends on Kubernetes metadata |
| In-memory graph | Fast analysis | Temporary memory usage |
| No CRDs | Zero cluster footprint | No persistent analysis history |

## Future Architectural Evolution

The current architecture intentionally favors simplicity over feature breadth. Future enhancements may include:

- Namespace-wide dependency analysis.
- Graph export formats.
- JSON output.
- Interactive terminal interface.
- Additional Kubernetes resource collectors.

These enhancements can be implemented without redesigning the existing architecture, because Hawk's execution pipeline is modular, stateless, and loosely coupled.