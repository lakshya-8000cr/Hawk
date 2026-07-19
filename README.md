<p align="center">
  <img src="img/Screenshot 2026-07-16 223815.png" alt="Hawk Banner" width="100%" height="80%">
</p>

# Hawk

> **Understand Kubernetes dependencies before production changes.**

Hawk is a **dependency analysis engine** packaged as a native `kubectl` plugin that discovers relationships between Kubernetes resources and evaluates the operational impact of modifying or deleting a workload.

Instead of manually tracing Deployments, ReplicaSets, Pods, Services, ConfigMaps, Secrets, PersistentVolumeClaims, and Ingresses, Hawk constructs a unified dependency graph and presents the workload's blast radius in a single report.

Designed for **Platform Engineers**, **DevOps Engineers**, **Site Reliability Engineers (SREs)**, and Kubernetes operators.

<p align="center">
  <!-- Demo GIF -->
</p>

<p align="center">
  <strong>Fast • Read-only • Zero Cluster Footprint • Krew Compatible</strong>
</p>


## Why Hawk?

Kubernetes exposes infrastructure as individual API objects.

Operators, however, think in terms of **applications**.

A single production workload often spans multiple Kubernetes resources:

- Deployments
- ReplicaSets
- Pods
- Services
- ConfigMaps
- Secrets
- PersistentVolumeClaims
- Ingresses

While `kubectl` allows you to inspect these resources individually, it does not explain **how they relate to one another** or what the operational impact of changing a workload might be.

Hawk bridges that gap by automatically discovering ownership relationships, building a dependency graph, and presenting the complete blast radius of a workload before changes are made.

<p align="center">
  <img src="img/hawk-dependency-tree.svg" alt="Hawk Dependency Tree" width="90%" height="90%">
</p>

## The Problem

Modern Kubernetes applications are composed of interconnected resources distributed across multiple API groups.

A Deployment may own ReplicaSets and Pods, expose traffic through Services and Ingresses, consume ConfigMaps and Secrets, and depend on PersistentVolumeClaims for storage.

Although Kubernetes stores these relationships internally, operators must manually correlate them using multiple `kubectl` commands before making production changes.

As clusters scale, manual dependency analysis becomes increasingly difficult, introducing unnecessary operational risk during deployments, upgrades, migrations, and incident response.

Before deleting or modifying a workload, engineers need clear answers to questions such as:

- Which Pods belong to this Deployment?
- Which Services expose these Pods?
- Is the workload externally accessible through an Ingress?
- Which ConfigMaps and Secrets are consumed?
- Does it rely on persistent storage?
- What is the overall operational blast radius?

Answering these questions manually is slow, repetitive, and error-prone.

---
## The Solution

Hawk performs dependency-aware analysis directly against the Kubernetes API using the official `client-go` library.

Starting from a target workload, Hawk recursively traverses Kubernetes ownership relationships, discovers dependent resources, and constructs an internal dependency graph representing the application's topology.

The graph is then evaluated by the Blast Radius Engine, which identifies operational dependencies such as exposed services, persistent storage, configuration resources, and sensitive secrets.

The final result is rendered as a structured terminal report that provides engineers with an immediate understanding of the workload's dependencies and potential operational impact.

<p align="center">
  <img src="img/hawk-cli-architecture-overview.svg" alt="Hawk Dependency Tree" width="80%" height="90%">
</p>


## Core Capabilities

Hawk provides a dependency-centric view of Kubernetes workloads through a native `kubectl` experience.

Key capabilities include:

- Automatic dependency discovery across supported Kubernetes resources
- Ownership traversal using Kubernetes `OwnerReferences`
- Dependency graph construction for workload analysis
- Blast radius evaluation for operational impact assessment
- Detection of Services, Ingresses, ConfigMaps, Secrets, and PersistentVolumeClaims
- Read-only analysis with zero modifications to cluster state
- Native `kubectl` plugin integration
- Cross-platform binaries with Krew support

---
## Installation

### Install with Krew (Recommended)

```bash
kubectl krew install hawk
```
Verify the installation:
```bash
kubectl hawk version
```

## Manual Installation
Download the latest release for your platform from the
Releases page and place the binary in your system's PATH as a
kubectl plugin.

For detailed installation instructions, see the [Installation Guide](docs/installation.md).
## Quick Start

Analyze a Deployment:
```bash
kubectl hawk analyze deployment nginx -n production
```
Analyze a Deployment in the default namespace:

```Bash
kubectl hawk analyze deployment nginx
```
Version
```bash
kubectl hawk version
```

Delete Deployment safely
```bash
kubectl hawk delete deployment nginx
```


One screenshot.

No multiple screenshots.

---

# Features

```markdown

-  Automatic dependency discovery for Kubernetes workloads
-  Dependency graph construction using Kubernetes ownership relationships
-  Detection of Services, ConfigMaps, Secrets, PersistentVolumeClaims, and Ingresses
-  Blast radius evaluation for operational impact assessment
-  Structured terminal reports optimized for production troubleshooting
-  Read-only analysis with zero modifications to cluster resources
-  Native `kubectl` plugin experience
-  Cross-platform binaries for Linux, macOS, and Windows
-  Krew-compatible distribution
```
## Supported Resources

| Kubernetes Resource | Discovery |
|---------------------|:---------:|
| Deployment | ✅ |
| ReplicaSet | ✅ |
| Pod | ✅ |
| Service | ✅ |
| ConfigMap | ✅ |
| Secret | ✅ |
| PersistentVolumeClaim | ✅ |
| Ingress | ✅ |


## Why not plain `kubectl`?

`kubectl` is excellent for interacting with Kubernetes resources individually.

Hawk complements `kubectl` by focusing on **resource relationships** rather than isolated objects, making dependency analysis significantly faster during production operations.

| Capability | `kubectl` | Hawk |
|------------|:---------:|:----:|
| List Kubernetes resources | ✅ | ✅ |
| Inspect individual objects | ✅ | ✅ |
| Automatic dependency discovery | ❌ | ✅ |
| Ownership traversal | ❌ | ✅ |
| Unified dependency graph | ❌ | ✅ |
| Blast radius evaluation | ❌ | ✅ |
| Production impact analysis | ❌ | ✅ |

---
## Design Goals

Hawk was built around a small set of engineering principles:

- **Zero Cluster Footprint** — No agents, controllers, CRDs, or admission webhooks.
- **Read-only by Design** — Cluster resources are never modified.
- **Native Kubernetes APIs** — Built on the official `client-go` library.
- **Deterministic Discovery** — Relationships are derived from Kubernetes metadata instead of heuristics.
- **Modular Architecture** — Independent collectors simplify maintenance and future extensions.
- **Production-first** — Designed to support operational decision making before infrastructure changes.

## Performance

Hawk has been validated against a synthetic Kubernetes environment containing more than **3,000 Kubernetes resources**, including Deployments, ReplicaSets, Pods, Services, ConfigMaps, Secrets, PersistentVolumeClaims, and Ingresses.

The dependency discovery pipeline performs read-only analysis using the Kubernetes API and is designed to scale efficiently for production environments.

Detailed benchmarking methodology and results are available in [`docs/benchmark.md`](docs/benchmark.md).


---
## Documentation

Additional technical documentation is available in the `docs/` directory.

| Document | Description |
|----------|-------------|
| Architecture | High-level system architecture and execution flow [`docs/architecture.md`](docs/architecture.md)
| Design Decisions | Engineering decisions and trade-offs [`docs/design-decision.md`](docs/design-decision.md)
| Installation | Platform-specific installation instructions [`docs/installation.md`](docs/installation.md)
| Benchmarks | Performance evaluation methodology and results [`docs/benchmark.md`](docs/benchmark.md)
| Internals | Dependency discovery pipeline and package structure [`docs/internals.md`](docs/internals.md)


---
## Contributions

Contributions are welcome.

If you discover a bug, have an idea for an enhancement, or would like to contribute code, please open an issue or submit a pull request.

Please review the contribution guidelines before submitting changes.

## License

This project is licensed under the MIT License. See the [`LICENSE`](LICENSE). file for details.
