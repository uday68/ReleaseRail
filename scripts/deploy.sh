#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "usage: ./scripts/deploy.sh <environment> <image-tag> [--dry-run]" >&2
  exit 1
fi

environment="$1"
image_tag="$2"
dry_run="${3:-}"

if [[ "$dry_run" == "--dry-run" ]]; then
  echo "[dry-run] Deploying ${image_tag} to ${environment}"
  exit 0
fi

echo "Deploying ${image_tag} to ${environment}"
