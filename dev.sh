#!/usr/bin/env bash
# dev.sh — compila caskai desde fuente e instala en ~/bin
# Uso: ./dev.sh [comando caskai y args]
# Ejemplos:
#   ./dev.sh                        → solo compila
#   ./dev.sh version                → compila + caskai version
#   ./dev.sh validate               → compila + caskai validate
#   ./dev.sh build --manifest ...   → compila + caskai build ...

set -euo pipefail

# ── ayuda ────────────────────────────────────────────────────
if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  cat << 'EOF'
dev.sh — compila caskai desde fuente e instala en ~/bin

Uso:
  ./dev.sh [--help]
  ./dev.sh [comando caskai] [args...]

Qué hace:
  1. Compila tools/caskai con 'go build' (requiere Go instalado)
  2. Instala el binario resultante en ~/bin/caskai
  3. Si se pasan argumentos, los ejecuta con caskai directamente

Ejemplos:
  ./dev.sh                        Solo compila y reemplaza ~/bin/caskai
  ./dev.sh version                Compila y ejecuta: caskai version
  ./dev.sh validate               Compila y ejecuta: caskai validate
  ./dev.sh build --manifest ...   Compila y ejecuta: caskai build ...
  ./dev.sh help                   Compila y muestra la ayuda de caskai

Nota: el binario compilado localmente reporta versión "dev" (sin tag de release).
EOF
  exit 0
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="$HOME/bin/caskai"

echo "🔨 Compilando desde fuente..."
go build -o "$OUT" "$SCRIPT_DIR/tools/caskai"
echo "✅ Instalado: $OUT ($(caskai version))"

if [[ $# -gt 0 ]]; then
  echo ""
  caskai "$@"
fi
