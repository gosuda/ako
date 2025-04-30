# Ako

## Prerequisites

- [Go](https://golang.org/doc/install/source) 1.24 or later
- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
- [K3D](https://k3d.io/) if you use K3S
- [Docker](https://docs.docker.com/get-docker/) if you use K3S
- [Kubernetes](https://kubernetes.io/docs/setup/) if you use K3S

## Installation

```bash
go install github.com/gosuda/ako@latest
```

# Usage

```bash
ako
```

```shell
NAME:
   ako - Manage your Go project with ako

USAGE:
   ako [global options] [command [command options]]

COMMANDS:
   init, i    Initialize a new Go module and Git repository
   go, g      Organize Go project
   branch, b  Organize Git branches and commits
   k3d, k     Manage K3S manifests and clusters
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

## References

- [Korean](./principle_ko.md)
- [English](./principle_en.md)