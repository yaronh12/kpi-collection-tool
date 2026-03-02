# Installation

## Without Cloning (recommended)

Install directly from GitHub using `go install`:

```bash
go install github.com/redhat-best-practices-for-k8s/kpi-collection-tool/cmd/kpi-collector@latest
```

Verify installation:

```bash
kpi-collector --help
```

If not found, add `~/go/bin` to your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

## From Source

Clone the repo and install locally:

```bash
git clone https://github.com/redhat-best-practices-for-k8s/kpi-collection-tool.git
cd kpi-collection-tool
make install
```

## Uninstall

```bash
make uninstall
```
