#!/usr/bin/env bash
# dev.sh — compila caskai desde fuente e instala en ~/bin
# Uso: ./dev.sh [comando caskai y args]
# Ejemplos:
#   ./dev.sh                        → solo compila
#   ./dev.sh version                → compila + caskai version
#   ./dev.sh validate               → compila + caskai validate
#   ./dev.sh build --manifest ...   → compila + caskai build ...

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$HOME/bin/caskai"

echo "🔨 Compilando desde fuente..."
go build -o "$OUT" "$SCRIPT_DIR/tools/caskai"
echo "✅ Instalado: $OUT ($(caskai version))"

if [[ $# -gt 0 ]]; then
  echo ""
  caskai "$@"
fi
