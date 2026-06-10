# Seguridad y control de acceso

> Cómo se securizan los packs y skills por grupos de **Entra ID**, y cómo se protege la cadena de suministro.
> Plataforma: **GitHub Enterprise** (SSO Entra + team sync) · enforcement **híbrido** · auth **OIDC federation + MSAL**.

## 1. Modelo de amenazas (qué protegemos)

| Amenaza | Mitigación principal |
|---|---|
| Consumo no autorizado de un pack sensible | Autorización por grupos de Entra en distribución (§4, §5) |
| Pack malicioso / con prompt-injection | Lint de seguridad en CI (gate 6) + releases firmados (§7) |
| Modificación no autorizada de un pack | CODEOWNERS + branch protection (plano de escritura, §3) |
| Manipulación en tránsito / suplantación de artefacto | Firmas + hashes de integridad en `ai.lock` (§7) |
| Fuga de credenciales de CI | OIDC federation sin secretos (§6) |
| Falta de trazabilidad | Auditoría: Entra sign-in + logs de API + inventario `ai.lock` (§8) |

## 2. Principio rector: Entra como única fuente de identidad

La pertenencia a grupos se gestiona **una sola vez en Entra ID**; el resto la hereda.

```
Grupos de Entra ID
   ├──(team sync de GitHub Enterprise)──▶ GitHub Teams ──▶ CODEOWNERS / acceso a repos   [ESCRITURA]
   └──(claims de grupo en el token OIDC/MSAL)──────────▶ API de distribución / bot        [LECTURA]
```

### Nota sobre el team sync (dependencia de sistemas)
- El **plano de lectura/consumo NO depende del team sync**: usa los claims de grupo del token de Entra directamente. Es lo que securiza skills/packs por grupo.
- El **team sync solo facilita el plano de escritura** (CODEOWNERS con GitHub Teams), y solo afecta a los equipos *owner* (gobernanza + dominios), no a los consumidores.
- **Fallback** si sistemas no habilita team sync org-wide: mantener esos pocos teams owner vía SCIM o manualmente. Fricción acotada; no bloquea la seguridad de consumo.

> Acción: involucrar al equipo de sistemas/IdP pronto para el team sync, pero el diseño **no se bloquea** si tarda.

## 3. Plano de ESCRITURA — quién crea/aprueba packs

Se securiza íntegramente en GitHub:
- **CODEOWNERS** por carpeta de pack → owners = GitHub Teams (sincronizados de Entra).
  - `packs/core/**` → `@org/ai-governance`
  - `packs/backend-python/**` → `@org/backend-guild`
- **Branch protection**: revisiones obligatorias, status checks de CI, sin push directo a `main`.
- **Tiers** (ver `architecture.md`): `core` exige board; `domain`, owner + 1 gobernanza; `experimental`, proponente.

## 4. Plano de LECTURA — quién consume un pack

### Clasificación de packs
Cada pack declara su nivel y los grupos autorizados en `pack.yaml`:

```yaml
access:
  classification: restricted          # internal | restricted | confidential
  allowed_groups:                     # nombres lógicos → mapeados a object IDs de Entra
    - security-engineering
    - platform-core
```

| Clasificación | Quién consume | Dónde vive | Enforcement |
|---|---|---|---|
| `internal` | Toda la ingeniería | Monorepo | Ninguno extra (default) |
| `restricted` | Solo `allowed_groups` | Monorepo | **Bot + API** validan grupo (§5) |
| `confidential` | Solo `allowed_groups` | **Repo aislado** | Acceso de repo por GitHub Team (Entra) **+** bot/API |

`confidential` añade aislamiento físico (repo aparte con read restringido) como defensa en profundidad: ni siquiera el código fuente es visible fuera del grupo.

## 5. Puntos de enforcement (modelo híbrido)

```
                 ┌─ internal ─────▶ se distribuye a todos
pack.access  ────┤
                 ├─ restricted ───▶ (1) bot: ¿owner del repo ∈ allowed_groups?
                 │                  (2) API pull: ¿claims del token ∈ allowed_groups?
                 └─ confidential ─▶ lo anterior + repo aislado (read por Team/Entra)
```

1. **Bot (push):** antes de abrir el PR a un repo, comprueba que el grupo dueño del repo está en `allowed_groups`. Si no, **no materializa** ese pack en ese repo.
   - Con team sync: `repo → GitHub Team → grupo Entra` es la fuente autoritativa.
   - Sin team sync: el `ai.manifest` declara `owner_group` y el bot lo verifica contra Entra (Graph) / un registro central de `repo → owner_group` mantenido por la plataforma.
2. **API de pull (on-demand / CLI / Action):** valida el token de Entra y comprueba `claims.groups ∩ allowed_groups`. Es el camino **más fuerte** y **no depende de team sync** → recomendado como vía principal para `restricted`/`confidential`.

## 6. Autenticación (sin secretos)

### En CI — OIDC federation (GitHub Actions ↔ Entra)
```
Job de GitHub Actions  (permissions: id-token: write)
   │ solicita un token OIDC de GitHub
   ▼
Token OIDC ── exchange ──▶ Entra ID (Workload Identity Federation)
   │                          valida el subject (org/repo/branch)
   ▼
Access token de Entra (con app role / grupos)
   │
   ▼
API de distribución: valida token + claims → sirve solo packs autorizados
```
**Cero secretos almacenados** ni credenciales que rotar.

### En local — MSAL device-code
La CLI (`CASKAi login`) abre login interactivo con Entra, obtiene un token con claims de grupo y lo usa para el pull.

### Detalle: "group overage"
Los tokens de Entra omiten el claim `groups` si el usuario pertenece a demasiados grupos (>200/150). Mitigación: usar **app roles** dedicados de la app de distribución (en vez de todos los grupos) **o** fallback a Microsoft Graph para resolver pertenencia. A documentar en la implementación.

## 7. Cadena de suministro (los packs inyectan instrucciones en la IA)

Un pack comprometido es un vector real (exfiltración, prompt-injection). Defensas:
- **Releases firmados** (cosign/sigstore o tags GPG firmados) en cada train.
- **Hashes de integridad** en `ai.lock` (la POC ya los genera) → el consumidor verifica que recibe exactamente lo publicado.
- **Bot = GitHub App de mínimo privilegio**: abre PRs, **no** mergea ni administra.
- **Gate 6 de CI**: lint de seguridad sobre `commands`/`skills` (secretos, comandos peligrosos, patrones de inyección).
- **Verificación en consumo**: la Action/CLI valida firma + hash antes de escribir los ficheros.

## 8. Auditoría

Trazabilidad de "quién consumió/modificó qué y cuándo":
- **Entra sign-in logs** → autenticaciones de pull.
- **Logs de la API de distribución** → qué pack/versión se sirvió a quién.
- **Inventario de `ai.lock`** (suma de todos los repos) → estado de adopción por versión.
- **Historial de PRs del monorepo** → cambios de autoría con revisor.

## 9. Componente nuevo: API de distribución

El enforcement de pull necesita un servicio ligero (validar token + servir packs autorizados). Encaja **Azure-native**: Azure Container Apps o Functions, detrás de Entra. Funciones:
- Validar token (OIDC/MSAL) y resolver grupos/app-roles.
- Servir packs según `access.allowed_groups`.
- Emitir logs de auditoría.

## 10. Dependencias y decisiones abiertas
- **Sistemas/IdP:** habilitar team sync (deseable) y registrar la app de distribución con federation + app roles.
- Estrategia de **group overage** (app roles vs Graph).
- Mapeo `repo → owner_group` para el bot sin team sync (registro central vs verificación Graph).
- Política de **firmas** (cosign vs GPG) y gestión de claves.
- Ventana y proceso de **rotación** de la app de distribución.

## 11. Fases de seguridad
| Fase | Alcance de seguridad |
|---|---|
| 0–1 | CODEOWNERS + branch protection; clasificación `internal` por defecto; hashes en `ai.lock`. |
| 2 | API de distribución + OIDC/MSAL; enforcement de `restricted` por grupos Entra. |
| 3 | `confidential` con repos aislados; releases firmados; auditoría centralizada. |
| 4 | Endurecimiento: rotación, alertas, revisión de group overage a escala. |

Ver también: [`architecture.md`](../governance/architecture.md) · [`operating-model.md`](./operating-model.md).
```
