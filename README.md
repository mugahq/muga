# muga

Observability for developers who live in the terminal.

## Overview

Muga is a monorepo containing SDKs for integrating observability into your applications.

## Repository Structure

```
cli/          # Go CLI (muga command)
sdks/
├── python/   # Python SDK (PyPI: muga)
└── node/     # Node.js SDK (coming soon)
```

## CLI

The `muga` CLI is written in Go and follows the `muga <noun> <verb>` command pattern.

Requires Go 1.26+.

```bash
cd cli
make build
./bin/muga
```

## SDKs

| SDK    | Version | Status   |
|--------|---------|----------|
| Python | 0.0.1   | Planning |
| Node   | —       | Planned  |

### Python

```bash
pip install muga
```

Requires Python 3.9+.

## Development

Clone the repository:

```bash
git clone https://github.com/mugahq/muga.git
cd muga
```

## License

[MIT](LICENSE)