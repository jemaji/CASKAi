# Flujo end-to-end (ejecutable) — de owner a core, con Go, seguridad por rol y trazabilidad

> Este recorrido **se ejecuta de verdad** sobre el repo. El engine es el binario Go
> `bin/caskai` (compilado, sin dependencias). Reproduce: creación de un pack por un
> owner, aprobación de gobernanza, promoción a `core`, control de acceso por grupo de
> Entra y trazabilidad 100% del uso.

## 0. El engine: dónde entra Go y cómo

`caskai` es el **engine de la plataforma**, escrito en Go y compilado a **un único
binario de ~3 MB sin dependencias**. Es lo que corre:
- en **CI** (gate de validación en cada PR),
- en el **bot** (build + control de acceso al generar los PRs a consumidores),
- en **local** (un dev puede previsualizar).

```bash
# se compila una vez, corre en cualquier sitio sin instalar nada
go build -o bin/caskai ./tools/caskai
```

Subcomandos: `validate` · `build` · `access` · `inventory` · `promote`.

---

## 1. Un owner crea un pack (PR #1)

El owner del dominio crea `packs/backend-python/` (con `pack.yaml`: `owners`,
`tier: domain`, `access.classification: restricted`, `allowed_groups`).

```bash
git checkout -b feat/backend-python-pack
git add packs/backend-python && git commit -m "feat(backend-python): nuevo pack de dominio"
```

**CODEOWNERS enruta la revisión** al owner del dominio (no a gobernanza):

```
Revisores requeridos por CODEOWNERS:
  @org/backend-guild           (3 fichero/s)
```

**CI (gate) con el engine Go:**

```
$ caskai validate
  pack backend-python@0.1.0 (tier=domain, restricted)
  pack core@0.1.0 (tier=core, internal)
✅ validación OK
```

Aprobado por `@org/backend-guild` → merge a `main`.

---

## 2. Gobernanza promociona un asset a `core` (PR #2)

El context `secure-logging` resultó útil para toda la organización. Gobernanza lo
**promociona de dominio a core** con el engine:

```bash
caskai promote --asset backend-python/assets/context/secure-logging.md --to core
#   → packs/core/assets/context/secure-logging.md
# se sube core 0.1.0 → 0.2.0 y se adjunta RFC-0001
```

**Aquí cambia el aprobador**: al tocar `packs/core/`, CODEOWNERS exige al **board**:

```
Revisores requeridos por CODEOWNERS:
  @org/ai-governance           (3 fichero/s)
```

Aprobado por el **BOARD `@org/ai-governance`** → merge. `core` pasa a `0.2.0` e
incluye `secure-logging` (ver `governance/rfcs/0001-promote-secure-logging-to-core.md`).

**Efecto inmediato**: cualquier consumidor de `core` lo hereda. Al reconstruir
`data-platform` (que solo consume `core`):

```
$ caskai build --manifest consumers/data-platform/ai.manifest.yaml --out consumers/data-platform
  ✓ core@0.2.0 [internal] PERMITIDO
.github/instructions: coding-conventions.instructions.md  secure-logging.instructions.md   ← nuevo
```

Historial real (dos PRs revisados por distintos owners):
```
*   Merge PR #2: promocion a core (aprobado por BOARD @org/ai-governance)
| * feat(core): promueve secure-logging a core 0.2.0 (RFC-0001)
*   Merge PR #1: backend-python (aprobado por @org/backend-guild)
| * feat(backend-python): nuevo pack de dominio
* chore: scaffold CASKAi
```

---

## 3. Seguridad y visibilidad de packs por rol (grupos de Entra)

Cada pack declara su `access` en `pack.yaml`. El engine **deniega en fail-closed** y
**audita** cada decisión. Dos consumidores con distinto grupo de Entra:

```
$ caskai access --manifest consumers/example-app/ai.manifest.yaml   # grupo platform-core
  🔒 backend-python   restricted     DENEGADO
  ✓  core             internal       PERMITIDO

$ caskai access --manifest consumers/payments-api/ai.manifest.yaml  # grupo backend-guild
  ✓  backend-python   restricted     PERMITIDO
  ✓  core             internal       PERMITIDO
```

En el `build`, lo denegado **no se materializa** (no llega ni un fichero):

```
$ caskai build --manifest consumers/example-app/ai.manifest.yaml --out consumers/example-app
  ✓ core@0.1.0 [internal] PERMITIDO
  🔒 backend-python [restricted] DENEGADO (requiere [backend-guild]) — no se materializa
```

Cada decisión queda en `governance/audit.log` (JSON line, auditable):
```json
{"action":"build","actor":"example-app","pack":"backend-python","classification":"restricted","decision":"DENEGADO","groups":["platform-core"],"ts":"..."}
```

> En producción el grupo no se auto-declara: viene de los **claims del token de Entra**
> (OIDC en CI / MSAL en local). El binario es el mismo; cambia de dónde lee los grupos.

---

## 4. Trazabilidad 100% del uso

Cada consumidor lleva un `ai.lock` vendorizado (pack@versión + hash de integridad).
El engine los escanea y produce el inventario de adopción:

```
$ caskai inventory
== inventario de adopción (trazabilidad) — 3 consumidores ==
  PACK               VERSIÓN    USOS    CONSUMIDORES
  backend-python     0.1.0      1       payments-api
  core               0.1.0      2       example-app, payments-api
  core               0.2.0      1       data-platform
    ⚠ deriva de versión en "core": [0.1.0 0.2.0] conviven
```

Esto da, con datos reales: **quién usa qué pack y en qué versión**, detección de
**deriva** (quién va atrasado tras la promoción) y la base para **deprecar con
seguridad** y para que el **bot abra PRs de actualización** a los rezagados.

---

## Cómo reproducirlo

```bash
go build -o bin/caskai ./tools/caskai          # compilar el engine
./bin/caskai validate                          # gate de CI
./bin/caskai access    --manifest consumers/example-app/ai.manifest.yaml
./bin/caskai build     --manifest consumers/payments-api/ai.manifest.yaml --out consumers/payments-api
./bin/caskai inventory
```

| Requisito | Dónde se ve |
|---|---|
| Owner crea pack → gobernanza aprueba | §1 (CODEOWNERS @backend-guild) + merge |
| Promoción a core (board) | §2 (CODEOWNERS @ai-governance + RFC-0001 + core 0.2.0) |
| Dónde entra Go y cómo | §0 + `tools/caskai/` (binario único, CI/bot/local) |
| Seguridad y visibilidad por rol | §3 (access fail-closed + audit.log) |
| Trazabilidad 100% del uso | §4 (`ai.lock` + `caskai inventory`) |
