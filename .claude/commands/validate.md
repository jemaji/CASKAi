---
description: Compila el engine y corre los gates de validación sobre todos los packs
---
Compila y valida CASKAi:

1. `go build -o bin/caskai ./tools/caskai`
2. `./bin/caskai validate`

Resume el resultado: packs detectados (con tier y clasificación), y cualquier fallo de schema o
de degradación fail-closed. Si algo falla, propón la corrección concreta.
