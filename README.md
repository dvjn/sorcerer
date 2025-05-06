# Sorcerer

A minimal OCI (Open Container Initiative) registry server implemented in Go.

## Overview

Sorcerer is a lightweight container registry that complies with the OCI
distribution specification. It provides a simple way to store and distribute
container images privately.

## Features

- OCI-compliant container registry
- Simple configuration
- Minimal dependencies
- Lightweight design


## Usage

### Docker

The official docker image is available on GitHub Container Registry (ghcr.io).

```bash
# Pull the latest image
docker pull ghcr.io/dvjn/sorcerer:latest

# Run docker image
docker run -p 3000:3000 -v "$(pwd)/data:/app/data" ghcr.io/dvjn/sorcerer:latest
```

### CLI

You can build and install `sorcerer` from source using goblin:

```bash
# Build and install from main
curl -sf http://goblin.run/github.com/dvjn/sorcerer | CMD_PATH="/cmd/sorcerer" sh

# Run the server
sorcerer
```


## Configuration

Sorcerer can be configured using command-line flags or environment variables:

| Flag                       | Environment Variable     | Default | Description                                                  |
| -------------------------- | ------------------------ | ------- | ------------------------------------------------------------ |
| `--store-path`             | `STORE_PATH`             | `data`  | Path to store registry data                                  |
| `--port`                   | `PORT`                   | `3000`  | Port to run the server on                                    |
| `--auth-user-header`       | `AUTH_USER_HEADER`       | ``      | The header to use for fetching the authenticated user name   |
| `--auth-groups-header`     | `AUTH_GROUPS_HEADER`     | ``      | The header to use for fetching the authenticated user groups |
| `--auth-groups-header-sep` | `AUTH_GROUPS_HEADER_SEP` | `,`     | The separator for the user groups                            |


## License

This project is licensed under the [MIT License](LICENSE.txt).
