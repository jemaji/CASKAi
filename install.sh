#!/usr/bin/env bash
# install.sh — instala el binario caskai correcto para el sistema actual
# Uso: curl -fsSL https://raw.githubusercontent.com/jemaji/CASKAi/main/install.sh | bash

set -euo pipefail

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

# ── aviso si el directorio no está en PATH ───────────────────
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo ""
  echo "⚠️  Añade esto a tu ~/.zshrc o ~/.bashrc para usar 'caskai' directamente:"
  echo "   export PATH=\"\$HOME/bin:\$PATH\""
fi
