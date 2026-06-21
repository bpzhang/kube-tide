#!/usr/bin/env bash
# Upgrade Go and web dependencies to their latest versions.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WEB_DIR="${ROOT}/web"

UPGRADE_GO=true
UPGRADE_WEB=true
RUN_VENDOR=false

usage() {
	cat <<'EOF'
Usage: scripts/upgrade-deps.sh [options]

Upgrade Go module dependencies and web (pnpm) dependencies to latest versions.

Options:
  --go-only     Upgrade Go dependencies only
  --web-only    Upgrade web dependencies only
  --vendor      Run "go mod vendor" after upgrading Go deps
  -h, --help    Show this help message

Examples:
  scripts/upgrade-deps.sh
  scripts/upgrade-deps.sh --go-only --vendor
  scripts/upgrade-deps.sh --web-only
EOF
}

while [[ $# -gt 0 ]]; do
	case "$1" in
	--go-only)
		UPGRADE_WEB=false
		;;
	--web-only)
		UPGRADE_GO=false
		;;
	--vendor)
		RUN_VENDOR=true
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "Unknown option: $1" >&2
		usage >&2
		exit 1
		;;
	esac
	shift
done

require_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "Error: required command not found: $1" >&2
		exit 1
	fi
}

upgrade_go() {
	require_cmd go

	echo "==> Upgrading Go dependencies to latest..."
	cd "$ROOT"

	# Bypass vendor/ when resolving modules; required when vendor directory exists.
	# go get does not accept -mod directly; use GOFLAGS instead.
	local prev_goflags="${GOFLAGS:-}"
	export GOFLAGS=-mod=mod

	local module_path
	module_path="$(go list -m -f '{{.Path}}')"

	found=false
	while IFS= read -r mod; do
		[[ -z "$mod" || "$mod" == "$module_path" ]] && continue
		found=true
		echo "  go get ${mod}@latest"
		go get "${mod}@latest"
	done < <(go list -m -f '{{if not .Indirect}}{{.Path}}{{end}}' all)

	if [[ -n "$prev_goflags" ]]; then
		export GOFLAGS="$prev_goflags"
	else
		unset GOFLAGS
	fi

	if [[ "$found" == false ]]; then
		echo "No direct Go modules found."
	fi

	echo "==> Tidying Go modules..."
	go mod tidy

	if [[ "$RUN_VENDOR" == true || -d vendor ]]; then
		echo "==> Vendoring Go modules..."
		go mod vendor
	fi

	echo "==> Go dependencies upgraded."
}

upgrade_web() {
	require_cmd pnpm

	if [[ ! -f "${WEB_DIR}/package.json" ]]; then
		echo "Error: web/package.json not found" >&2
		exit 1
	fi

	echo "==> Upgrading web dependencies to latest..."
	cd "$WEB_DIR"
	pnpm update --latest
	echo "==> Web dependencies upgraded."
}

if [[ "$UPGRADE_GO" == false && "$UPGRADE_WEB" == false ]]; then
	echo "Error: nothing to upgrade; use --go-only or --web-only, not both" >&2
	exit 1
fi

if [[ "$UPGRADE_GO" == true ]]; then
	upgrade_go
fi

if [[ "$UPGRADE_WEB" == true ]]; then
	upgrade_web
fi

echo "Done."
