# Installation

## Overview

Hawk is distributed as a native `kubectl` plugin and can be installed using Krew, downloaded as a prebuilt binary from the GitHub Releases page, or compiled directly from source.

The recommended installation method is **Krew**, the official package manager for `kubectl` plugins.

---

# Prerequisites

Before installing Hawk, ensure the following tools are available.

| Requirement | Version |
|------------|---------|
| Kubernetes | v1.24+ (recommended) |
| kubectl | v1.24+ |
| Go *(source build only)* | 1.22+ |
| Git *(source build only)* | Latest |

Verify your installation:

```bash
kubectl version --client
```

---

# Install via Krew (Recommended)

If Krew is already installed:

```bash
kubectl krew install hawk
```

Verify the installation:

```bash
kubectl hawk version
```

You should see the installed Hawk version.

---

# Install from GitHub Releases

Download the appropriate binary for your operating system from the project's GitHub Releases page.

Example:

```text
hawk-linux-amd64
hawk-darwin-arm64
hawk-windows-amd64.exe
```

Move the binary into a directory included in your system PATH.

Linux/macOS:

```bash
chmod +x hawk
sudo mv hawk /usr/local/bin/
```

Windows:

Place `hawk.exe` inside any directory already included in your PATH environment variable.

Verify:

```bash
hawk version
```

---

# Build from Source

Clone the repository.

```bash
git clone https://github.com/<username>/hawk.git
```

Enter the project directory.

```bash
cd hawk
```

Build the project.

```bash
go build -o hawk ./cmd
```

Run:

```bash
./hawk version
```

---

# Verify Installation

Confirm the plugin is correctly installed.

```bash
kubectl plugin list
```

You should see:

```text
hawk
```

Run a simple analysis.

```bash
kubectl hawk analyze deployment nginx -n default
```

If dependency analysis executes successfully, the installation is complete.

---

# Updating Hawk

For Krew installations:

```bash
kubectl krew upgrade hawk
```

For manual installations, download the latest release and replace the existing binary.

---

# Uninstalling

Krew:

```bash
kubectl krew uninstall hawk
```

Manual installation:

Remove the Hawk binary from your system PATH.

---

# Troubleshooting

### Hawk command not found

Ensure the binary exists in your system PATH.

---

### kubectl does not recognize Hawk

Verify the plugin appears in:

```bash
kubectl plugin list
```

If not, reinstall Hawk or check the binary location.

---

### Unable to connect to the cluster

Verify your Kubernetes context.

```bash
kubectl config current-context
```

Check cluster connectivity.

```bash
kubectl get nodes
```

---

### Permission denied

Verify the authenticated Kubernetes user has permission to read the required resources.

For example:

```bash
kubectl auth can-i get deployments
kubectl auth can-i get pods
kubectl auth can-i get services
```

---

# Next Steps

After installation, see the following documentation:

- `README.md` — Project overview and quick start.
- `docs/architecture.md` — High-level system architecture.
- `docs/internals.md` — Internal implementation details.
- `docs/design-decisions.md` — Engineering rationale.
- `docs/benchmark.md` — Performance evaluation.