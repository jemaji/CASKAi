# Spike comparativo — rulesync vs Rosetta (sobre nuestro modelo)

> Objetivo: decidir con evidencia si reutilizamos rulesync o Rosetta, y qué implica para
> lenguaje, consumo selectivo y securización. **rulesync se ejecutó de verdad**; Rosetta
> se caracterizó desde su documentación oficial (no ofrece `pip install`; su despliegue es
> un servicio MCP hospedado/self-host + plugin de marketplace, no demostrable headless).

## rulesync — EJECUTADO ✅

```bash
rulesync init                  # crea .rulesync/{rules,commands,subagents,skills,mcp.json,hooks.json}
rulesync generate --targets claudecode,copilot --features rules,commands,subagents,skills
```

Resultado real (8 ficheros, ambos targets):

| Tipo canónico (rulesync) | → Claude | → Copilot |
|---|---|---|
| rules | `CLAUDE.md` | `.github/copilot-instructions.md` |
| commands | `.claude/commands/review-pr.md` | `.github/prompts/review-pr.prompt.md` |
| subagents | `.claude/agents/planner.md` | `.github/agents/planner.agent.md` |
| skills | `.claude/skills/.../SKILL.md` | `.github/skills/.../SKILL.md` |

**Conclusión:** rulesync ES, ya hecho, nuestra **capa de emisión multi-herramienta (C1)**.
Su `.rulesync/` y su frontmatter (`targets`, `description`, `globs`) son casi idénticos a
nuestro modelo. TypeScript, MIT, **ofrece binario único** (curl) además de npm/brew.

**Pero NO cubre (seguiría siendo nuestro):**
- **Consumo selectivo por repo (C2):** rulesync es **autoría de UN repo** (qué ficheros
  hay en `.rulesync/`), no un **catálogo de packs versionado** del que un repo "elige A y C".
- **Acceso por grupo de Entra (C5)** y **trazabilidad (C6):** no existen.
- **Degradación fail-closed:** rulesync es **optimista** — emitió `.github/skills/` aunque
  Copilot no consume skills nativas (nosotros rompemos y exigimos mapeo).
- **Fidelidad de variables:** dejó `$ARGUMENTS` tal cual en el prompt de Copilot (no lo
  tradujo a `${input:...}`). Adaptación más "copia ligera" que la nuestra.

## Rosetta — NO ejecutado (por su modelo)

Su documentación **no da `pip install`**. Despliegue de referencia:
- **MCP hospedado:** `https://mcp.rosetta.griddynamics.net/mcp` (servicio externo de Grid Dynamics).
- **Plugin de marketplace:** `claude plugin marketplace add griddynamics/rosetta`; para Copilot, marketplace de plugins en VS Code.

Modelo: **servir instrucciones por MCP en runtime**, **merge en 3 capas (core/org/proyecto)**,
**auto-clasificación en 12 workflows** + metodología Prepare/Research/Plan/Act. *"Rosetta nunca
ve tu código; el agente pide por tag."* **Sin control de acceso por rol. Sin telemetría de uso.**

**Implicaciones frente a tus dudas:**
- **Lenguaje:** adoptar Rosetta = **adoptar su framework** (Python/TS) y vivir **río abajo de
  Grid Dynamics**; nuestro engine Go se cae.
- **Consumo selectivo:** no es "elige packs en un manifiesto"; es **merge de 3 capas + routing
  por workflow en runtime**. Menos granular y con un **servicio MCP vivo** en el camino.
- **Securización:** sin RBAC; en su modelo, la seguridad sería un **gateway de auth delante del
  MCP** (chokepoint vivo, siempre disponible y blindado). Y por defecto el MCP es **externo**
  (perímetro). Frente a nuestro modelo de **ficheros**, donde el acceso se aplica **en
  build/distribución** (estático, auditable, sin servicio que atacar).
- **Copilot:** soportado vía plugin de VS Code (no vía las superficies de instrucciones por
  fichero de Copilot). A verificar si encaja con vuestro uso real de Copilot.

## Comparativa frente a tus prioridades

| Eje | rulesync | Rosetta | Nuestro modelo |
|---|---|---|---|
| Lenguaje | TypeScript (equipo) | Framework Python/TS ajeno | Go (o TS si reusamos rulesync) |
| Entrega | **Ficheros** (vendorizado) | **MCP en runtime** (servicio vivo) | **Ficheros** (vendorizado) |
| Multi-tool Claude+Copilot | ✅ demostrado | ◐ vía plugins (Copilot a verificar) | ✅ |
| Consumo selectivo por repo | ◐ (por fichero, 1 repo) | ◐ (3 capas runtime) | ✅ (manifiesto + packs versionados) |
| Acceso por grupo Entra (C5) | ✗ | ✗ | ✅ (en build) |
| Trazabilidad de uso (C6) | ✗ | ✗ | ✅ (caskai.lock + inventory) |
| Perímetro / on-prem | ✅ (local) | ◐ (hospedado por defecto) | ✅ |
| Gobernanza tiers / promoción a core | ✗ | ◐ (3 capas, sin RBAC) | ✅ |

## Recomendación (con evidencia)

1. **Reutilizar rulesync como capa de emisión (C1)** y **conservar nuestra capa** para lo que
   nadie da: catálogo de packs **versionado** + **consumo selectivo por manifiesto** (C2),
   **acceso por grupo de Entra en build** (C5), **bot de propagación** (C3), **inventario**
   (C6) y **gobernanza por tiers/promoción a core** (C4).
   - Lo que debemos seguir poniendo nosotros sobre rulesync: **degradación fail-closed**,
     **traducción de variables** fiel, acceso, registro/versionado, propagación y telemetría.
   - Lenguaje: esto inclina a **TypeScript** (equipo + rulesync). Alternativa: mantener Go e
     **invocar el binario de rulesync** como subproceso (sin Node en runtime).
2. **No adoptar Rosetta como base:** su entrega MCP-runtime + hospedado + metodología choca con
   perímetro/Entra/ficheros y con tus dos prioridades (consumo selectivo granular y securización
   simple). **Sí** tomarlo como **referencia** (merge 3-capas, metodología, review-subagent).

## Decisión registrada (ADR-10)
**Nuestra capa + rulesync, manteniendo Go.** El engine `caskai` sigue en Go (binario único) con
la orquestación propia (validación, acceso por grupo de Entra en build, lockfile, inventario,
promoción, gobernanza) e **invoca el binario de `rulesync` como subproceso** para la emisión
multi-herramienta. Rosetta queda como **referencia**, no como base.

Implicaciones / pendientes:
- Refactor de `caskai build`: en vez de adapters Go propios, traducir nuestros assets de pack a
  entradas `.rulesync/` (o usar `rulesync convert/generate`) e invocar el binario.
- Reañadir por encima de rulesync lo que no trae: **fail-closed con mapeo**, **traducción de
  variables** por herramienta, acceso, versionado/registro, propagación e inventario (ya nuestros).
- Empaquetar el binario de `rulesync` junto al de `caskai` en la GitHub Action / bot (sin Node).
