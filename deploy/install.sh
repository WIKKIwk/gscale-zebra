#!/usr/bin/env bash
set -euo pipefail

PREFIX="/opt/gscale-zebra"
APP_USER="${SUDO_USER:-}"
APP_GROUP=""
START_AFTER_INSTALL=0

usage() {
  cat <<'EOF'
Usage: ./install.sh [options]

Install gscale-zebra binaries and systemd services.

Options:
  --prefix <path>  Install root (default: /opt/gscale-zebra)
  --user <name>    Service user (default: SUDO_USER/current user)
  --group <name>   Service group (default: primary group of user)
  --start          Start services after install
  -h, --help       Show help
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --prefix)
      PREFIX="${2:-}"
      shift 2
      ;;
    --user)
      APP_USER="${2:-}"
      shift 2
      ;;
    --group)
      APP_GROUP="${2:-}"
      shift 2
      ;;
    --start)
      START_AFTER_INSTALL=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown arg: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run install.sh as root (example: sudo ./install.sh --start)." >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

required=(
  "${SCRIPT_DIR}/bin/bot"
  "${SCRIPT_DIR}/bin/scale"
  "${SCRIPT_DIR}/config/bot.env.example"
  "${SCRIPT_DIR}/config/scale.env.example"
  "${SCRIPT_DIR}/systemd/gscale-bot.service"
  "${SCRIPT_DIR}/systemd/gscale-scale.service"
)

for f in "${required[@]}"; do
  if [[ ! -f "$f" ]]; then
    echo "Required file not found: $f" >&2
    exit 1
  fi
done

if [[ -z "${APP_USER}" ]]; then
  APP_USER="$(id -un)"
fi

if ! id -u "${APP_USER}" >/dev/null 2>&1; then
  echo "User does not exist: ${APP_USER}" >&2
  exit 1
fi

if [[ -z "${APP_GROUP}" ]]; then
  APP_GROUP="$(id -gn "${APP_USER}")"
fi

if ! getent group "${APP_GROUP}" >/dev/null 2>&1; then
  echo "Group does not exist: ${APP_GROUP}" >&2
  exit 1
fi

echo "==> Installing to ${PREFIX}"
install -d -m 0755 "${PREFIX}" "${PREFIX}/bin" "${PREFIX}/config" "${PREFIX}/logs"

install -m 0755 "${SCRIPT_DIR}/bin/bot" "${PREFIX}/bin/bot"
install -m 0755 "${SCRIPT_DIR}/bin/scale" "${PREFIX}/bin/scale"
if [[ -f "${SCRIPT_DIR}/bin/zebra" ]]; then
  install -m 0755 "${SCRIPT_DIR}/bin/zebra" "${PREFIX}/bin/zebra"
fi

if [[ ! -f "${PREFIX}/config/bot.env" ]]; then
  install -m 0640 "${SCRIPT_DIR}/config/bot.env.example" "${PREFIX}/config/bot.env"
fi
if [[ ! -f "${PREFIX}/config/scale.env" ]]; then
  install -m 0640 "${SCRIPT_DIR}/config/scale.env.example" "${PREFIX}/config/scale.env"
fi

chown -R "${APP_USER}:${APP_GROUP}" "${PREFIX}/logs"

escape_sed() {
  printf '%s' "$1" | sed 's/[\\/&]/\\&/g'
}

prefix_esc="$(escape_sed "${PREFIX}")"
user_esc="$(escape_sed "${APP_USER}")"
group_esc="$(escape_sed "${APP_GROUP}")"

render_unit() {
  local in_file="$1"
  local out_file="$2"
  sed \
    -e "s/__PREFIX__/${prefix_esc}/g" \
    -e "s/__APP_USER__/${user_esc}/g" \
    -e "s/__APP_GROUP__/${group_esc}/g" \
    "${in_file}" > "${out_file}"
}

tmp_scale="$(mktemp)"
tmp_bot="$(mktemp)"
trap 'rm -f "${tmp_scale}" "${tmp_bot}"' EXIT

render_unit "${SCRIPT_DIR}/systemd/gscale-scale.service" "${tmp_scale}"
render_unit "${SCRIPT_DIR}/systemd/gscale-bot.service" "${tmp_bot}"

install -m 0644 "${tmp_scale}" /etc/systemd/system/gscale-scale.service
install -m 0644 "${tmp_bot}" /etc/systemd/system/gscale-bot.service

echo "==> Reloading systemd"
systemctl daemon-reload
systemctl enable gscale-scale.service gscale-bot.service >/dev/null

if [[ "${START_AFTER_INSTALL}" == "1" ]]; then
  echo "==> Starting services"
  systemctl restart gscale-scale.service gscale-bot.service
fi

echo
echo "Installed."
echo "Config files:"
echo " - ${PREFIX}/config/scale.env"
echo " - ${PREFIX}/config/bot.env"
echo
echo "Useful commands:"
echo " - systemctl status gscale-scale.service"
echo " - systemctl status gscale-bot.service"
echo " - journalctl -u gscale-scale.service -f"
echo " - journalctl -u gscale-bot.service -f"
