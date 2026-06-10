# Arquitectura: Plataforma centralizada de IA (`CASKAi`)

> Documento de arquitectura / ADR maestro
> Estado: **Propuesto** — pendiente de validación por el equipo de gobernanza
> Fecha: 2026-06-05
> Autores: Equipo de gobernanza de IA
> Versión del documento: 1.0

---

## 1. Contexto y problema

Los activos de IA para asistentes de código —**agents**, **skills**, **commands** y **contexts/reglas**— hoy viven dispersos y duplicados en cada proyecto (`.claude/`, `.github/`, etc.). Esto provoca:

- **Deriva**: cada repo evoluciona sus propias versiones; no hay una única verdad.
- **Falta de gobierno**: nadie revisa la calidad ni la seguridad de forma transversal.
- **No escala**: con 100+ repos, propagar una mejora o un fix es manual e inconsistente.
- **Fragmentación entre herramientas**: Claude Code y GitHub Copilot usan formatos incompatibles.

Necesitamos una **plataforma interna** que centralice estos activos en una fuente de verdad, permita a cada proyecto **consumir selectivamente** lo que necesita, permita **proponer cambios de vuelta**, y mantenga todo **gobernado, versionado y actualizado** para todos los consumidores.

## 2. Objetivos y no-objetivos

### Objetivos
- Fuente de verdad única para todos los activos de IA.
- Consumo selectivo por proyecto (cada repo toma solo lo que le interesa).
- Soporte multi-herramienta: **Claude Code** y **GitHub Copilot** desde el día 1, **extensible** a otras (Cursor, etc.).
- Bucle de contribución gobernado: los proyectos proponen, gobernanza evalúa, se propaga.
- Gobernanza **estricta** y escalable a 100+ repos.
- Versionado, trazabilidad y telemetría de adopción.

### No-objetivos (de momento)
- Evals automatizados de calidad de assets (se difieren a **fase 2**).
- Soporte a herramientas más allá de Claude y Copilot (la arquitectura lo permite, pero no se implementa aún).
- Telemetría en runtime (se usa escaneo de lockfiles, no instrumentación de ejecución).

## 3. Restricciones que condicionan el diseño

| Herramienta | Contexto/reglas | Comando | Agente | Skill | Empaquetado | Fuente remota |
|---|---|---|---|---|---|---|
| **Claude Code** | `CLAUDE.md`, memoria | `commands/*.md` | `agents/*.md` | `skills/<n>/SKILL.md` | Plugin + marketplace | **Sí** (marketplace) |
| **GitHub Copilot** | `.github/copilot-instructions.md`, `.github/instructions/*` (`applyTo`) | `.github/prompts/*.prompt.md` | `.github/chatmodes/*.chatmode.md` | *(sin equivalente nativo)* | *(solo archivos locales)* | **No** |

**Consecuencias clave:**
1. Los formatos no coinciden → hace falta una capa de abstracción propia.
2. Las **skills no tienen destino nativo en Copilot** → requieren una política de degradación.
3. Copilot **solo lee archivos locales** → para Copilot es obligatorio "vendorizar" (commitear los artefactos generados en cada repo consumidor).

---

## 4. Decisiones de arquitectura (ADRs)

### ADR-1 — Fuente de verdad canónica neutral + adapters
**Decisión:** Los autores escriben en un **formato canónico neutral propio** (markdown + frontmatter con esquema propio). Unos **adapters** (generadores) compilan ese formato a los formatos nativos de cada herramienta.

**Alternativas consideradas:**
- *Claude como formato canónico y traducir a Copilot*: más barato al inicio, pero acopla la plataforma a la evolución de un vendor y ensucia la autoría con detalles específicos.
- *Mantener formatos nativos duplicados a mano*: garantiza deriva y duplicación a escala.

**Justificación:** A 100+ repos y con la meta de ser extensible a más herramientas, el formato neutral envejece mejor y desacopla autoría de distribución. Coste asumido: mantener los adapters.

### ADR-2 — Monorepo de packs
**Decisión:** Todos los activos viven en un **monorepo**, organizados en **packs** (bundles por dominio: `core`, `backend-python`, `frontend-react`, …), no por tipo de artefacto.

**Justificación:** Agrupar por dominio es lo que permite el consumo selectivo. El monorepo unifica CI, versionado y gobernanza (ownership por carpeta vía CODEOWNERS), evitando el overhead de coordinación de multi-repo.

### ADR-3 — Cuatro tipos de artefacto neutrales
**Decisión:** El esquema canónico define 4 tipos: `context`, `command`, `agent`, `skill`. Cada uno mapea a su equivalente por herramienta (ver §5).

### ADR-4 — Versionado por pack (semver)
**Decisión:** Cada pack tiene su propio semver. Los proyectos consumen `pack@versión`. El `id` de cada asset es **global y estable** (el versionado y las deprecaciones se anclan al `id`, nunca al nombre de archivo).

**Alternativas:** versionado por asset (explosión de combinaciones); pack + lockfile estricto (se adopta el lockfile, ver ADR-8, pero el semver vive a nivel de pack).

### ADR-5 — Degradación *fail-closed con mapeo registrado*
**Decisión:** Cuando un asset declara como destino una herramienta que **no soporta** su tipo (p. ej. una `skill` hacia Copilot), la CI **rompe** con un error accionable. El autor (o gobernanza) declara un **mapeo explícito** que queda **persistido y auditado**, y a partir de entonces se reutiliza automáticamente.

Dos capas de mapeo:
- **Por tipo** (`governance/degradation.yaml`): default revisado por gobernanza para todo el tipo (`skill -> copilot`).
- **Override por-asset**: en el frontmatter, cuando el default no encaja.

**Justificación:** Combina disciplina (nada se degrada en silencio), memoria (la decisión queda registrada) y escalabilidad (el segundo asset del mismo tipo hereda el default).

### ADR-6 — Gobernanza por tiers
**Decisión:** Tres tiers de ownership con distinto nivel de aprobación (ver §6).

### ADR-7 — Release trains quincenales
**Decisión:** Los cambios se publican en ventanas fijas cada 2 semanas, con una **vía rápida (hotfix)** para seguridad y bugs críticos (doble aprobación).

### ADR-8 — Adopción vía lockfile + escaneo
**Decisión:** Cada repo consumidor mantiene un `caskai.lock` con `pack@versión` exactos. Un job escanea los repos de la organización y construye el inventario de adopción. Sin telemetría en runtime.

### ADR-9 — Distribución: bot de PRs + vendorizado de ambas herramientas
**Decisión:** Un **bot estilo Renovate** abre PRs de actualización a cada repo consumidor cuando sale un train. Tanto los artefactos de **Claude** como los de **Copilot** se **vendorizan** (se commitean en el repo) por consistencia, aunque Claude soporte marketplace remoto.

**Justificación:** Un único modelo mental, diffs revisables por PR, reproducibilidad y funcionamiento offline. Para Copilot el vendorizado es obligatorio de todos modos.

### ADR-10 — Engine en Go con adapters nativos (emisión propia)
**Decisión:** El engine (`caskai`) se mantiene en **Go** (binario único) y **conserva toda la
emisión a ficheros nativos** (`.claude/` y `.github/`) mediante **adapters propios** (`emitClaude`,
`emitCopilot`), además de la orquestación: validación, **acceso por grupo de Entra en build**,
lockfile, inventario, promoción y gobernanza. **Cero dependencias en runtime**: ni Node ni binarios
de terceros.

**Justificación:** Con **dos herramientas objetivo** (Claude Code + Copilot) el coste de mantener
adapters nativos es bajo y ya está implementado. A cambio se obtiene control total sobre la
**fidelidad de la traducción de variables**, el paso de **degradación fail-closed con mapeo**
(ADR-5) y la integración con el **control de acceso**, sin acoplar el core a la evolución de un
binario externo.

**Alternativas consideradas:**
- **Delegar la emisión en `rulesync`** (subproceso) — descartado *de momento*: con solo 2 targets el
  ahorro es marginal y, aun delegando, habría que construir igualmente encima degradación, acceso,
  lockfile, bot y telemetría (rulesync no lo cubre). Acopla el core a un tercero sin beneficio neto
  a esta escala. **Se reserva como propuesta de futuro escalable** (ver §11, Fase 5): cuando el
  número de herramientas objetivo crezca a ~4+, reevaluar rulesync como motor de emisión para no
  mantener un adapter nativo por herramienta.
- **Reescribir todo en TypeScript** (rulesync nativo) — descartado para conservar el binario único
  Go y la propiedad del core.
- **Adoptar Rosetta como base** — descartado por su entrega MCP-runtime/hospedada, ausencia de
  RBAC/telemetría y choque con perímetro+Entra (solo se usa como **referencia**: merge 3-capas,
  metodología).

> **Nota de evolución:** una versión previa de este ADR proponía delegar la emisión en `rulesync`.
> Se revirtió el 2026-06-10 al alinear el diseño con la implementación real del engine (adapters
> nativos). El spike `docs/spike-rulesync-vs-rosetta.md` se conserva como análisis de referencia
> para la reevaluación de Fase 5.

---

## 5. El esquema canónico

### 5.1 Tipos de artefacto y mapeo

| Tipo neutral | Qué es | → Claude | → Copilot |
|---|---|---|---|
| `context` | Conocimiento/reglas | `CLAUDE.md` o memoria | `instructions` (con `applyTo`) |
| `command` | Acción/prompt reutilizable | `commands/*.md` | `prompts/*.prompt.md` |
| `agent` | Persona con herramientas | `agents/*.md` | `chatmodes/*.chatmode.md` |
| `skill` | Capacidad con recursos (scripts, refs) | `skills/<n>/SKILL.md` | *degradar (ver ADR-5)* |

### 5.2 `pack.yaml` (manifiesto del pack)

```yaml
name: backend-python
version: 2.4.0                 # semver del pack
tier: domain                  # core | domain | experimental
owners: ["@org/backend-guild"]
targets: [claude, copilot]
depends_on:
  - core@^3.0.0
channels: [stable]            # stable | beta
```

### 5.3 Frontmatter de assets

`command`:
```yaml
---
id: migrate-db                # global, estable
type: command
title: "Run DB migration"
description: "Genera y aplica una migración con validación previa"
targets: [claude, copilot]
---
# cuerpo en markdown neutral, con variables {{project.db_url}}
```

`context` con scope por ruta:
```yaml
---
id: py-conventions
type: context
scope: { applyTo: "**/*.py" }   # el adapter lo traduce a CLAUDE.md / instructions(applyTo)
targets: [claude, copilot]
---
```

`skill` (caso de degradación):
```yaml
---
id: db-tuning
type: skill
resources: [scripts/, references/]
targets:
  claude:  { emit: skill }
  copilot: { emit: prompt }     # override; o se hereda de governance/degradation.yaml
---
```

### 5.4 Principios de diseño del esquema
1. **`id` estable y global** — ancla de versionado y deprecaciones.
2. **Variables/plantillas `{{...}}`** en el cuerpo, para que el adapter inyecte la sintaxis específica de cada herramienta.
3. **Degradación explícita, nunca mágica** (ADR-5).

### 5.5 Registro de degradación

```yaml
# governance/degradation.yaml — reglas por tipo (auditadas por gobernanza)
skill -> copilot:
  strategy: prompt              # prompt | instructions | composite | skip
  template: templates/skill-to-prompt.md
  approved_by: "@org/ai-governance"
  approved_at: 2026-06-05
```

---

## 6. Modelo de gobernanza (estricto)

### 6.1 Tiers de ownership

| Tier | Contenido | Aprobación | Riesgo |
|---|---|---|---|
| `core` | Lo que heredan todos (estilo, seguridad, base) | **Board de gobernanza** (obligatorio) | Alto |
| `domain` | Packs por dominio | Owner del dominio + 1 revisor de gobernanza | Medio |
| `experimental` | Sandbox para incubar | Solo el proponente | Bajo (opt-in) |

### 6.2 Disparadores de RFC obligatorio
- Cambios en **`core`**.
- **Breaking changes** (rename de `id`, cambio de contrato de un command/agent).
- Cambios de **política de seguridad** (permisos, ejecución de comandos, secretos).
- Cambios de **alto impacto de adopción** (cuando la telemetría muestra que >N proyectos consumen el asset).

**No** requieren RFC (van por PR directa con aprobación del owner): nuevo pack, nuevo target/herramienta, nuevas reglas de degradación.

### 6.3 Flujo de contribución

```
Proyecto detecta mejora
   │
   ├─ Cambio pequeño/local → PR directa al pack → review por tier → merge
   │
   └─ Cambio estructural / breaking / seguridad / alto impacto → RFC
          │
       Board evalúa (impacto, adopción, alternativas)
          │
       Aceptado → implementación → review → merge
   │
Release train (quincenal) → changelog → bot propaga → adopción
```

Intake = **PR directa** al monorepo; CODEOWNERS enruta la revisión al tier/dominio correspondiente.

### 6.4 Política de deprecación
- Cada asset/pack declara `deprecated_in` y `removed_in` en su metadata.
- Ventana de soporte garantizada antes de retirar (a definir: p. ej. 2 trains / 1 mes).
- La CI y el bot **avisan a los consumidores** que aún usan un asset deprecado (cruzando con la telemetría de adopción).

---

## 7. CI / validación

Pipeline en cada PR al monorepo:

```
1. Lint de esquema        → frontmatter válido, id único y estable, owners presentes
2. Check de degradación   → FAIL-CLOSED: target sin mapeo declarado rompe (ADR-5)
3. Build de adapters      → compila a .claude/ y .github/; si un adapter falla, rompe
4. Snapshot/diff          → muestra el output generado por herramienta (revisión humana real)
5. Checks de contrato     → breaking change (id renombrado, arg eliminado) ⇒ exige label + RFC
6. Lint de seguridad      → prompt-injection, secretos, comandos peligrosos
7. Gate de release        → publica solo en la ventana del train (excepto hotfix)
```

- **Detección de breaking changes automática**: compara `id` + contrato contra la versión publicada; sin esto el semver sería opcional/manual.
- **Evals de calidad**: diferidos a **fase 2** (casos dorados ejecutados en CI).

---

## 8. Distribución y consumo

### 8.1 Topología

```
                 CASKAi (monorepo, fuente de verdad)
                          │  release train quincenal
              ┌───────────┴───────────┐
              ▼                        ▼
      Artefactos Claude          Artefactos Copilot
        (.claude/ generado)       (.github/ generado)
              └───────────┬───────────┘
                          ▼
            Bot estilo Renovate abre PR
            a cada repo consumidor
                          ▼
              Repo consumidor:
               - caskai.yaml  (qué packs / canal)  ← mantenido a mano
               - caskai.lock           (versiones exactas)  ← generado, alimenta telemetría
               - .claude/          (vendorizado)        ← generado
               - .github/          (vendorizado)        ← generado
```

### 8.2 Manifiesto por proyecto (única pieza manual)

```yaml
# caskai.yaml
channel: stable
packs:
  - core
  - backend-python
  - security
```

### 8.3 Flujo de actualización
1. Sale un train → el bot lee cada `caskai.yaml`, regenera artefactos y abre PR.
2. El equipo revisa el diff y mergea cuando quiere (respeta su autonomía).
3. El `caskai.lock` resultante alimenta el inventario de adopción.

---

## 9. Estructura del repositorio

```
CASKAi/
├─ packs/
│  ├─ core/                       # tier core
│  │  ├─ pack.yaml
│  │  └─ assets/{context,commands,agents,skills}/
│  ├─ backend-python/             # tier domain
│  └─ frontend-react/
├─ adapters/
│  ├─ claude/                     # → plugin + entrada de marketplace
│  └─ copilot/                    # → .github/{instructions,prompts,chatmodes}
├─ governance/
│  ├─ architecture.md             # este documento
│  ├─ degradation.yaml            # mapeos tipo→target (auditados)
│  ├─ rfcs/                       # RFCs aceptados/rechazados
│  ├─ deprecation-policy.md
│  └─ release-trains.md           # calendario quincenal
├─ ci/                            # los 7 gates
├─ schemas/                       # JSON Schema de pack.yaml y frontmatter
├─ CODEOWNERS                     # routing por tier/dominio
└─ tools/
   ├─ build/                      # corre adapters
   ├─ inventory/                  # escaneo de adopción (lock → telemetría)
   └─ bot/                        # genera PRs a consumidores
```

---

## 10. Ciclo de vida end-to-end

```
1.  Autor edita assets/ en formato canónico
2.  PR al monorepo → CODEOWNERS enruta por tier
3.  CI: schema ✓ degradación ✓ build ✓ snapshot ✓ breaking? ✓ seguridad ✓
4.  Si breaking/core/seguridad/alto-impacto → RFC + board
5.  Merge → espera ventana del train (o hotfix)
6.  Release: semver + changelog + artefactos generados
7.  Bot abre PRs a repos consumidores según caskai.yaml
8.  Equipos revisan diff y mergean → caskai.lock actualizado
9.  Inventory escanea locks → telemetría de adopción
10. Telemetría retroalimenta umbrales de RFC y deprecaciones seguras
```

---

## 11. Roadmap por fases

| Fase | Alcance |
|---|---|
| **0 — Fundaciones** | Esquema canónico + JSON Schemas, repo scaffold, pack `core` mínimo, CODEOWNERS. |
| **1 — Compilación** | Adapters Claude y Copilot, gates de CI 1-6, registro de degradación. |
| **2 — Distribución** | `caskai.yaml`/`caskai.lock`, bot de PRs, vendorizado en consumidores piloto. |
| **3 — Gobierno a escala** | Inventario de adopción, deprecation policy operativa, release trains formales. |
| **4 — Calidad** | Evals automatizados (golden tests) en CI. |
| **5 — Extensión** | Nuevas herramientas (Cursor, etc.) vía nuevos adapters nativos. **Punto de reevaluación de `rulesync`** (ADR-10): si los targets llegan a ~4+, valorar delegar la emisión en rulesync para no mantener un adapter por herramienta. |

---

## 12. Preguntas abiertas (a resolver con gobernanza)
- Ventana exacta de soporte en la política de deprecación (¿2 trains? ¿1 mes?).
- Umbral `N` de "alto impacto de adopción" que dispara RFC.
- Composición y SLA de revisión del board de gobernanza.
- ¿`core` puede ser breaking nunca, o solo con ventana de migración extendida?
- Estrategia de `templates/` para degradaciones (qué fidelidad mínima aceptamos).

---

## 13. Glosario
- **Pack**: bundle de assets agrupados por dominio; unidad de versionado y consumo.
- **Asset**: un `context`, `command`, `agent` o `skill` en formato canónico.
- **Adapter**: generador que compila el formato canónico al formato nativo de una herramienta.
- **Vendorizar**: commitear los artefactos generados dentro del repo consumidor.
- **Train**: ventana fija de release (quincenal).
- **Degradación**: estrategia para representar un tipo de asset en una herramienta que no lo soporta nativamente.
```
