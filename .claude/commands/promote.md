---
description: Promociona un asset de un pack de dominio a core (con RFC y bump de versión)
argument-hint: "<pack>/assets/<ruta-del-asset>"
---
Promociona el asset **$ARGUMENTS** a `core`. Es un cambio de tier que **requiere RFC y aprobación
del board** (`@org/ai-governance`).

Pasos:
1. `./bin/caskai promote --asset $ARGUMENTS --to core`
2. Sube la versión de `packs/core/pack.yaml` (semver minor) y crea un RFC en
   `governance/rfcs/NNNN-<slug>.md` (contexto, propuesta, decisión, aprobador = board).
3. `./bin/caskai validate`.
4. Muestra el routing del PR (`python3 tools/codeowners-route.py $(git diff --name-only)`),
   que debe exigir `@org/ai-governance` por tocar `packs/core/`.

Resume el impacto: qué consumidores de `core` heredarán el asset tras el próximo release train.
