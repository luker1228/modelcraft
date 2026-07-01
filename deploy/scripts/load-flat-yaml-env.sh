#!/usr/bin/env sh
set -eu

if [ "$#" -lt 2 ]; then
  echo "usage: $0 <yaml-file> <command> [args...]" >&2
  exit 1
fi

yaml_file="$1"
shift

if [ ! -f "$yaml_file" ]; then
  echo "config file not found: $yaml_file" >&2
  exit 1
fi

trim() {
  value="$1"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

while IFS= read -r raw_line || [ -n "$raw_line" ]; do
  line="$(printf '%s' "$raw_line" | tr -d '\r')"
  case "$line" in
    ''|'#'*)
      continue
      ;;
  esac

  key="$(trim "${line%%:*}")"
  value_part="${line#*:}"
  value="$(trim "$value_part")"

  case "$value" in
    \"*\")
      value="${value#\"}"
      value="${value%\"}"
      ;;
    \'*\')
      value="${value#\'}"
      value="${value%\'}"
      ;;
  esac

  export "$key=$value"
done < "$yaml_file"

exec "$@"
