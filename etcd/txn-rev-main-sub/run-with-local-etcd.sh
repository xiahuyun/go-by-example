#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP_DIR="${SCRIPT_DIR}/.tmp"
ETCD_DATA_DIR="${ETCD_DATA_DIR:-${TMP_DIR}/etcd-data}"
ETCD_LOG_FILE="${ETCD_LOG_FILE:-${TMP_DIR}/etcd.log}"
ETCD_SNAPSHOT="${ETCD_SNAPSHOT:-${TMP_DIR}/txn-rev-snapshot.db}"

ETCD_BIN="${ETCD_BIN:-etcd}"
ETCD_HOST="${ETCD_HOST:-127.0.0.1}"
ETCD_PORT="${ETCD_PORT:-12379}"
ETCD_PEER_PORT="${ETCD_PEER_PORT:-12380}"
ETCD_ENDPOINT="${ETCD_HOST}:${ETCD_PORT}"
WAIT_SECONDS="${WAIT_SECONDS:-20}"
KEEP_ETCD_DATA="${KEEP_ETCD_DATA:-0}"

ETCD_PID=""
STARTED_BY_SCRIPT=0

cleanup() {
  if [[ "${STARTED_BY_SCRIPT}" == "1" && -n "${ETCD_PID}" ]] && kill -0 "${ETCD_PID}" 2>/dev/null; then
    echo
    echo "[cleanup] stopping etcd pid=${ETCD_PID}"
    kill "${ETCD_PID}" || true
    wait "${ETCD_PID}" 2>/dev/null || true
  fi

  if [[ "${STARTED_BY_SCRIPT}" == "1" && "${KEEP_ETCD_DATA}" != "1" ]]; then
    rm -rf "${ETCD_DATA_DIR}"
  fi
}
trap cleanup EXIT

if ! command -v "${ETCD_BIN}" >/dev/null 2>&1; then
  echo "error: etcd binary not found."
  echo "please install etcd first, or set ETCD_BIN to an executable path."
  exit 1
fi

mkdir -p "${TMP_DIR}"

if curl -fsS "http://${ETCD_ENDPOINT}/health" >/dev/null 2>&1; then
  echo "[info] reuse existing etcd endpoint: ${ETCD_ENDPOINT}"
else
  echo "[info] starting temporary etcd at ${ETCD_ENDPOINT}"
  rm -rf "${ETCD_DATA_DIR}"

  "${ETCD_BIN}" \
    --name txn-rev-demo \
    --data-dir "${ETCD_DATA_DIR}" \
    --listen-client-urls "http://${ETCD_ENDPOINT}" \
    --advertise-client-urls "http://${ETCD_ENDPOINT}" \
    --listen-peer-urls "http://${ETCD_HOST}:${ETCD_PEER_PORT}" \
    --initial-advertise-peer-urls "http://${ETCD_HOST}:${ETCD_PEER_PORT}" \
    --initial-cluster "txn-rev-demo=http://${ETCD_HOST}:${ETCD_PEER_PORT}" \
    --initial-cluster-state new \
    --initial-cluster-token txn-rev-demo-token \
    >"${ETCD_LOG_FILE}" 2>&1 &

  ETCD_PID=$!
  STARTED_BY_SCRIPT=1

  start_ts=$(date +%s)
  while true; do
    if curl -fsS "http://${ETCD_ENDPOINT}/health" >/dev/null 2>&1; then
      echo "[info] etcd is healthy"
      break
    fi

    now_ts=$(date +%s)
    if (( now_ts - start_ts >= WAIT_SECONDS )); then
      echo "error: etcd did not become healthy within ${WAIT_SECONDS}s"
      echo "see logs: ${ETCD_LOG_FILE}"
      exit 1
    fi
    sleep 1
  done
fi

echo "[info] running txn revision demo..."
(
  cd "${SCRIPT_DIR}"
  ETCD_ENDPOINTS="${ETCD_ENDPOINT}" ETCD_SNAPSHOT="${ETCD_SNAPSHOT}" go run .
)

echo "[done] txn revision demo completed"
