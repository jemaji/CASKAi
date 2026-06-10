# Análisis de mercado — ¿existe algo que haga lo nuestro?

> Pregunta: ¿hay herramientas que hagan lo que `CASKAi` necesita, o una aproximación?
> Respuesta corta: **el "compilar a varias herramientas" es ya un commodity** (lo resuelven
> varios OSS); el **"registro + gobernanza + compartición"** tiene análogos comerciales
> fuertes (Continue Hub) y un acelerador muy cercano (Grid Dynamics Rosetta). Pero **ninguno
> combina las 6 capacidades** que pedimos, en especial **visibilidad de packs por grupo de
> Entra + propagación automática a repos existentes + gobernanza por tiers (promoción a core)**.

## Las 6 capacidades que necesitamos

- **C1** Multi-herramienta: una fuente → Claude Code + Copilot (y extensible)
- **C2** Catálogo versionado + consumo selectivo por repo
- **C3** Propagación automática a repos existentes (bot de PRs, estilo Renovate)
- **C4** Gobernanza estricta: tiers, RFC, **promoción a core**, CODEOWNERS
- **C5** Seguridad/visibilidad de packs **por rol (grupos de Entra)** + auditoría
- **C6** Trazabilidad 100% del uso (inventario de adopción)
- **C7** On-prem / dentro del perímetro + integración GitHub Enterprise + Entra

## Panorama por categorías

### 1. Compiladores multi-herramienta (OSS) — resuelven **C1**
- **rulesync** — genera/sincroniza configs para 20+ herramientas (Claude, Cursor, Copilot, Windsurf, Cline, Gemini…) desde `.rulesync/`.
- **ruler** — "una fuente de verdad para todos los agentes"; distribuye a cada fichero nativo.
- **rule-porter** — convierte reglas Cursor → Claude/Copilot/AGENTS.md.

**Veredicto:** hacen exactamente nuestra capa de *adapters* (ADR-1). **Foco en reglas/instructions**; commands/skills/agents lo cubren parcialmente. Sin registro versionado, sin distribución a 100+ repos, sin gobernanza, sin acceso por rol, sin trazabilidad. → **Candidatos a reutilizar para no mantener adapters propios.**

### 2. Estándar AGENTS.md — convención para **C1** (formato canónico)
Estándar abierto emergente de "single source of truth" con **ficheros anidados por monorepo** (precedencia por cercanía; OpenAI usa ~88). Útil como base de nuestro **formato canónico**, pero es una convención, no gobierna ni distribuye.

### 3. Continue Hub (comercial) — el análogo de producto más cercano a **C2/C4/C5**
Registro tipo "Docker Hub" de *blocks* (rules, prompts, context, docs, **MCP**, models) y *assistants*, con **gobernanza de organización** (roles admin/member, visibilidad privado/equipo/público, políticas sobre qué blocks se pueden usar; tiers Solo/Teams/Enterprise).
**Gaps para nosotros:** centrado en el asistente/IDE de Continue (no genera nativamente `.claude/`+`.github/` para Claude/Copilot como ciudadanos de primera); hub **SaaS** (perímetro/datos); visibilidad por rol propia, **no por grupos de Entra por pack**; no hace **PRs automáticos a repos existentes**.

### 4. Grid Dynamics Rosetta (acelerador + OSS) — el más cercano a la **visión completa**
"Gestión centralizada de instrucciones": rules/skills/workflows/sub-agents **versionados, desplegados como código**, distribuidos a cualquier agente/IDE, **dentro del perímetro del cliente**, con **gobernanza y visibilidad centralizadas del footprint**. Conceptualmente cubre C1, C2, C4, C7 y parte de C6.
**A verificar:** ¿visibilidad por grupo de Entra (C5)? ¿propagación automática estilo bot (C3)? ¿integración GitHub Enterprise? → **Merece un spike serio: podría ser base o referencia.**

### 5. Registros de agentes cloud (Microsoft Agent 365 / Google Gemini Agent Platform / AWS Agent Registry) — **otra capa**
Catálogos gobernados de **agentes en runtime** (identidad, owner, versiones, IAM/SSO, métricas de adopción, políticas de ciclo de vida). Gran patrón de gobernanza (C4/C5/C6) **pero para agentes desplegados, no para distribuir configuración de asistentes IDE a repos**. Relevante si más adelante operamos agentes; no resuelve nuestro problema de hoy.

### 6. GitHub Copilot nativo — parcial y solo Copilot
**Org custom instructions** (GA abr-2026) + instructions por repo/ruta + herencia vía **template repos**. Útil y nativo, pero: solo Copilot, solo algunas superficies (chat / code review / cloud agent en github.com), instrucción org **única** (sin packs/versiones/selección), sin visibilidad por grupo, sin trazabilidad. (El patrón de **preset compartido de Renovate** sí inspira nuestra C3.)

### 7. MCP Gateway/Registry (OSS) — para el eje MCP/tools + patrón Entra
Centraliza **servidores MCP** con OAuth (**Keycloak/Entra**), acceso gobernado y auditable para agentes y asistentes. No distribuye packs, pero es **buen patrón de integración Entra + auditoría** si añadimos MCP a los packs.

## Tabla comparativa (capacidad × solución)

| | C1 multi-tool | C2 versionado+selectivo | C3 propagación auto | C4 gobernanza/tiers | C5 acceso por grupo Entra | C6 trazabilidad | C7 on-prem+Entra |
|---|:--:|:--:|:--:|:--:|:--:|:--:|:--:|
| rulesync / ruler (OSS) | ✅ | ◐ | ✗ | ✗ | ✗ | ✗ | ✅ |
| Continue Hub | ◐ | ✅ | ✗ | ◐ | ◐ | ◐ | ✗ (SaaS) |
| Grid Dynamics Rosetta | ✅ | ✅ | ◐ | ✅ | ❓ | ◐ | ✅ |
| Registros agentes cloud | ✗* | ✅ | ✗* | ✅ | ✅ | ✅ | ◐ |
| Copilot nativo (org) | ✗ | ✗ | ◐ | ✗ | ✗ | ✗ | ✅ |
| **CASKAi (nuestro)** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

`✅ sí · ◐ parcial · ✗ no · ❓ por verificar · * otra capa (agentes runtime)`

## Veredicto: build vs buy

- **No reinventar el compilador.** C1 es un commodity → **adoptar rulesync o ruler** como capa de adapters reduce justo el coste que nos preocupaba del enfoque "canónico neutral". (O mantener `caskai build`, pero conviene comparar.)
- **El diferencial real es la capa de gobierno + seguridad + distribución** integrada con **GitHub Enterprise + Entra** (C3+C4+C5+C6), que **ningún producto ofrece junta**. Eso es lo que justifica construir (nuestro `caskai` + bot + API).
- **Antes de comprometer "todo a medida": evaluar Rosetta** (puede cubrir gran parte y vive en el perímetro del cliente) y **Continue Hub** (si se aceptara SaaS y su asistente). Un spike evita construir lo ya hecho.

## Acciones recomendadas (spikes, ~1 semana)
1. **Rosetta**: instalar y mapear contra nuestras C1–C7 (sobre todo C5 Entra y C3 propagación). ¿Base, referencia o descartar?
2. **rulesync/ruler**: PoC de usarlo como adapter detrás de nuestro formato → ¿retiramos adapters propios?
3. **Continue Hub Enterprise**: ¿permite self-host/perímetro y generación nativa Claude+Copilot? Si no, descartar como core.
4. Confirmar que **C3+C5+C6** (propagación + acceso por Entra + trazabilidad) siguen siendo "construir": es nuestro foso.

## Fuentes
- rulesync — https://github.com/dyoshikawa/rulesync
- ruler — https://github.com/intellectronica/ruler
- rule-porter — https://dev.to/nedcodes/rule-porter-convert-cursor-rules-to-claudemd-agentsmd-and-copilot-4hjc
- AGENTS.md (estándar) — https://tessl.io/blog/the-rise-of-agents-md-an-open-standard-and-single-source-of-truth-for-ai-coding-agents/
- Continue Hub — gobernanza — https://docs.continue.dev/hub/governance/org-permissions
- Continue (lanzamiento/visión) — https://techcrunch.com/2025/02/26/continue-wants-to-help-developers-create-and-share-custom-ai-coding-assistants/
- Grid Dynamics Rosetta — https://github.com/griddynamics/rosetta · https://griddynamics.github.io/rosetta/
- GitHub Copilot org custom instructions (GA) — https://github.blog/changelog/2026-04-02-copilot-organization-custom-instructions-are-generally-available/
- Renovate (preset compartido) — https://github.com/renovatebot/renovate
- MCP Gateway/Registry (Entra/OAuth) — https://github.com/agentic-community/mcp-gateway-registry
- Registros de agentes cloud — Microsoft Agent 365 · Google Gemini Enterprise Agent Platform · AWS Agent Registry (AgentCore)
