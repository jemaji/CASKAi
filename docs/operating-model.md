# Modelo operativo de `CASKAi`

> Cómo se crean, consumen, evolucionan y gobiernan los activos de IA.
> Audiencia: equipos de ingeniería que van a producir y/o consumir packs.

## La analogía

`CASKAi` es **un gestor de paquetes interno para activos de IA** — como `npm` + Renovate, pero para `agents`, `skills`, `commands` y `contexts`, y **multi-herramienta** (Claude Code + GitHub Copilot).

> Si tu equipo entiende **Dependabot/Renovate**, ya entiende el 80% de esto.

Una fuente de verdad (monorepo) publica **packs versionados**; cada proyecto declara en un manifiesto qué packs quiere; un **bot** mantiene los ficheros sincronizados en cada repo automáticamente.

## Actores

| Actor | Rol |
|---|---|
| **Equipo de gobernanza** | Dueños del pack `core` y del proceso. Aprueban lo crítico. |
| **Owners de dominio** | Mantienen los packs de su área (`backend-python`, etc.). |
| **Consumidores** | Devs de los 100+ repos. Usan packs y proponen mejoras. |
| **El bot** | GitHub App que distribuye actualizaciones automáticamente. |
| **CI del monorepo** | Hace cumplir las reglas (los 7 gates). |

---

## Flujo A — Crear (autoría)

```
Autor edita packs/<pack>/assets/**   (formato canónico: markdown + frontmatter)
   │  CASKAi build --preview        ← ve el output Claude/Copilot en local
   ▼
PR al monorepo
   │  CODEOWNERS enruta al owner del pack/tier
   │  CI: schema ✓ degradación ✓ build ✓ snapshot ✓ breaking? ✓ seguridad ✓
   ▼
¿breaking / core / seguridad / alto impacto?  ──Sí──▶  RFC + board
   │ No
   ▼
Merge → bump de versión del pack + changelog → espera al release train
```

Los autores solo escriben el **formato canónico**. Los `.claude/` y `.github/` **nunca se editan a mano**: los genera el build (adapters).

---

## Flujo B — Consumir (onboarding de un repo)

Un consumidor crea **un único fichero**:

```yaml
# ai.manifest.yaml — la única pieza que se mantiene a mano
channel: stable
packs:
  - core
  - backend-python      # consumo selectivo: solo lo que interesa
```

Dos modos de materialización:

- **Automático (recomendado):** el bot detecta el manifiesto y abre un PR con `.claude/`, `.github/` y `ai.lock` ya generados. El dev no instala nada.
- **On-demand (local o CI):** instala el engine y ejecuta el build directamente:
  ```bash
  # Instalar el engine (una vez por máquina):
  curl -fsSL https://raw.githubusercontent.com/jemaji/CASKAi/main/install.sh | bash
  # Windows: irm https://raw.githubusercontent.com/jemaji/CASKAi/main/install.ps1 | iex

  # Generar los ficheros:
  caskai build --root ~/CODE/CASKAi --manifest ai.manifest.yaml --out .
  ```

Resultado en el repo del consumidor:
```
su-repo/
├─ ai.manifest.yaml   ← a mano
├─ ai.lock            ← generado (versiones exactas + integridad)
├─ .claude/           ← generado (lo usa Claude Code)
└─ .github/           ← generado (lo usa Copilot)
```

---

## Flujo C — Actualizar (automático)

```
Sale el release train quincenal (core 0.1.0 → 0.2.0)
   │
   ▼
El bot recorre TODOS los repos con ai.manifest
   │  por cada uno: lee manifiesto → regenera artefactos → calcula diff
   ▼
Abre un PR por repo: "chore(ai): actualizar core 0.1.0 → 0.2.0"
   ├─ diff de .claude/ y .github/
   ├─ cambio en ai.lock
   └─ resumen del changelog
   ▼
El equipo del repo revisa y mergea CUANDO QUIERE   ← autonomía
```

Idéntico a Dependabot. Para **hotfix de seguridad**, el bot puede marcar el PR urgente o auto-merge según política.

---

## Flujo D — Proponer cambios (contribución)

**Regla de oro:** los consumidores **no editan** los ficheros generados (el bot los sobrescribe). Si quieren cambiar algo, lo proponen **upstream**.

```
Dev consumidor detecta una mejora
   │
   ▼
PR directa al MONOREPO, editando el asset canónico
   │  CODEOWNERS enruta al owner; CI valida; si estructural → RFC + board
   ▼
Aprobado y mergeado → próximo train
   ▼
El bot lo propaga a TODOS los consumidores de ese pack
```

**Efecto multiplicador:** una mejora de un equipo, aprobada, llega a los 100+ repos.

> Overrides locales: si un repo necesita algo propio que no debe subir, se permite una carpeta de overrides que el bot no toca (política de gobernanza).

---

## Flujo E — Retirar (deprecación segura)

```
Asset marcado: deprecated_in 0.3.0, removed_in 0.5.0
   ▼
El bot/CI avisa SOLO a quien aún lo usa
   (cruzando todos los ai.lock = inventario de adopción)
   ▼
Ventana de soporte (p. ej. 2 trains) → eliminación
```

Deprecación **con datos**: la suma de todos los `ai.lock` dice quién usa qué versión.

---

## El ciclo en una frase

> Se autora en formato canónico → CI y gobernanza validan → se publica versionado en un train → el bot abre PRs a todos los consumidores → ellos mergean → y cualquiera puede proponer mejoras de vuelta que, aprobadas, llegan a todos.

Ver también: [`architecture.md`](../governance/architecture.md) (decisiones de diseño) y [`security-and-access-control.md`](./security-and-access-control.md) (control de acceso con Entra ID).
```
