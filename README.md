# Runtipi CLI - Manage your Runtipi instance from the command line!

[![License](https://img.shields.io/github/license/runtipi/cli)](https://github.com/runtipi/cli/blob/main/LICENSE)
[![Version](https://img.shields.io/github/v/release/runtipi/cli?color=%235351FB&label=version)](https://github.com/runtipi/cli/releases)
![Issues](https://img.shields.io/github/issues/runtipi/cli)

> [!NOTE]
> Runtipi CLI is written in Go! If you want to collaborate on a cool project, join the discussion on Discord!

Runtipi CLI is an updated version of the previous CLI implementations. Written in Go, this version offers improved performance, reliability, and a smaller footprint. It provides a streamlined experience for managing your Runtipi instance directly from the command line. To get started, follow the instructions below.

## Getting Started

The CLI is a core component of Runtipi, so it already exists in your installation. To use it simply run:

```bash
./runtipi-cli help
```

CLI authentication uses a private key created by the backend at `/data/.secrets/cli-secret` in the persistent Docker-managed `runtipi_auth` volume. Upgrade the CLI and backend together and recreate the backend with the updated Compose file before running authenticated CLI commands. The CLI reads the key through `docker exec`, so Docker-enabled users can authenticate without reading root-owned host files. Keep this volume when recreating the backend; deleting it invalidates existing CLI tokens.

Shared app/configuration permissions, installation-root modes, and `--no-permissions` retain their existing behavior. Existing `JWT_SECRET` and seed values remain in use for encryption.

## 🔨 Building locally

If you would like to build the CLI locally on your own machine, you will need to have Go installed on your system. Then you can clone the repository with:

```bash
git clone --depth 1 https://github.com/runtipi/cli
```

Then you can build the CLI with:

```bash
make build
```

Or you can run it directly with:

```bash
make run ARGS="command arguments"
```

The built CLI binary will be in the root folder named `runtipi-cli`.

## Dependency updates

Renovate opens dependency update pull requests for Go modules (including indirect dependencies), the Go toolchain, GitHub Actions, and container images. CI and release workflows read the Go version from `go.mod` so toolchain updates stay consistent.

The embedded `internal/assets/docker-compose.yml.tmpl` uses a custom image extractor because it contains Go template directives. It tracks explicit release tags for Traefik, PostgreSQL, and RabbitMQ while leaving the runtime-selected `${TIPI_VERSION}` unchanged. PostgreSQL stays on major version 14; moving to another major requires a separate data migration.

New releases wait two days, or three days for Go dependencies; GitHub Actions digest pins and refreshes are exempt. Go module minor and patch updates are grouped. Renovate proposes commit-digest pins for GitHub Actions, and merging remains manual.

## 🗣 Community

- [Twitter](https://twitter.com/runtipi)
- [Discord](https://discord.gg/Bu9qEPnHsc)

## ❤️ Contributing

We welcome contributions to the Runtipi CLI! If you have Go programming experience or ideas to improve the CLI, feel free to submit pull requests or open issues. Your name will be added to the "Contributors" section in our main repository when your changes get merged.

## 📜 License

Runtipi CLI is licensed under the GNU General Public License v3.0. TL;DR — You may copy, distribute and modify the software as long as you track changes/dates in source files. Any modifications to or software including (via compiler) GPL-licensed code must also be made available under the GPL along with build & install instructions.
