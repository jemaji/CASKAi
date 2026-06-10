---
id: commit-helper
type: skill
name: commit-helper
description: "Genera mensajes de commit siguiendo Conventional Commits. Úsalo al preparar un commit."
resources: [scripts/]
targets:
  claude:
    emit: skill
  copilot:
    emit: prompt   # override explícito; existe además un default en governance/degradation.yaml
---
# Commit Helper

Cuando el usuario prepare un commit:

1. Analiza el diff en *staging* (`git diff --cached`).
2. Determina el tipo: `feat` | `fix` | `chore` | `docs` | `refactor` | `test`.
3. Genera un mensaje Conventional Commit de una línea + cuerpo opcional.

Consulta `scripts/format.py` para el formato exacto.
