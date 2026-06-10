---
name: pack-authoring
description: Esquema canónico de CASKAi para autorar assets (context/command/agent/skill) y packs. Úsalo al crear o editar cualquier cosa bajo packs/, para acertar con el frontmatter, la clasificación de acceso y la degradación fail-closed.
---

# Autoría de packs CASKAi

## pack.yaml
```yaml
name: <pack>
version: 0.1.0            # semver — sube minor al añadir; major si rompes
tier: domain              # core | domain | experimental
description: "..."
owners: ["@org/<equipo>"]
targets: [claude, copilot]
channels: [stable]
access:
  classification: internal       # internal | restricted | confidential
  allowed_groups: [<grupo-entra>] # obligatorio si NO es internal
```

## Frontmatter por tipo de asset (en packs/<pack>/assets/<tipo>/)

**context** (`context/<id>.md`)
```yaml
---
id: <id>
type: context
description: "..."
scope: { applyTo: "**/*.py" }   # opcional; restringe por ruta
targets: [claude, copilot]
---
```

**command** (`commands/<id>.md`)
```yaml
---
id: <id>
type: command
title: "..."
description: "..."
argument_hint: "[...]"          # opcional
targets: [claude, copilot]
---
Cuerpo con {{ARGS}} (el engine lo traduce: $ARGUMENTS / ${input:args}).
```

**agent** (`agents/<id>.md`)
```yaml
---
id: <id>
type: agent
description: "..."
targets: [claude, copilot]
---
```

**skill** (`skills/<id>/SKILL.md`)
```yaml
---
id: <id>
type: skill
name: <id>
description: "..."
resources: [scripts/]           # opcional; se copian solo a Claude
targets:
  claude: { emit: skill }
  copilot: { emit: prompt }     # degradación explícita (Copilot no tiene skills)
---
```

## Reglas de oro
- `id` **estable y global**; renombrarlo es un **breaking change**.
- Variables del cuerpo: `{{ARGS}}`, `{{TARGET}}`. No escribas sintaxis específica de una herramienta.
- **Degradación fail-closed**: si un tipo no existe en una herramienta destino (skill→Copilot),
  declara `emit` en el frontmatter o una regla en `governance/degradation.yaml`. Si no, el build rompe.
- **Seguridad**: nada de secretos/PII. Clasifica el pack según su sensibilidad y pon `allowed_groups`.
- **Gobernanza**: `tier: core`, breaking, seguridad o alto impacto → **RFC + board**. Resto → PR con owner.

## Verificación
Siempre termina con `./bin/caskai validate` y revisa el routing con
`python3 tools/codeowners-route.py <ficheros-cambiados>`.
