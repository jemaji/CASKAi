---
id: review-pr
type: command
title: "Revisar PR"
description: "Revisa el diff actual buscando bugs y problemas de seguridad."
argument_hint: "[número de PR opcional]"
targets: [claude, copilot]
---
Revisa el diff actual {{ARGS}} y reporta, en este orden:

1. **Bugs de correctitud** — lógica incorrecta, casos límite, posibles crashes.
2. **Problemas de seguridad** — inyección, secretos, validación de entrada.
3. **Simplificaciones** — duplicación, código muerto, oportunidades de reutilización.

Sé conciso y prioriza por impacto.
