# CASKAi

**C**ontext · **A**gents · **S**kills · **K**nowledge — **A**rtificial **i**ntelligence, **Governanced**.

Fuente de verdad **gobernada** para los activos de IA de desarrollo (contexts, commands,
agents, skills), **multi-herramienta** (Claude Code + GitHub Copilot, extensible) y consumida
de forma **selectiva, versionada y segura** por 100+ repositorios.

> Si conoces **Renovate/Dependabot**, ya entiendes CASKAi: un catálogo central de *packs*
> versionados que un bot propaga a cada proyecto vía Pull Requests; cualquiera propone mejoras
> de vuelta y la gobernanza las aprueba.

## Qué resuelve
- **Una sola verdad** para la IA de desarrollo (fin de la deriva y la duplicación).
- **Consumo selectivo** por repo (cada proyecto elige los packs que necesita).
- **Seguridad por rol**: visibilidad de packs por **grupos de Entra ID**, aplicada en build.
- **Gobernanza estricta**: tiers (`core`/`domain`/`experimental`), RFC, promoción a core, CODEOWNERS.
- **Trazabilidad 100%** del uso (inventario de adopción).

## Estructura
```
CASKAi/
├─ packs/              # FUENTE DE VERDAD: packs por dominio (core, backend-python…)
│  └─ <pack>/pack.yaml + assets/{context,commands,agents,skills}/
├─ tools/caskai/       # engine en Go (binario único `caskai`)
├─ governance/         # architecture.md (ADRs), degradation.yaml, rfcs/
├─ docs/               # operating-model, security, spike, flujo-e2e…
├─ .claude/            # agentización del propio repo (agents, commands, skills)
├─ install.sh          # instalador Mac/Linux (auto-detecta OS/arch)
├─ install.ps1         # instalador Windows (PowerShell)
└─ CODEOWNERS
```

## Instalación del engine

**Mac / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/jemaji/CASKAi/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/jemaji/CASKAi/main/install.ps1 | iex
```

El script detecta OS y arquitectura automáticamente y descarga el binario correcto de la release más reciente. Los binarios están publicados en [Releases](https://github.com/jemaji/CASKAi/releases) para: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`.

## Quickstart
```bash
caskai version                                                        # verifica la instalación

caskai validate                                                       # gate de CI sobre todos los packs
caskai access    --manifest caskai.yaml                          # visibilidad por rol
caskai build     --manifest caskai.yaml --out .                  # genera .claude/ y .github/
caskai inventory --consumers /ruta/a/consumers                        # trazabilidad de adopción
caskai promote   --asset backend-python/assets/context/x.md --to core # promoción a core
```

## Decisiones clave
- **Engine en Go** (binario único, sin runtime) con **adapters nativos** que emiten a `.claude/` y `.github/` (ver `governance/architecture.md` ADR-10). `rulesync` queda como **propuesta de futuro escalable** para cuando crezca el nº de herramientas objetivo (Fase 5).
- **Distribución vendorizada** (ficheros en cada repo) vía **bot de PRs**; acceso aplicado en build.
- Ver `docs/operating-model.md` (cómo funciona), `docs/security-and-access-control.md` (Entra) y `docs/flujo-e2e.md` (flujo ejecutable).

## Contribuir
Los consumidores **no editan** los ficheros generados; proponen cambios **upstream** vía PR a un
pack. CODEOWNERS enruta por tier/dominio. Ver `CLAUDE.md` para las reglas de trabajo en este repo.
