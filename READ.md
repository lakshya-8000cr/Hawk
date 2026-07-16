<p align="center">
  <img src="docs/assets/banner.png" alt="Hawk Banner" width="100%">
</p>

<h1 align="center">Hawk</h1>

<p align="center">
Dependency-aware Kubernetes impact analysis.
</p>


---

## Overview

Hawk is a Kubernetes CLI plugin that performs dependency-aware impact analysis before destructive cluster operations.

Instead of inspecting Kubernetes resources independently, Hawk reconstructs runtime relationships between workloads by traversing ownership references, selector-based associations, and networking dependencies to generate a directed dependency graph representing the operational topology of the cluster.

The resulting graph is used to calculate blast radius, dependency chains, and infrastructure risk before resources are modified or removed.

---


## Why Hawk?

Deleting a Kubernetes resource is rarely an isolated operation.

A single Deployment may indirectly affect:

- ReplicaSets
- Pods
- Services
- Ingress traffic
- HorizontalPodAutoscalers
- PodDisruptionBudgets
- ConfigMaps
- Secrets
- Persistent Volumes

Native `kubectl delete` performs the requested operation but provides little visibility into downstream operational impact.

Hawk performs dependency discovery before deletion, allowing engineers to understand infrastructure relationships before modifying a running cluster.
---
## Features

- Deployment dependency discovery
- ReplicaSet ownership traversal
- Pod ownership analysis
- Service selector resolution
- Ingress backend discovery
- Directed dependency graph construction
- Blast radius analysis
- Risk classification

---

## Architecture

```
                     Kubernetes Cluster
                            │
                     Kubernetes API Server
                            │
                     client-go API Client
                            │
                  Repository Abstraction Layer
                            │
                  Resource Discovery Engine
                            │
                 Directed Dependency Graph
                            │
                  Blast Radius Analyzer
                            │
                     Terminal Renderer
```

---

## Dependency Graph

```
                 Ingress
                    │
              ROUTES_TO
                    │
                 Service
                    │
               SELECTS
                    │
                   Pod
                    │
                 OWNED BY
                    │
               ReplicaSet
                    │
                 OWNED BY
                    │
               Deployment
```

Unlike ownership trees, Hawk models Kubernetes objects as graph vertices connected through semantic relationships, enabling graph traversal for dependency analysis and future graph-based reasoning.

---

## Installation

### Build from source

```bash
git clone https://github.com/lakshya-8000cr/Hawk

cd Hawk

go build -o kubectl-hawk
```

Linux

```bash
sudo mv kubectl-hawk /usr/local/bin/
```

Windows

Move `kubectl-hawk.exe` into any directory present in your `PATH`.

Verify installation

```bash
kubectl plugin list

kubectl hawk --help
```

---

## Usage

Analyze a Deployment

```bash
kubectl hawk analyze deployment frontend
```

Specify namespace

```bash
kubectl hawk analyze deployment frontend \
    --namespace production
```

---

## Example Output

```text
HAWK   Impact Analysis

Target:
Deployment/frontend

Directly owned

├── ReplicaSet/frontend-84dcb97
│
└── Pod/frontend-84dcb97-52hf2

Referenced by

└── Service/frontend

Traffic Exposure

└── Ingress/frontend

Blast Radius

Risk: HIGH

External traffic reaches this workload.
```

---

## Design Principles

### Graph-first analysis

Infrastructure relationships are modeled as directed graph edges rather than recursively traversed resource trees.

### Repository abstraction

Kubernetes resource acquisition is isolated behind repository interfaces, allowing alternate transport implementations without modifying domain logic.

### Domain isolation

Business logic operates on internal domain models instead of Kubernetes SDK objects, reducing framework coupling.

### Layered architecture

```
CLI

↓

Analyzer

↓

Repositories

↓

client-go

↓

Kubernetes API
```

Each layer owns a single responsibility.

---

## Technology

- Go
- Cobra
- Kubernetes client-go
- Directed Graph Model
- Repository Pattern
- Domain-driven Design
- Layered Architecture

---

## Roadmap

- [x] Deployment analysis
- [x] ReplicaSet traversal
- [x] Pod discovery
- [x] Service dependency discovery
- [x] Ingress dependency discovery
- [x] Blast radius analysis

Upcoming

- [ ] HPA dependency analysis
- [ ] PodDisruptionBudget support
- [ ] ConfigMap relationships
- [ ] Secret relationships
- [ ] PVC dependency analysis
- [ ] Interactive delete confirmation
- [ ] Graph export
- [ ] Krew distribution

---

## License

MIT
