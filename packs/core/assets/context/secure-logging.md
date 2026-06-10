---
id: secure-logging
type: context
description: "Reglas de logging seguro (sin datos sensibles)."
scope:
  applyTo: "**/*"
targets: [claude, copilot]
---
# Logging seguro

- Nunca registres secretos, tokens, PII ni credenciales.
- Usa logging estructurado con niveles correctos.
- Redacta cabeceras de autorización y cookies.

> Generado para **{{TARGET}}**.
