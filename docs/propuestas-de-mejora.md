# Propuestas de mejora — CASKAi

> Documento vivo. Se actualiza a medida que surgen ideas durante el desarrollo de la PoC.
> Las propuestas aprobadas se convierten en RFCs o tareas de fase siguiente.

---

## [MEJ-001] Proxy de descarga autenticado con Entra ID

**Contexto:** La descarga del binario `caskai` vía `install.sh`/`install.ps1` es anónima.
GitHub solo ofrece counts agregados por asset de release, sin identificar al usuario.

**Propuesta:** Añadir un proxy ligero (Azure Function o Cloudflare Worker) entre el instalador
y la release de GitHub que:
1. Exige un token de Entra ID (device flow en el script de instalación).
2. Registra `{user, email, os, arch, version, timestamp}` en un log centralizado.
3. Redirige al asset correcto de la release de GitHub tras validar el token.

**Beneficio:** Visibilidad real de quién ha descargado qué versión del engine — útil para
notificaciones de seguridad y métricas de adopción por usuario.

**Dependencia:** Reutiliza la misma infraestructura de autenticación Entra que se usará
para el control de acceso a packs restringidos (Fase 2). Recomendado implementar en paralelo
a esa fase para amortizar el coste de integración.

**Complejidad estimada:** Media (2-3 días). Baja prioridad hasta tener el bot operativo.

---

## [MEJ-002] Telemetría de uso opcional en el binario (`caskai register`)

**Contexto:** Alternativa ligera a MEJ-001 sin necesidad de autenticación.

**Propuesta:** Al instalar, `caskai` hace un POST anónimo a un endpoint con
`{machine_id, os, arch, version}`. No identifica al usuario, solo cuenta máquinas activas
y versiones en uso.

**Beneficio:** Métricas de adopción sin fricción de autenticación. Opt-out con variable
de entorno `CASKAI_NO_TELEMETRY=1`.

**Complejidad estimada:** Baja (1 día).

---

## [MEJ-003] `targets` en `ai.manifest.yaml` (consumidor declara herramientas objetivo)

**Contexto:** Actualmente los targets (claude/copilot) se declaran en el frontmatter de
cada asset. El consumidor no puede restringir para qué herramientas quiere el build.

**Propuesta:** Añadir campo `targets` al manifiesto del consumidor:
```yaml
channel: stable
targets: [claude]        # solo genera .claude/, no .github/
packs:
  - core
```
`caskai build` solo emite para esos targets y da error si hay un asset sin mapeo para
alguno de los solicitados.

**Beneficio:** Consumidores que solo usan Claude Code no generan ficheros de Copilot
innecesarios. Control explícito y detección temprana de assets sin mapeo.

**Complejidad estimada:** Baja (medio día en el engine + tests).

---

## [MEJ-004] Bot de distribución automática (GitHub App)

**Contexto:** Actualmente la actualización de los ficheros generados en consumidores es manual.

**Propuesta:** GitHub App que detecta nuevas versiones de packs en CASKAi y abre PRs
automáticos en cada repo consumidor con los ficheros regenerados y el `ai.lock` actualizado.
Idéntico al modelo Dependabot/Renovate.

**Beneficio:** Elimina la fricción de actualización manual. Un cambio aprobado en CASKAi
llega a todos los consumidores sin intervención humana.

**Complejidad estimada:** Alta (1-2 semanas). Pieza central de la Fase 3.

---

## [MEJ-005] Escalado masivo de herramientas con rulesync (Fase 5)

**Contexto:** El engine usa adapters nativos en Go (ADR-10) para emitir a `.claude/` y
`.github/`. Escala bien hasta ~10 herramientas objetivo.

**Propuesta:** Evaluar `rulesync` como capa de emisión cuando el número de herramientas
objetivo crezca significativamente (Cursor, Windsurf, etc.) y el mantenimiento de adapters
nativos sea costoso.

**Referencia:** `docs/spike-rulesync-vs-rosetta.md` — reevaluar en Fase 5.

**Complejidad estimada:** Media-Alta. No bloquea nada hasta escalar a 5+ herramientas.

---

## [MEJ-006] Evals automatizados de calidad de assets

**Contexto:** No hay validación semántica del contenido de los assets, solo estructural
(schema, degradación).

**Propuesta:** Suite de evals que puntúa la calidad de un asset (claridad, completitud,
coherencia con el pack) usando un modelo. Se ejecuta en CI como gate opcional.

**Complejidad estimada:** Media. Diferido a Fase 2 según arquitectura.md.
