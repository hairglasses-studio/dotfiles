#!/usr/bin/env bash
# local boot override: remove dashboard profile from boot path
set -euo pipefail

ROOT_DIR="${JELLYFIN_STACK_ROOT:-${HOME}/jellyfin-stack}"
ENV_FILE="${ROOT_DIR}/.env"
STABLE_PROFILES=(core fetch tooling music reading transcode channels automation)
EXPERIMENTAL_PROFILES=(experimental)
BOOTSTRAP_DISABLED_SERVICES=(gluetun jdownloader)
include_experimental=0

usage() {
  cat <<'USAGE'
Usage:
  jellyfin-stack-boot.sh [--env-file PATH] [--include-experimental]

Starts stable Jellyfin stack at boot without dashboard profile.

If WireGuard placeholder values are still present, boot falls back to the
temporary qBittorrent bootstrap path and skips VPN-bound fetch services that
cannot start without a real Gluetun configuration.

Experimental helpers stay gated by default. Pass --include-experimental or set
BOOT_STACK_INCLUDE_EXPERIMENTAL=1 to opt them in explicitly.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --env-file)
      ENV_FILE="$2"
      shift 2
      ;;
    --include-experimental)
      include_experimental=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 1
      ;;
  esac
done

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "Missing env file: ${ENV_FILE}" >&2
  exit 1
fi

if [[ "${BOOT_STACK_INCLUDE_EXPERIMENTAL:-0}" =~ ^(1|true|TRUE|yes|YES|on|ON)$ ]]; then
  include_experimental=1
fi

PROFILES=("${STABLE_PROFILES[@]}")
if [[ "${include_experimental}" -eq 1 ]]; then
  PROFILES+=("${EXPERIMENTAL_PROFILES[@]}")
fi

placeholder_value() {
  local value="${1:-}"
  [[ -z "${value}" ]] && return 0
  [[ "${value}" == "__CHANGE_ME__" ]] && return 0
  [[ "${value}" == "__change_me__" ]] && return 0
  [[ "${value}" == CHANGE_ME ]] && return 0
  [[ "${value}" == change_me ]] && return 0
  return 1
}

needs_fetch_bootstrap=0
if (
  set -a
  # shellcheck disable=SC1090
  source "${ENV_FILE}"
  set +a

  if placeholder_value "${WIREGUARD_PRIVATE_KEY:-}" ||
     placeholder_value "${WIREGUARD_ADDRESSES:-}" ||
     placeholder_value "${WIREGUARD_PUBLIC_KEY:-}" ||
     placeholder_value "${WIREGUARD_ENDPOINT_IP:-}" ||
     placeholder_value "${WIREGUARD_ENDPOINT_PORT:-}"; then
    exit 0
  fi

  exit 1
); then
  needs_fetch_bootstrap=1
fi

compose_args=(--env-file "${ENV_FILE}" -f "${ROOT_DIR}/compose.yaml")
preflight_args=(--env-file "${ENV_FILE}")
if [[ "${needs_fetch_bootstrap}" -eq 1 ]]; then
  compose_args+=(-f "${ROOT_DIR}/compose.fetch-bootstrap.yaml")
  preflight_args+=(--fetch-bootstrap)
fi
for profile in "${PROFILES[@]}"; do
  compose_args+=(--profile "${profile}")
done
preflight_args+=("${PROFILES[@]}")

remove_project_orphans() {
  local project_name
  project_name="$(basename "${ROOT_DIR}")"

  declare -A desired_services=()
  while IFS= read -r service_name; do
    [[ -z "${service_name}" ]] && continue
    desired_services["${service_name}"]=1
  done < <(docker compose "${compose_args[@]}" config --services)

  while IFS=$'\t' read -r container_name service_name; do
    [[ -z "${container_name}" || -z "${service_name}" ]] && continue
    [[ -n "${desired_services[${service_name}]:-}" ]] && continue
    docker rm -f "${container_name}" >/dev/null
    echo "Removed orphan container ${container_name} (service=${service_name})"
  done < <(
    docker ps -a \
      --filter "label=com.docker.compose.project=${project_name}" \
      --format '{{.Names}}{{printf "\t"}}{{.Label "com.docker.compose.service"}}'
  )
}

ensure_fetch_bootstrap_network() {
  local project_name
  local network_name

  project_name="$(basename "${ROOT_DIR}")"
  network_name="${project_name}_media"

  if ! command -v docker >/dev/null 2>&1; then
    return 0
  fi

  docker inspect qbittorrent --format '{{json .NetworkSettings.Networks}}' 2>/dev/null |
    grep -q "\"${network_name}\"" && return 0

  docker network connect --alias qbittorrent "${network_name}" qbittorrent >/dev/null 2>&1 || true
}

up_services=()
if [[ "${needs_fetch_bootstrap}" -eq 1 ]]; then
  while IFS= read -r service_name; do
    [[ -z "${service_name}" ]] && continue

    skip_service=0
    for disabled_service in "${BOOTSTRAP_DISABLED_SERVICES[@]}"; do
      if [[ "${service_name}" == "${disabled_service}" ]]; then
        skip_service=1
        break
      fi
    done

    [[ "${skip_service}" -eq 1 ]] && continue
    up_services+=("${service_name}")
  done < <(docker compose "${compose_args[@]}" config --services)

  for disabled_service in "${BOOTSTRAP_DISABLED_SERVICES[@]}"; do
    docker rm -f "${disabled_service}" >/dev/null 2>&1 || true
  done
fi

remove_project_orphans
"${ROOT_DIR}/scripts/preflight.sh" "${preflight_args[@]}"

if [[ "${needs_fetch_bootstrap}" -eq 1 ]]; then
  echo "WireGuard placeholders detected; starting bootstrap fetch path without gluetun/jdownloader"
  docker compose "${compose_args[@]}" up -d --no-deps --remove-orphans "${up_services[@]}"
  ensure_fetch_bootstrap_network
else
  docker compose "${compose_args[@]}" up -d --remove-orphans
fi

# Refresh the Jellyfin container after the host FUSE mounts are confirmed live.
# Existing bind mounts can hold stale rclone endpoints across stack restarts,
# which then causes library scans to fail with "Transport endpoint is not connected".
docker compose "${compose_args[@]}" up -d --no-deps --force-recreate jellyfin
