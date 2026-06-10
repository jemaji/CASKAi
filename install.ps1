# install.ps1 — instala caskai en Windows
# Uso: irm https://raw.githubusercontent.com/jemaji/CASKAi/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo    = "jemaji/CASKAi"
$BinName = "caskai.exe"
$InstallDir = "$env:USERPROFILE\bin"

# ── detectar arquitectura ────────────────────────────────────
$Arch = (Get-CimInstance Win32_OperatingSystem).OSArchitecture
switch -Wildcard ($Arch) {
    "*64*ARM*" {
        Write-Host "❌ Windows ARM64 no está soportado por ahora." -ForegroundColor Red
        exit 1
    }
    "*64*" { $GoArch = "amd64" }
    default {
        Write-Host "❌ Arquitectura no soportada: $Arch" -ForegroundColor Red
        exit 1
    }
}

$Binary = "caskai-windows-$GoArch.exe"
$Url    = "https://github.com/$Repo/releases/latest/download/$Binary"

Write-Host "▶ Sistema detectado: windows/$GoArch"
Write-Host "▶ Descargando $Binary desde la release más reciente..."

# ── crear directorio de instalación ─────────────────────────
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

# ── descargar binario ────────────────────────────────────────
$Dest = Join-Path $InstallDir $BinName
Invoke-WebRequest -Uri $Url -OutFile $Dest -UseBasicParsing

# ── verificar instalación ────────────────────────────────────
$Version = & $Dest version
Write-Host "✅ Instalado: $Version" -ForegroundColor Green
Write-Host "   → $Dest"

# ── aviso si el directorio no está en PATH ───────────────────
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host ""
    Write-Host "⚠️  Añade $InstallDir al PATH para usar 'caskai' directamente:" -ForegroundColor Yellow
    Write-Host "   [Environment]::SetEnvironmentVariable('PATH', `"`$env:PATH;$InstallDir`", 'User')"
}
