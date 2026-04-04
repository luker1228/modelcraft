#!/usr/bin/env bash
set -euo pipefail

usage() {
	cat <<'EOF'
Usage:
  ./scripts/generate-safe-querier.sh [options]

Options:
  --package-path <path>      Package path passed to gowrap -p
  --interface-name <name>    Interface name passed to gowrap -i
  --output-file <path>       Output file path (relative to repo root or absolute)
  --wrapper-name <name>      Generated wrapper struct name
  --constructor-name <name>  Generated constructor function name
  --delegate-type <type>     Delegate field type (e.g. dbgen.Querier)
  -h, --help                 Show this help message

Defaults wrap the full dbgen.Querier interface.
EOF
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PACKAGE_PATH="./internal/infrastructure/dbgen"
INTERFACE_NAME="Querier"
OUTPUT_FILE="internal/infrastructure/dbgenwrap/safe_querier_gen.go"
WRAPPER_NAME="SafeQuerier"
CONSTRUCTOR_NAME="NewSafeQuerier"
DELEGATE_TYPE="dbgen.Querier"

while [[ $# -gt 0 ]]; do
	case "$1" in
	--package-path)
		PACKAGE_PATH="$2"
		shift 2
		;;
	--interface-name)
		INTERFACE_NAME="$2"
		shift 2
		;;
	--output-file)
		OUTPUT_FILE="$2"
		shift 2
		;;
	--wrapper-name)
		WRAPPER_NAME="$2"
		shift 2
		;;
	--constructor-name)
		CONSTRUCTOR_NAME="$2"
		shift 2
		;;
	--delegate-type)
		DELEGATE_TYPE="$2"
		shift 2
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "Unknown option: $1" >&2
		usage
		exit 1
		;;
	esac
done

if [[ "${OUTPUT_FILE}" != /* ]]; then
	OUTPUT_FILE="${ROOT_DIR}/${OUTPUT_FILE}"
fi

mkdir -p "$(dirname "${OUTPUT_FILE}")"

TEMPLATE_FILE="$(mktemp)"
trap 'rm -f "${TEMPLATE_FILE}"' EXIT

cat >"${TEMPLATE_FILE}" <<EOF
type ${WRAPPER_NAME} struct {
	delegate ${DELEGATE_TYPE}
}

var _ ${DELEGATE_TYPE} = (*${WRAPPER_NAME})(nil)

func ${CONSTRUCTOR_NAME}(delegate ${DELEGATE_TYPE}) ${DELEGATE_TYPE} {
	return &${WRAPPER_NAME}{delegate: delegate}
}

{{- range \$name, \$method := .Interface.Methods }}
func (s *${WRAPPER_NAME}) {{\$method.Declaration}} {
	{{- if \$method.ReturnsError }}
	{{\$method.Results.Pass}} = s.delegate.{{\$method.Name}}({{\$method.Params.Pass}})
	WrapSQLErrorInPlace(&err)
	return
	{{- else if \$method.HasResults }}
	return s.delegate.{{\$method.Name}}({{\$method.Params.Pass}})
	{{- else }}
	s.delegate.{{\$method.Name}}({{\$method.Params.Pass}})
	{{- end }}
}

{{- end }}
EOF

go run -C "${ROOT_DIR}" github.com/hexdigest/gowrap/cmd/gowrap@latest gen \
	-p "${PACKAGE_PATH}" \
	-i "${INTERFACE_NAME}" \
	-t "${TEMPLATE_FILE}" \
	-o "${OUTPUT_FILE}" \
	-g

echo "Generated: ${OUTPUT_FILE}"
