---
description: Comprueba qué packs puede consumir un consumidor según sus grupos de Entra
argument-hint: "<ruta-al-ai.manifest.yaml>"
---
Comprueba la visibilidad por rol del manifiesto **$ARGUMENTS**:

`caskai access --manifest $ARGUMENTS`

Resume qué packs quedan PERMITIDOS y DENEGADOS para sus `owner_groups`, explicando la clasificación
de cada pack y, en los denegados, qué grupo de Entra haría falta. Recuerda que la decisión queda
registrada en `governance/audit.log`.
