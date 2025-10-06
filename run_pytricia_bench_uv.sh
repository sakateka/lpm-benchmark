#!/usr/bin/env bash
set -euo pipefail

# This script sets up a uv-based virtual environment, installs Python dependencies,
# and runs the PyTricia 1M benchmark.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$SCRIPT_DIR"
VENV_DIR="$REPO_ROOT/.venv-pytricia"
REQ_FILE="$REPO_ROOT/requirements-pytricia.txt"
BENCH_SCRIPT="$REPO_ROOT/py_bench_pytricia_1m.py"

if [[ ! -f "$BENCH_SCRIPT" ]]; then
  echo "Benchmark script not found at $BENCH_SCRIPT" >&2
  exit 1
fi

# Find or install uv
UV_BIN=""
if command -v uv >/dev/null 2>&1; then
  UV_BIN="$(command -v uv)"
elif [[ -x "$HOME/.local/bin/uv" ]]; then
  UV_BIN="$HOME/.local/bin/uv"
elif [[ -x "$HOME/.cargo/bin/uv" ]]; then
  UV_BIN="$HOME/.cargo/bin/uv"
else
  echo "uv not found; install it using: curl -fsSL https://astral.sh/uv/install.sh | sh" >&2
  exit 1
fi

echo "Using uv at: $UV_BIN"

# Ensure requirements file exists
if [[ ! -f "$REQ_FILE" ]]; then
  cat > "$REQ_FILE" << 'EOF'
pytricia
psutil
EOF
fi

# Create venv (idempotent)
"$UV_BIN" venv "$VENV_DIR"

# Install requirements into venv
"$UV_BIN" pip install -r "$REQ_FILE" --python "$VENV_DIR/bin/python"

# Run benchmark (defaults: both families, 1M prefixes, 5M probes)
"$VENV_DIR/bin/python" "$BENCH_SCRIPT" --family both --count 1000000 --lookup-probes 5000000 --lookup-set-size 1000
