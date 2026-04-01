#!/usr/bin/env bash
set -euo pipefail

cleanup() {
  local exit_code=$?

  trap - EXIT INT TERM

  if [[ -n "${api_pid:-}" ]]; then
    kill "$api_pid" 2>/dev/null || true
    wait "$api_pid" 2>/dev/null || true
  fi

  if [[ -n "${worker_pid:-}" ]]; then
    kill "$worker_pid" 2>/dev/null || true
    wait "$worker_pid" 2>/dev/null || true
  fi

  exit "$exit_code"
}

trap cleanup EXIT INT TERM

go run ./cmd/go-jarvis &
api_pid=$!

go run ./cmd/worker &
worker_pid=$!

while true; do
  if ! kill -0 "$api_pid" 2>/dev/null; then
    wait "$api_pid"
    exit $?
  fi

  if ! kill -0 "$worker_pid" 2>/dev/null; then
    wait "$worker_pid"
    exit $?
  fi

  sleep 1
done
