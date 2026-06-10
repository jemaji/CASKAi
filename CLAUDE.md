# CASKAi — guía para agentes de código

CASKAi es la **fuente de verdad gobernada** de los activos de IA de desarrollo (Context, Agents,
Skills, Knowledge) servidos a Claude Code y GitHub Copilot. Este repo contiene **los packs
(fuente de verdad)** y **el engine `caskai` (Go)**. Trabaja respetando el modelo y la gobernanza.

## Modelo mental
- **Pack**: bundle de assets por dominio; unidad de versionado (semver) y consumo. Vive en `packs/<pack>/`.
- **Asset**: uno de 4 tipos canónicos (markdown + frontmatter) en `packs/<pack>/assets/`:
  - `context` → reglas/conocimiento · `command` → acción/prompt · `agent` → persona · `skill` → capacidad con recursos.
- **Adapters/emisión**: el engine (`caskai`) compila el formato canónico a `.claude/` y `.github/`
  con **adapters nativos en Go** (ADR-10); `rulesync` es propuesta de futuro (Fase 5), no se usa hoy.
  Nunca edites los ficheros generados de un consumidor: se regeneran.
- **Distribución**: vendorizada (ficheros en cada repo) vía bot de PRs. Consumo declarado en `caskai.yaml`.

## Reglas de trabajo (importantes)
1. **No edites ficheros generados** (`.claude/`, `.github/`, `caskai.lock` en repos consumidores): se regeneran.
2. **Crear/editar packs** se hace en `packs/<pack>/assets/` en formato canónico. Todo asset lleva
   `id` (estable, global) y `type`. Todo pack lleva `owners`, `tier` y `access`.
3. **Gobernanza por tiers**: `core` (board `@org/ai-governance`) · `domain` (owner del dominio) ·
   `experimental` (proponente). CODEOWNERS enruta la revisión.
4. **RFC obligatorio** para: cambios en `core`, breaking changes, política de seguridad y alto
   impacto de adopción. Nuevo pack / nuevo target / regla de degradación → PR directa con owner.
5. **Degradación fail-closed**: si un asset apunta a una herramienta que no soporta su tipo (p. ej.
   `skill`→Copilot) sin mapeo en `governance/degradation.yaml` ni override, **el build rompe**. No
   degradar en silencio.
6. **Seguridad**: cada pack declara `access.classification` (`internal`/`restricted`/`confidential`)
   y `allowed_groups` (grupos de Entra). Nunca metas secretos/PII en un asset.
7. **Versionado por pack** (semver). Promocionar un asset a `core` sube la versión de `core` y exige RFC.

## Instalación del engine
```bash
# Mac / Linux (auto-detecta OS y arquitectura):
curl -fsSL https://raw.githubusercontent.com/jemaji/CASKAi/main/install.sh | bash

# Windows (PowerShell):
irm https://raw.githubusercontent.com/jemaji/CASKAi/main/install.ps1 | iex

# O compilar desde fuente (requiere Go):
go build -o ~/bin/caskai ./tools/caskai
```

## Comandos del engine
```bash
caskai version                                          # verifica la instalación
caskai validate                                         # gates: schema, degradación fail-closed
caskai access    --manifest <consumidor>/caskai.yaml   # visibilidad por rol (audita)
caskai build     --manifest <consumidor>/caskai.yaml --out <consumidor>/
caskai inventory --consumers <dir-con-locks>            # trazabilidad de adopción (lee caskai.lock)
caskai promote   --asset <pack>/assets/<sub> --to core
python3 tools/codeowners-route.py <ficheros>            # qué owners exige CODEOWNERS
```

## Antes de dar por bueno un cambio
- Ejecuta `caskai validate` (debe pasar).
- Si tocaste `packs/core/` o hiciste un cambio incompatible/seguridad → recuerda el **RFC** y el routing a `@org/ai-governance`.
- Mantén el estilo del código y de los assets que te rodean.

## Más contexto
`governance/architecture.md` (ADRs y diseño) · `docs/operating-model.md` (flujos) ·
`docs/security-and-access-control.md` (Entra) · `docs/flujo-e2e.md` (recorrido ejecutable) ·
`docs/spike-rulesync-vs-rosetta.md` (análisis de referencia; reevaluación de emisión en Fase 5).
