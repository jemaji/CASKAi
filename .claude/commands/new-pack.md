---
description: Crea un nuevo pack de dominio siguiendo el formato canónico y la gobernanza
argument-hint: "<nombre-del-pack> [tier] [@org/owner]"
---
Crea un nuevo pack llamado **$ARGUMENTS** delegando en el subagente `pack-author`.

Antes de escribir nada, asegúrate de tener (pregunta solo si falta):
- **tier** (`domain` por defecto; `core`/`experimental` si procede),
- **owner** (equipo `@org/...`),
- **clasificación de acceso** (`internal` / `restricted` + `allowed_groups`).

Luego: crea `packs/$ARGUMENTS/pack.yaml` y al menos un asset de ejemplo, ejecuta
`caskai validate`, y muestra qué CODEOWNERS revisarían el PR
(`python3 tools/codeowners-route.py packs/$ARGUMENTS/...`). Recuerda al usuario que un pack `core`
o un cambio de seguridad requeriría RFC.
