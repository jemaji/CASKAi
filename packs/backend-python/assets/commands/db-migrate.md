---
id: db-migrate
type: command
title: "Generar migración de BD"
description: "Crea y valida una migración de base de datos para el servicio."
argument_hint: "[nombre de la migración]"
targets: [claude, copilot]
---
Genera una migración de base de datos {{ARGS}}:

1. Detecta los cambios de modelo pendientes.
2. Crea el script de migración (up/down) reversible.
3. Valida en una BD efímera antes de proponerla.
