# Installation

Dex ships as a single binary without any external dependencies making the installation very simple.

Dex is available for all major platforms (e.g., macOS, Windows, Linux, OpenBSD, FreeBSD, etc.)

Approaches to install Dex:

1. Using a [pre-compiled binary](#binary-cross-platform)
2. Installing with [package manager](#homebrew)
3. Installing from [source](#building-from-source)
4. Installing with [Docker](#using-docker-image)

## Binary (Cross-platform)

Download the appropriate version for your platform from [releases](https://github.com/odpf/dex/releases) page. Once
downloaded, the binary can be run from anywhere. You don’t need to install it into a global location. This works well
for shared hosts and other systems where you don’t have a privileged account. Ideally, you should install it somewhere
in your PATH for easy use. `/usr/local/bin` is the most probable location.

## Homebrew

```sh
# Install dex (requires homebrew installed)
brew install odpf/taps/dex

# Upgrade dex (requires homebrew installed)
brew upgrade dex

# Check for installed dex version
dex version
```

## Building from source

To compile from source, you will need [Go](https://golang.org/) installed in your `PATH`.

```bash
# Clone the repo
https://github.com/odpf/dex.git

# Build dex binary file
make build

# Check for installed dex version
./dex version
```

## Using Docker image

Dex ships a Docker image [odpf/dex](https://hub.docker.com/r/odpf/dex) that enables you to use `dex` as part of your Docker workflow.

For example, you can run `dex version` with this command:

```bash
docker run odpf/dex version
```
