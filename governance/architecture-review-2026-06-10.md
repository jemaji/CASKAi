# Revisión de arquitectura — CASKAi

**Tipo:** Evaluación de diseño existente (no es un ADR nuevo)
**Estado:** Borrador para discusión con gobernanza
**Fecha:** 2026-06-10
**Revisor:** Delivery / Arquitectura
**Material revisado:** `governance/architecture.md` (v1.0, ADR-1…10), `README.md`, `CLAUDE.md`, `governance/degradation.yaml`, engine `tools/caskai` (`main.go`, `yaml.go`, ~720 LOC), packs y consumers de ejemplo.

---

## 1. Veredicto general

El diseño es **sólido, coherente y bien razonado**. El modelo mental (packs versionados + bot de PRs estilo Renovate) es acertado y fácil de comunicar; los ADRs documentan alternativas y justificaciones de forma honesta; la separación canónico → adapters → vendorizado envejece bien. La gobernanza por tiers con RFC está bien pensada.

El diseño **no es el riesgo principal**. El riesgo está en (a) una **contradicción entre el ADR-10 y el código actual**, y (b) varios supuestos operativos que aún no están validados y que la POC debería atacar primero.

---

## 2. Hallazgo crítico — el ADR-10 contradice la implementación

| | Lo que dice el diseño | Lo que hace el código hoy |
|---|---|---|
| Emisión a `.claude/` y `.github/` | **ADR-10**: se delega en el binario `rulesync` (subproceso); se descartan adapters propios | El engine **tiene adapters propios en Go** (`emitClaude`, `emitCopilot` en `main.go`); **no hay ninguna invocación a `rulesync`** (`grep` de `rulesync`/`exec.Command` = 0 coincidencias). `adapters/` está vacío. |

Esto significa que el documento maestro, el README y `CLAUDE.md` describen una decisión (delegar en rulesync) que **el engine real no implementa** — sigue con adapters nativos. No es necesariamente un error: los adapters en Go ya funcionan para los 2 targets. Pero **una de las dos fuentes de verdad está obsoleta** y eso es peligroso en un repo cuyo propósito es ser *la fuente de verdad*.

**Decisión a tomar antes de la POC** (es la pregunta arquitectónica de fondo):

### Opción A — Mantener adapters nativos en Go (alinear el doc al código)
| Dimensión | Valoración |
|---|---|
| Complejidad | Baja-Media (ya está hecho) |
| Coste | Mantener 1 adapter por herramienta |
| Escalabilidad | Buena hasta ~3-4 herramientas; lineal por target |
| Familiaridad equipo | Alta (Go propio, sin dependencia externa) |

**Pros:** binario único real, cero dependencias en runtime, control total de fidelidad de variables y del paso de degradación, ya funciona. **Contras:** cada herramienta nueva = adapter nuevo a mano; reinventas lo que rulesync ya resuelve.

### Opción B — Delegar en rulesync (implementar el ADR-10 tal cual)
| Dimensión | Valoración |
|---|---|
| Complejidad | Media-Alta (integrar subproceso + mapear su modelo al canónico) |
| Coste | Mantener la integración + seguir rulesync upstream |
| Escalabilidad | Mejor si rulesync ya soporta la herramienta nueva |
| Familiaridad equipo | Menor (modelo y formato de un tercero) |

**Pros:** te ahorras adapters cuando rulesync ya cubre el target. **Contras:** con **solo 2 targets** el ahorro es marginal y *aun así* tienes que construir encima degradación fail-closed, traducción fiel de variables, acceso por Entra, lockfile, bot y telemetría (lo dice el propio ADR-10). Acoplas tu core a la evolución de un binario externo.

**Recomendación:** dado que hoy hay **2 herramientas** y los adapters nativos **ya existen y funcionan**, lo pragmático para la POC es **Opción A**: alinear el doc al código (marcar ADR-10 como *revertido/diferido*), y **reservar rulesync para cuando el nº de targets justifique el acoplamiento** (Fase 5). Si se prefiere B por estrategia, hay que tratarlo como un spike con criterio de salida medible (¿cuántos adapters ahorra realmente?).

---

## 3. Estado real vs. roadmap

El repo está en **Fase 0 → inicio de Fase 1**, no más allá. Conviene calibrar expectativas de la POC:

- Packs: solo `core` (con assets de ejemplo) y `backend-python` (solo `pack.yaml`, sin assets). El `frontend-react` del doc no existe.
- `adapters/`, `ci/`, `schemas/`, `tools/bot/`, `tools/inventory/` descritos en §9 del doc: **no presentes** (la emisión vive dentro de `main.go`).
- Los 7 gates de CI (§7): **diseñados, no implementados** como pipeline.
- Bot de PRs, lockfile y telemetría de adopción: **no implementados**.

No es una crítica —es una POC— pero el doc se lee como si todo existiera. Sugiero una columna "estado" (planificado / en progreso / hecho) en §9 y §11.

---

## 4. Riesgos de diseño a vigilar

1. **Control de acceso "en build" ≠ confidencialidad.** El acceso por grupos de Entra se aplica al *generar*, pero los artefactos se **vendorizan** (se commitean) en cada repo. Cualquiera con acceso al repo consumidor lee el contenido. El modelo real es *need-to-consume*, no *secreto*. Para `confidential` de verdad haría falta no-vendorizar o cifrar. **Acción:** documentar explícitamente el modelo de amenaza para que nadie meta secretos reales en un pack `confidential` creyendo que están protegidos (ya lo prohíbe `CLAUDE.md`, pero conviene explicar el porqué).

2. **Fatiga de PRs a escala (100+ repos).** Vendorizar Claude *y* Copilot genera diffs grandes y ruidosos en cada train. Sin política de auto-merge / agrupación, los equipos ignorarán las PRs y la adopción se estancará. **Acción:** definir en la POC la estrategia del bot (auto-merge para `core` no-breaking, batching, etiquetado).

3. **Resolución de dependencias entre packs.** `depends_on: core@^3.0.0` implica un *solver* de versiones transitivas. Aún no existe. A 100+ repos y varios packs con rangos semver, esto se complica rápido. **Acción:** decidir pronto si soportas rangos (`^`) o solo versiones fijas en la POC.

4. **Cuello de botella de gobernanza vs. trains quincenales.** El SLA y composición del board está en "preguntas abiertas" (§12). Si el board es lento, el train quincenal se incumple y la vía hotfix se abusa. **Acción:** cerrar el SLA del board antes de prometer cadencia quincenal.

5. **Fidelidad de la degradación skill→prompt (Copilot).** Las skills pierden `scripts/` y `resources/` al degradarse a prompt. El engine ya inserta un comentario de aviso (`<!-- degradado… -->`), bien. Pero una skill rica degradada puede quedar inservible. **Acción:** definir la fidelidad mínima aceptable (está en §12) y considerar `skip` explícito cuando la skill dependa fuertemente de scripts.

---

## 5. Lo que está especialmente bien

- **ADR-5 (degradación fail-closed con mapeo auditado):** excelente. Combina disciplina, memoria y escalabilidad; el `degradation.yaml` por tipo + override por-asset es el patrón correcto.
- **ADR-4 (`id` global y estable como ancla de versionado):** evita el clásico bug de anclar a nombres de fichero.
- **Analogía Renovate/Dependabot:** baja la barrera de entrada del concepto a cero.
- **No-objetivos explícitos** (evals y runtime telemetry diferidos): foco honesto para una v1.

---

## 6. Action items sugeridos para la POC

1. [ ] **Resolver la contradicción del ADR-10**: decidir Opción A o B y **alinear doc ↔ código** (es el bloqueante #1).
2. [ ] Añadir columna de **estado real** a §9 y §11 del doc de arquitectura.
3. [ ] Implementar **al menos un gate de CI real** (lint de esquema + degradación fail-closed) sobre `caskai validate`.
4. [ ] Completar los assets de **`backend-python`** para tener un segundo pack no trivial y probar `depends_on`.
5. [ ] Ejecutar el **flujo e2e** sobre un consumer de ejemplo (`payments-api`) y validar el output vendorizado de ambas herramientas.
6. [ ] Documentar el **modelo de amenaza** del acceso build-time (riesgo #1).
7. [ ] Definir la **política del bot** (auto-merge / batching) antes de escalar (riesgo #2).
8. [ ] Cerrar **SLA del board** y **umbral N de alto impacto** (preguntas abiertas §12) — condicionan trains y RFC.

---

## 7. Consecuencias

- **Se facilita:** tener un doc y un código que dicen lo mismo; una POC con criterios de salida claros en vez de "construir toda la plataforma".
- **Se complica:** habrá que elegir y posiblemente revertir formalmente el ADR-10 (con su propio mini-RFC, ya que toca decisión de arquitectura).
- **A revisitar:** rulesync como capa de emisión cuando el nº de herramientas objetivo pase de 2 a 4+.
