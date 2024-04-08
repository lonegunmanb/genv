# `genv` - A CLI tool for managing environments

`genv` is a command-line interface (CLI) tool designed to manage different environments. It allows users to create their own repositories using this template repository, and control the installation and version switching of binaries.

## How to Use

To use `genv`, follow the steps below:

1. Clone this template repository to create your own repository.
2. Navigate into the `generate` directory.
3. Run the `go run` command with the appropriate flags. For example:

```bash
go run main.go -u "https://releases.hashicorp.com/vault/{{ .Version }}/vault_{{ .Version }}_{{ .Os }}_{{ .Arch }}.zip" -n vaultenv -b vault --git-repo https://github.com/hashicorp/vault.git 
```

In this command:

- `-u` specifies the download URL template.
- `-n` specifies the control plane binary name.
- `-b` specifies the binary name.
- `--git-repo` specifies the github repository url when download install fail and fallback to use go build to install

This command will install two binaries: `vaultenv` and `vault`.

- `vaultenv` is the control plane. It can be used to install the binary and switch versions.
- `vault` is a dummy binary that forwards all flags to the actual binary that the control plane downloaded.

For example, you can install vault 1.6.0 by:

```shell
vaultenv install 1.6.0
```

Then you can switch to 1.6.0:

```shell
vaultenv use 1.6.0
```

Then you can try vault:

```shell
vault -v
Vault v1.6.0
```

## Features

- **Environment Management**: `genv` allows you to manage different environments with ease. You can switch between different versions of a binary without any hassle.
- **Binary Installation**: `genv` provides a straightforward way to install binaries. You can specify the download URL, home directory, environment name, and binary name to install the binary.
- **Version Switching**: With `genv`, you can easily switch between different versions of a binary. The control plane (`vaultenv` in the example) handles the version switching.

## Getting Started

To get started with `genv`, clone this template repository and navigate into the `generate` directory. Then, run the `go run` command with the appropriate flags to install the binaries and manage your environments.

To run the test, you must install [gomock](https://github.com/uber-go/mock) first:

```shell
go install go.uber.org/mock/mockgen@latest
```

Then run:

```shell
go generate ./...
```