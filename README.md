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
├─ consumers/          # manifiestos de ejemplo (ai.manifest.yaml)
├─ docs/               # operating-model, security, spike, flujo-e2e…
├─ .claude/            # agentización del propio repo (agents, commands, skills)
└─ CODEOWNERS
```

## Quickstart
```bash
go build -o bin/caskai ./tools/caskai      # compila el engine (binario único, sin deps)

./bin/caskai validate                       # gate de CI sobre todos los packs
./bin/caskai access    --manifest consumers/payments-api/ai.manifest.yaml   # visibilidad por rol
./bin/caskai build     --manifest consumers/payments-api/ai.manifest.yaml --out consumers/payments-api
./bin/caskai inventory                       # trazabilidad de adopción
./bin/caskai promote   --asset backend-python/assets/context/x.md --to core # promoción a core
```

## Decisiones clave
- **Engine en Go** (binario único, sin runtime) + **rulesync** como capa de emisión multi-herramienta (ver `governance/architecture.md` ADR-10).
- **Distribución vendorizada** (ficheros en cada repo) vía **bot de PRs**; acceso aplicado en build.
- Ver `docs/operating-model.md` (cómo funciona), `docs/security-and-access-control.md` (Entra) y `docs/flujo-e2e.md` (flujo ejecutable).

## Contribuir
Los consumidores **no editan** los ficheros generados; proponen cambios **upstream** vía PR a un
pack. CODEOWNERS enruta por tier/dominio. Ver `CLAUDE.md` para las reglas de trabajo en este repo.
