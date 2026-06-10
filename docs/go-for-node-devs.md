# Go para desarrolladores Node / TypeScript

> Guía de onboarding del equipo de `CASKAi`.
> Asume que vienes de **Node + TypeScript** y no has tocado Go. El objetivo es que en una tarde puedas leer, modificar y compilar el engine.

## Índice
1. [El cambio de mentalidad](#1-el-cambio-de-mentalidad)
2. [Toolchain: el equivalente a npm/node](#2-toolchain-el-equivalente-a-npmnode)
3. [Estructura de un proyecto Go](#3-estructura-de-un-proyecto-go)
4. [Sintaxis traducida desde TypeScript](#4-sintaxis-traducida-desde-typescript)
5. [Lo que más sorprende viniendo de TS](#5-lo-que-más-sorprende-viniendo-de-ts)
6. [Tooling de calidad (prettier/eslint/jest → ...)](#6-tooling-de-calidad)
7. [Ejemplo real de nuestro engine](#7-ejemplo-real-de-nuestro-engine)
8. [Cheat sheet TS → Go](#8-cheat-sheet-ts--go)
9. [Recursos para aprender](#9-recursos-para-aprender)

---

## 1. El cambio de mentalidad

| | Node/TypeScript | Go |
|---|---|---|
| Cómo se ejecuta | TS → se *transpila* a JS → lo corre el runtime de **Node** | Se **compila** a un **binario nativo** autónomo |
| Qué necesita la máquina destino | Node + `node_modules` | **Nada** (el binario lleva todo dentro) |
| Tipos | Se borran al compilar (solo dev-time) | Reales en tiempo de ejecución, parte del lenguaje |
| Concurrencia | event loop, `async/await` | *goroutines* + *channels* (hilos ligeros) |
| Filosofía | flexible, muchas formas de hacer algo | minimalista, "una forma obvia", opiniones fuertes |

La idea central: en TS, los tipos son una capa que desaparece; en Go, **el lenguaje ES tipado y compilado**, y el resultado es un ejecutable que copias y corre en cualquier sitio sin instalar Go.

---

## 2. Toolchain: el equivalente a npm/node

| Tarea | Node/TS | Go |
|---|---|---|
| Instalar el lenguaje | instalas Node | instalas Go (`brew install go` o instalador oficial) |
| Manifiesto de deps | `package.json` | `go.mod` |
| Lockfile | `package-lock.json` | `go.sum` |
| Carpeta de deps | `node_modules/` | caché global (no hay carpeta por proyecto) |
| Añadir una dependencia | `npm install lib` | `go get lib` |
| Limpiar/ordenar deps | — | `go mod tidy` |
| Ejecutar en dev | `ts-node x.ts` / `npm run dev` | `go run .` |
| Tests | `jest` | `go test ./...` (incluido en el lenguaje) |
| Build de producción | `tsc` / bundler | `go build` → **un binario** |
| Build para otro SO | (no aplica igual) | `GOOS=linux GOARCH=amd64 go build` |

No hay `node_modules` por proyecto: las dependencias se cachean globalmente y el binario final no las necesita. **No hay un `npx`**: o `go run` (dev) o el binario compilado.

---

## 3. Estructura de un proyecto Go

```
CASKAi/
├─ go.mod                 # como package.json: module + versión de Go + deps
├─ go.sum                 # como package-lock.json
├─ main.go                # punto de entrada (package main, func main)
└─ internal/
   ├─ canonical/          # un "package" = una carpeta
   │  └─ asset.go
   └─ adapter/
      ├─ claude.go
      └─ copilot.go
```

- **Un package = una carpeta.** Todos los `.go` de una carpeta comparten el mismo package (su "namespace").
- `package main` + `func main()` = el ejecutable (como el `"bin"` de tu package.json).
- `internal/` es una convención: lo que está dentro **no puede ser importado desde fuera del módulo** (encapsulación a nivel de repo).

`go.mod` mínimo:
```
module github.com/org/CASKAi
go 1.23
require gopkg.in/yaml.v3 v3.0.1
```

---

## 4. Sintaxis traducida desde TypeScript

### Variables
```ts
// TypeScript
let name = "core";
const version: string = "0.1.0";
```
```go
// Go
name := "core"          // := infiere el tipo (solo dentro de funciones)
var version string = "0.1.0"
const tier = "core"     // const para constantes
```
`:=` es el "let con inferencia". `var` cuando necesitas declarar sin asignar o a nivel de package.

### Funciones y el gran cambio: errores como valores
```ts
// TypeScript: lanzas excepciones
function parse(p: string): Pack {
  if (!ok) throw new Error("bad");
  return pack;
}
try { parse(x) } catch (e) { ... }
```
```go
// Go: NO hay try/catch. Las funciones devuelven (valor, error)
func parse(p string) (Pack, error) {
    if !ok {
        return Pack{}, fmt.Errorf("bad: %s", p)
    }
    return pack, nil   // nil = sin error
}

pack, err := parse(x)
if err != nil {        // este patrón lo verás MUCHÍSIMO
    return err
}
```
Esto es lo más distinto. En Go el manejo de errores es **explícito**: cada llamada que puede fallar devuelve un `error` y tú decides qué hacer. Verboso pero predecible — no hay errores "invisibles" que se propagan solos.

### Tipos: structs e interfaces (no hay clases)
```ts
// TypeScript
interface Pack { name: string; version: string }
class Builder {
  build(p: Pack): void { ... }
}
```
```go
// Go: structs para datos, métodos colgados de structs
type Pack struct {
    Name    string
    Version string
}

type Builder struct{}
func (b Builder) Build(p Pack) error { ... }   // método de Builder
```
No hay `class`. Defines `struct` (datos) y le "cuelgas" métodos con `func (b Builder) ...`. Composición en vez de herencia.

### Interfaces: ¡como el tipado estructural de TS!
```ts
// TS: structural typing — si tiene la forma, encaja
interface Emitter { emit(): string }
```
```go
// Go: las interfaces se satisfacen IMPLÍCITAMENTE (igual que TS)
type Emitter interface {
    Emit() string
}
// Cualquier tipo con un método Emit() string ES un Emitter,
// sin declararlo. No hay "implements".
```
Buena noticia: esto te va a resultar familiar. Go no usa `implements`; si tu tipo tiene los métodos, cumple la interfaz.

### Colecciones
```ts
const names: string[] = ["a", "b"];
names.push("c");
const m: Record<string, number> = { a: 1 };
```
```go
names := []string{"a", "b"}      // slice (array dinámico)
names = append(names, "c")        // no hay .push; usas append

m := map[string]int{"a": 1}       // map = Record/objeto
val, ok := m["a"]                 // "comma-ok": ok=false si no existe
```

### No hay null/undefined: "zero values" y `nil`
```ts
let x: string | undefined;   // undefined
let y: Foo | null = null;
```
```go
var s string   // "" (string vacío, NO null)
var n int      // 0
var b bool     // false
var p *Pack    // nil (los punteros y slices/maps sí pueden ser nil)
```
En Go cada tipo tiene un **valor cero** por defecto. No existe `undefined`. `nil` es solo para punteros, slices, maps, interfaces y funciones.

### Exportar = mayúscula inicial
```ts
export function build() {}     // export explícito
function helper() {}           // privado
```
```go
func Build() {}    // Mayúscula inicial = EXPORTADO (público)
func helper() {}   // minúscula = privado al package
```
No hay palabra `export`: **la mayúscula inicial decide la visibilidad**. Vale para funciones, tipos, campos de struct, todo.

### JSON/YAML: tags en vez de parse manual
```ts
const p = JSON.parse(raw) as Pack;
```
```go
type Pack struct {
    Name    string `yaml:"name"`     // "struct tags" mapean campos
    Version string `yaml:"version"`
}
var p Pack
yaml.Unmarshal(data, &p)             // &p = pasar puntero para que lo rellene
```

### `defer`: como un `finally` automático
```go
f, _ := os.Open("x")
defer f.Close()   // se ejecuta al salir de la función, pase lo que pase
```

### Concurrencia (lo verás poco en este proyecto)
```ts
await Promise.all([a(), b()]);
```
```go
go doWork()        // lanza una goroutine (hilo ligero)
ch := make(chan int)  // channel para comunicar entre goroutines
```
Nuestro engine es casi todo síncrono, así que esto es secundario por ahora.

---

## 5. Lo que más sorprende viniendo de TS

- **No hay try/catch.** Errores como valores (`if err != nil`). Al principio choca; luego se agradece la claridad.
- **Variables sin usar = error de compilación.** Go no compila si importas algo o declaras una variable que no usas. Es estricto a propósito.
- **Formato no negociable.** `gofmt` impone el estilo (¡usa tabs!). No hay debates de comas ni comillas: hay UNA forma.
- **API mínima.** La librería estándar y el lenguaje son pequeños. Menos magia, más explícito.
- **Sin `null` ni `undefined`.** Valores cero + `nil` controlado.
- **Punteros, pero suaves.** `&x` (dirección de) y `*x` (valor en). Sin aritmética de punteros peligrosa como en C.

---

## 6. Tooling de calidad

| Propósito | Node/TS | Go |
|---|---|---|
| Formateo | Prettier | **`gofmt` / `go fmt`** (estándar, sin config) |
| Linter | ESLint | `go vet` + **`golangci-lint`** (meta-linter) |
| Tests | Jest/Vitest | `go test` (incluido) |
| Type-check | `tsc` | el propio compilador `go build` |
| CLI framework | commander/oclif | **cobra** (el de `kubectl`, `gh`) |
| Manifiesto | package.json | go.mod |

Mucho de lo que en Node son dependencias separadas (formateo, test runner, type-check), en Go **viene en la caja**.

---

## 7. Ejemplo real de nuestro engine

Así se vería en Go la lógica que en la POC está en `build.py` (parsear un asset y resolver la degradación *fail-closed*):

```go
package canonical

import "fmt"

// Asset = un command/context/agent/skill en formato canónico.
type Asset struct {
    ID      string                  `yaml:"id"`
    Type    string                  `yaml:"type"`
    Targets map[string]TargetOpts   `yaml:"targets"`
}

type TargetOpts struct {
    Emit string `yaml:"emit"`
}

// nativeTypes: qué tipos soporta cada herramienta de forma nativa.
var nativeTypes = map[string]map[string]bool{
    "claude":  {"context": true, "command": true, "agent": true, "skill": true},
    "copilot": {"context": true, "command": true, "agent": true}, // NO skill
}

// ResolveEmit implementa el gate fail-closed.
// Devuelve la estrategia de emisión, o un error si no hay mapeo.
func ResolveEmit(a Asset, target string, degr map[string]Rule) (string, error) {
    if nativeTypes[target][a.Type] {
        return "native", nil
    }
    // No nativo: ¿override por-asset?
    if opts, ok := a.Targets[target]; ok && opts.Emit != "" {
        return opts.Emit, nil
    }
    // ¿Regla por tipo en degradation.yaml?
    if rule, ok := degr[fmt.Sprintf("%s->%s", a.Type, target)]; ok {
        return rule.Strategy, nil
    }
    // Fail-closed: rompe con error accionable.
    return "", fmt.Errorf(
        "degradación sin mapeo: asset %q (%s) apunta a %q, que no soporta ese tipo",
        a.ID, a.Type, target,
    )
}
```

Compáralo con el Python de `tools/build/build.py` (`resolve_emit`): la lógica es idéntica, solo cambia que en Go los errores son valores que devuelves (`return "", fmt.Errorf(...)`) en vez de `raise`.

---

## 8. Cheat sheet TS → Go

| TypeScript | Go |
|---|---|
| `let x = 1` | `x := 1` |
| `const X = 1` | `const X = 1` |
| `string`, `number`, `boolean` | `string`, `int`/`float64`, `bool` |
| `string[]` | `[]string` |
| `Record<string, T>` / objeto | `map[string]T` |
| `interface Foo { ... }` (forma de datos) | `type Foo struct { ... }` |
| `interface Foo { m(): void }` (contrato) | `type Foo interface { M() }` |
| `class C { method() {} }` | `type C struct{}` + `func (c C) Method() {}` |
| `function f(a: string): number` | `func f(a string) int` |
| múltiples retornos | `func f() (int, error)` |
| `throw` / `try/catch` | `return ..., err` / `if err != nil` |
| `null` / `undefined` | valor cero / `nil` |
| `JSON.parse` | `json.Unmarshal(data, &v)` |
| `export function` | `func Nombre` (mayúscula) |
| `import { x } from "y"` | `import "y"` y usas `y.X` |
| `async/await` | goroutines + channels |
| `arr.map(...)` | `for i, v := range arr { ... }` |
| Prettier | `gofmt` |
| ESLint | `golangci-lint` |

---

## 9. Recursos para aprender

- **[A Tour of Go](https://go.dev/tour/)** — interactivo, en el navegador. La mejor primera hora. **Empieza aquí.**
- **[Go by Example](https://gobyexample.com/)** — recetas cortas por tema; ideal para consultar.
- **[Effective Go](https://go.dev/doc/effective_go)** — el estilo y las convenciones idiomáticas.
- **[Standard library](https://pkg.go.dev/std)** — documentación de la librería estándar.
- Para CLIs: **[cobra](https://github.com/spf13/cobra)** (el framework que usaremos).

### Ruta sugerida para el equipo (≈ medio día)
1. *A Tour of Go* hasta "Methods and interfaces" (~1h).
2. Leer `tools/build/build.py` (la POC) y el ejemplo de la §7 en paralelo.
3. Hacer un cambio pequeño en el engine Go cuando exista y correr `go test ./...`.
```
