---
description: Muestra la trazabilidad de adopción (quién usa qué pack y versión) y detecta deriva
argument-hint: "<directorio-con-caskai.lock>"
---
Ejecuta `caskai inventory --consumers $ARGUMENTS` y resume:

- Qué packs y versiones están en uso y por qué consumidores.
- **Deriva de versión** (mismo pack en varias versiones) → candidatos a PR de actualización por el bot.
- Packs sin consumidores (posibles candidatos a deprecación).

Si hay deriva relevante, sugiere los repos que deberían actualizarse.
