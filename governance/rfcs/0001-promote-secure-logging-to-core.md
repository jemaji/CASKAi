# RFC-0001 — Promover `secure-logging` de `backend-python` a `core`

- Estado: **Aceptado**
- Autor: @org/backend-guild
- Aprobado por: **Board de gobernanza** (@org/ai-governance)
- Fecha: 2026-06-08

## Contexto
El context `secure-logging` nació en el pack de dominio `backend-python`, pero la
regla (no registrar secretos/PII, redactar cabeceras de auth) aplica a **toda la
organización**, no solo a backend.

## Propuesta
Promover el asset a `core` (clasificación `internal`) para que lo herede toda la
ingeniería. Implica versión nueva de `core` (0.1.0 → 0.2.0).

## Decisión
Aceptada. Al tocar `packs/core/`, CODEOWNERS exige aprobación del **board**
(@org/ai-governance). Tras el merge, el bot propagará el cambio a todos los
consumidores de `core` en el próximo release train.
