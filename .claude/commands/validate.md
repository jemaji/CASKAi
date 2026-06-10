---
description: Corre los gates de validación sobre todos los packs
---
Valida CASKAi con el engine instalado:

1. Comprueba que `caskai` está disponible: `caskai version`
   - Si no está instalado: `curl -fsSL https://raw.githubusercontent.com/jemaji/CASKAi/main/install.sh | bash`
2. `caskai validate`

Resume el resultado: packs detectados (con tier y clasificación), y cualquier fallo de schema o
de degradación fail-closed. Si algo falla, propón la corrección concreta.
