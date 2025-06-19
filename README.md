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

## 🗣 Community

- [Twitter](https://twitter.com/runtipi)
- [Discord](https://discord.gg/Bu9qEPnHsc)

## ❤️ Contributing

We welcome contributions to the Runtipi CLI! If you have Go programming experience or ideas to improve the CLI, feel free to submit pull requests or open issues. Your name will be added to the "Contributors" section in our main repository when your changes get merged.

## 📜 License

Runtipi CLI is licensed under the GNU General Public License v3.0. TL;DR — You may copy, distribute and modify the software as long as you track changes/dates in source files. Any modifications to or software including (via compiler) GPL-licensed code must also be made available under the GPL along with build & install instructions.
