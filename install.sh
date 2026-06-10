#!/usr/bin/env bash
# install.sh — instala el binario caskai correcto para el sistema actual
# Uso: curl -fsSL https://raw.githubusercontent.com/jemaji/CASKAi/main/install.sh | bash

set -euo pipefail

# ── ayuda ────────────────────────────────────────────────────
if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  cat << 'EOF'
install.sh — instalador de caskai (CASKAi engine)

Uso:
  curl -fsSL https://raw.githubusercontent.com/jemaji/CASKAi/main/install.sh | bash
  bash install.sh [--help]

Qué hace:
  1. Detecta tu OS y arquitectura automáticamente
  2. Descarga el binario correcto de la release más reciente de GitHub
  3. Lo instala en ~/bin/caskai (configurable con CASKAI_INSTALL_DIR)
  4. Añade ~/bin al PATH en tu perfil de shell si no estaba ya

Variables de entorno:
  CASKAI_INSTALL_DIR   Directorio de instalación (por defecto: ~/bin)

Plataformas soportadas:
  linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64

Tras instalar:
  caskai help          Ver todos los comandos disponibles
  caskai version       Verificar la versión instalada
EOF
  exit 0
fi

REPO="jemaji/CASKAi"
BIN_NAME="caskai"
INSTALL_DIR="${CASKAI_INSTALL_DIR:-$HOME/bin}"

# ── detectar OS ──────────────────────────────────────────────
OS="$(uname -s)"
case "$OS" in
  Linux)   GOOS="linux"   ;;
  Darwin)  GOOS="darwin"  ;;
  MINGW*|MSYS*|CYGWIN*) GOOS="windows" ;;
  *)
    echo "❌ Sistema operativo no soportado: $OS"
    exit 1
    ;;
esac

# ── detectar arquitectura ────────────────────────────────────
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)  GOARCH="amd64" ;;
  arm64|aarch64) GOARCH="arm64" ;;
  *)
    echo "❌ Arquitectura no soportada: $ARCH"
    exit 1
    ;;
esac

# ── windows solo tiene amd64 por ahora ──────────────────────
if [[ "$GOOS" == "windows" && "$GOARCH" != "amd64" ]]; then
  echo "❌ Windows solo soporta amd64 por ahora"
  exit 1
fi

# ── nombre del binario ───────────────────────────────────────
BINARY="${BIN_NAME}-${GOOS}-${GOARCH}"
[[ "$GOOS" == "windows" ]] && BINARY="${BINARY}.exe"

# ── URL de descarga ──────────────────────────────────────────
URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"

echo "▶ Sistema detectado: ${GOOS}/${GOARCH}"
echo "▶ Descargando ${BINARY} desde la release más reciente..."

mkdir -p "$INSTALL_DIR"
curl -fsSL "$URL" -o "${INSTALL_DIR}/${BIN_NAME}"
chmod +x "${INSTALL_DIR}/${BIN_NAME}"

# ── verificar instalación ────────────────────────────────────
VERSION="$("${INSTALL_DIR}/${BIN_NAME}" version 2>&1)"
echo "✅ Instalado: ${VERSION}"
echo "   → ${INSTALL_DIR}/${BIN_NAME}"

# ── añadir al PATH si no está ya ─────────────────────────────
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  EXPORT_LINE="export PATH=\"\$HOME/bin:\$PATH\""

  # detectar el shell y el fichero de perfil correcto
  PROFILE=""
  if [[ "$SHELL" == */zsh ]]; then
    PROFILE="$HOME/.zshrc"
  elif [[ "$SHELL" == */bash ]]; then
    PROFILE="${BASH_PROFILE:-$HOME/.bashrc}"
    [[ "$GOOS" == "darwin" ]] && PROFILE="$HOME/.bash_profile"
  fi

  if [[ -n "$PROFILE" ]]; then
    # solo añadir si la línea no existe ya en el fichero
    if ! grep -qF "$INSTALL_DIR" "$PROFILE" 2>/dev/null; then
      echo "" >> "$PROFILE"
      echo "# caskai (CASKAi engine)" >> "$PROFILE"
      echo "$EXPORT_LINE" >> "$PROFILE"
    fi
    # aplicar en la sesión actual y hacer source del perfil
    export PATH="$INSTALL_DIR:$PATH"
    # shellcheck disable=SC1090
    source "$PROFILE" 2>/dev/null || true
    echo "✅ PATH actualizado en $PROFILE y aplicado en la sesión actual"
  else
    # sin perfil detectado: al menos aplicar en la sesión actual
    export PATH="$INSTALL_DIR:$PATH"
    echo "⚠️  No se pudo detectar el perfil del shell. Añade manualmente a tu perfil:"
    echo "   $EXPORT_LINE"
  fi
fi

# ── recordatorio CASKAI_ROOT ─────────────────────────────────
echo ""
echo "──────────────────────────────────────────────"
echo "  ✅ caskai instalado correctamente."
echo ""
echo "  Paso siguiente: define CASKAI_ROOT en tu perfil"
echo "  para que caskai sepa dónde están los packs:"
echo ""
echo "    export CASKAI_ROOT=~/CODE/CASKAi   # ajusta la ruta"
echo ""
echo "  Luego, desde cualquier consumidor basta con:"
echo "    cd ~/mi-proyecto && caskai build"
echo ""
echo "  Si 'caskai' no responde aún en este terminal:"
echo "    source ${PROFILE:-~/.zshrc}"
echo "──────────────────────────────────────────────"
