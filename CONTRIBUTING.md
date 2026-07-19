# Contributing to Hawk

Thank you for your interest in contributing to Hawk.

We welcome bug reports, feature requests, documentation improvements, and code contributions.

## Getting Started

Clone the repository.

```bash
git clone https://github.com/lakshya-8000cr/hawk.git
cd hawk
```

Build the project.

```bash
go build ./cmd
```

Run Hawk locally.

```bash
kubectl hawk analyze deployment nginx -n default
```

---

## Development Guidelines

- Keep packages focused on a single responsibility.
- Prefer readable code over clever implementations.
- Avoid introducing unnecessary dependencies.
- Maintain the stateless execution model.
- Follow existing project conventions.

---

## Commit Messages

Use clear and descriptive commit messages.

Examples:

```
feat: add StatefulSet collector

fix: correct Service selector matching

docs: update architecture documentation
```

---

## Pull Requests

Before opening a Pull Request:

- Ensure the project builds successfully.
- Update documentation when required.
- Keep changes focused on a single feature or fix.
- Include sufficient testing information.

---

## Reporting Bugs

When reporting bugs, include:

- Kubernetes version
- kubectl version
- Hawk version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior