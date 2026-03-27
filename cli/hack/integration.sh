#!/usr/bin/env bash
# Run CLI integration tests against a mock API server.
#
# Usage:
#   ./hack/integration.sh          # run all integration tests
#   ./hack/integration.sh -run Auth # run only auth-related tests
#   ./hack/integration.sh -v       # verbose output
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$CLI_DIR"

echo "==> Running integration tests"
go test -tags integration -count=1 -timeout 60s ./integration/ "$@"
echo "==> Done"
