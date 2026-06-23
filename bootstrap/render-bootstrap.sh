#!/usr/bin/env sh
set -eu

required_vars="
BACKEND_IMAGE
FRONTEND_IMAGE
DATABASE_URL
JWT_SECRET
DEFAULT_DOMAIN
DEFAULT_ADMIN_EMAIL
DEFAULT_ADMIN_PASSWORD
BACKEND_HOST
FRONTEND_HOST
CORS_ALLOWED
LOAD_BALANCER_IP
SERVER_IP
"

for name in $required_vars; do
  eval "value=\${$name:-}"
  if [ -z "$value" ]; then
    echo "Missing required environment variable: $name" >&2
    exit 1
  fi
done

DEFAULT_ADMIN_USERNAME="${DEFAULT_ADMIN_USERNAME:-admin}"
DEFAULT_ADMIN_NAME="${DEFAULT_ADMIN_NAME:-Default Admin}"
TCP_PROXY_HOST="${TCP_PROXY_HOST:-proxy.${DEFAULT_DOMAIN}}"
TCP_PROXY_PORT_START="${TCP_PROXY_PORT_START:-24000}"
TCP_PROXY_PORT_END="${TCP_PROXY_PORT_END:-24999}"

export BACKEND_IMAGE FRONTEND_IMAGE DATABASE_URL JWT_SECRET DEFAULT_DOMAIN
export DEFAULT_ADMIN_EMAIL DEFAULT_ADMIN_PASSWORD DEFAULT_ADMIN_USERNAME DEFAULT_ADMIN_NAME
export BACKEND_HOST FRONTEND_HOST CORS_ALLOWED LOAD_BALANCER_IP SERVER_IP
export TCP_PROXY_HOST TCP_PROXY_PORT_START TCP_PROXY_PORT_END

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

python3 - "$script_dir/kubesa-bootstrap.yaml.tpl" <<'PY'
import os
import re
import sys

path = sys.argv[1]
with open(path, "r", encoding="utf-8") as f:
    rendered = f.read()

placeholders = sorted(set(re.findall(r"__([A-Z0-9_]+)__", rendered)))
missing = [key for key in placeholders if key not in os.environ]
if missing:
    raise SystemExit(f"Unresolved template placeholders: {', '.join(missing)}")

for key in placeholders:
    value = os.environ[key]
    value = value.replace("\\", "\\\\").replace('"', '\\"').replace("\n", "\\n")
    rendered = rendered.replace(f"__{key}__", value)

print(rendered)
PY
