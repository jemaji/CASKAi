// caskai — engine de CASKAi (Go, binario único sin dependencias).
//
// Subcomandos:
//   validate    corre los gates sobre los packs (CI)
//   build       compila canónico -> .claude/.github para un consumidor, con control de acceso por grupo
//   access      muestra qué packs puede consumir un rol/grupo (visibilidad por rol), audita la decisión
//   inventory   escanea los ai.lock de todos los consumidores -> trazabilidad 100% de uso
//   promote     mueve un asset a un pack (p. ej. promoción de dominio -> core)
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// version es sobreescrita en build release vía -ldflags "-X main.version=vX.Y.Z"
var version = "dev"

var native = map[string]map[string]bool{
	"claude":  {"context": true, "command": true, "agent": true, "skill": true},
	"copilot": {"context": true, "command": true, "agent": true}, // sin skill
}
var vars = map[string]map[string]string{
	"claude":  {"ARGS": "$ARGUMENTS", "TARGET": "Claude Code"},
	"copilot": {"ARGS": "${input:args}", "TARGET": "GitHub Copilot"},
}

// ---------- helpers genéricos ----------
func asMap(v any) map[string]any  { m, _ := v.(map[string]any); return m }
func asStr(v any) string          { s, _ := v.(string); return s }
func mget(m map[string]any, k string) any {
	if m == nil {
		return nil
	}
	return m[k]
}
func asList(v any) []any {
	if l, ok := v.([]any); ok {
		return l
	}
	if v == nil {
		return nil
	}
	return []any{v}
}
func strList(v any) []string {
	var o []string
	for _, e := range asList(v) {
		o = append(o, asStr(e))
	}
	return o
}
func intersects(a, b []string) bool {
	set := map[string]bool{}
	for _, x := range a {
		set[x] = true
	}
	for _, y := range b {
		if set[y] {
			return true
		}
	}
	return false
}
func loadYAML(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseYAML(string(b)), nil
}
func mustWrite(path, content string) {
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(strings.TrimRight(content, "\n")+"\n"), 0644)
}
func render(s, target string) string {
	for k, v := range vars[target] {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
}

// ---------- assets ----------
func parseAsset(path string) (map[string]any, string) {
	b, _ := os.ReadFile(path)
	s := string(b)
	if !strings.HasPrefix(s, "---") {
		return map[string]any{}, s
	}
	parts := strings.SplitN(s, "---", 3)
	meta := parseYAML(parts[1])
	return meta, strings.TrimLeft(parts[2], "\n")
}
func normalizeTargets(meta map[string]any) map[string]map[string]any {
	res := map[string]map[string]any{}
	switch t := meta["targets"].(type) {
	case []any:
		for _, e := range t {
			res[asStr(e)] = map[string]any{}
		}
	case map[string]any:
		for k, v := range t {
			res[k] = asMap(v)
		}
	}
	return res
}
func resolveEmit(meta map[string]any, target string, opts, degr map[string]any) (string, error) {
	atype := asStr(meta["type"])
	if native[target][atype] {
		return "native", nil
	}
	if e := asStr(mget(opts, "emit")); e != "" {
		return e, nil
	}
	if s := asStr(mget(asMap(mget(degr, atype+"->"+target)), "strategy")); s != "" {
		return s, nil
	}
	return "", fmt.Errorf("DEGRADACIÓN SIN MAPEO: asset %q (%s) -> %s sin mapeo declarado",
		asStr(meta["id"]), atype, target)
}

func listPacks(root string) []string {
	var out []string
	entries, _ := os.ReadDir(filepath.Join(root, "packs"))
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}
func assetFiles(packDir string) []string {
	var out []string
	filepath.WalkDir(filepath.Join(packDir, "assets"), func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".md") {
			out = append(out, p)
		}
		return nil
	})
	sort.Strings(out)
	return out
}
func packHash(packDir string) string {
	h := sha256.New()
	var files []string
	filepath.WalkDir(packDir, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			files = append(files, p)
		}
		return nil
	})
	sort.Strings(files)
	for _, f := range files {
		b, _ := os.ReadFile(f)
		h.Write(b)
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))[:16]
}

// ---------- audit ----------
func audit(root string, rec map[string]any) {
	rec["ts"] = time.Now().UTC().Format(time.RFC3339)
	b, _ := json.Marshal(rec)
	os.MkdirAll(filepath.Join(root, "governance"), 0755)
	f, err := os.OpenFile(filepath.Join(root, "governance", "audit.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		f.Write(append(b, '\n'))
	}
}

// ---------- access ----------
type decision struct {
	pack, classification, verdict string
	allowed                       []string
}

func decide(packMeta map[string]any, groups []string) decision {
	acc := asMap(mget(packMeta, "access"))
	cls := asStr(mget(acc, "classification"))
	if cls == "" {
		cls = "internal"
	}
	allowed := strList(mget(acc, "allowed_groups"))
	v := "DENEGADO"
	if cls == "internal" || intersects(groups, allowed) {
		v = "PERMITIDO"
	}
	return decision{asStr(packMeta["name"]), cls, v, allowed}
}

// ---------- adapters ----------
func emitClaude(meta map[string]any, body, assetDir, out string, ctx *[]string) {
	atype := asStr(meta["type"])
	body = render(body, "claude")
	id := asStr(meta["id"])
	switch atype {
	case "context":
		*ctx = append(*ctx, "<!-- "+id+" -->\n"+body)
	case "command":
		fm := "---\ndescription: " + asStr(meta["description"]) + "\n"
		if ah := asStr(meta["argument_hint"]); ah != "" {
			fm += "argument-hint: " + ah + "\n"
		}
		fm += "---\n"
		mustWrite(filepath.Join(out, ".claude/commands", id+".md"), fm+body)
	case "agent":
		fm := "---\ndescription: " + asStr(meta["description"]) + "\n---\n"
		mustWrite(filepath.Join(out, ".claude/agents", id+".md"), fm+body)
	case "skill":
		name := asStr(meta["name"])
		if name == "" {
			name = id
		}
		fm := "---\nname: " + name + "\ndescription: " + asStr(meta["description"]) + "\n---\n"
		sdir := filepath.Join(out, ".claude/skills", id)
		mustWrite(filepath.Join(sdir, "SKILL.md"), fm+body)
		for _, r := range strList(meta["resources"]) {
			copyTree(filepath.Join(assetDir, r), filepath.Join(sdir, r))
		}
	}
}
func emitCopilot(meta map[string]any, body, out, emit string, ctx *[]string) {
	atype := asStr(meta["type"])
	body = render(body, "copilot")
	id := asStr(meta["id"])
	eff := atype
	if atype == "skill" && emit == "prompt" {
		eff = "command"
	}
	switch {
	case atype == "context":
		if ap := asStr(mget(asMap(meta["scope"]), "applyTo")); ap != "" {
			mustWrite(filepath.Join(out, ".github/instructions", id+".instructions.md"),
				"---\napplyTo: '"+ap+"'\n---\n"+body)
		} else {
			*ctx = append(*ctx, body)
		}
	case eff == "command":
		note := ""
		if atype == "skill" {
			note = "<!-- degradado desde skill; recursos no disponibles en Copilot -->\n"
		}
		mustWrite(filepath.Join(out, ".github/prompts", id+".prompt.md"),
			"---\nmode: agent\ndescription: "+asStr(meta["description"])+"\n---\n"+note+body)
	case atype == "agent":
		mustWrite(filepath.Join(out, ".github/chatmodes", id+".chatmode.md"),
			"---\ndescription: "+asStr(meta["description"])+"\n---\n"+body)
	}
}
func copyTree(src, dst string) {
	filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			os.MkdirAll(target, 0755)
		} else {
			b, _ := os.ReadFile(p)
			os.MkdirAll(filepath.Dir(target), 0755)
			os.WriteFile(target, b, 0644)
		}
		return nil
	})
}

// ---------- comandos ----------
func cmdValidate(root string) int {
	degr, _ := loadYAML(filepath.Join(root, "governance", "degradation.yaml"))
	fail := false
	fmt.Println("== gate: validación de packs ==")
	for _, pk := range listPacks(root) {
		pdir := filepath.Join(root, "packs", pk)
		meta, err := loadYAML(filepath.Join(pdir, "pack.yaml"))
		if err != nil {
			fmt.Printf("  ✗ %s: falta pack.yaml\n", pk)
			fail = true
			continue
		}
		acc := asMap(mget(meta, "access"))
		cls := asStr(mget(acc, "classification"))
		if cls == "" {
			cls = "internal"
		}
		fmt.Printf("  pack %s@%s (tier=%s, %s)\n", asStr(meta["name"]), asStr(meta["version"]), asStr(meta["tier"]), cls)
		if len(strList(meta["owners"])) == 0 {
			fmt.Printf("      ✗ schema: faltan owners\n")
			fail = true
		}
		for _, af := range assetFiles(pdir) {
			am, _ := parseAsset(af)
			if asStr(am["id"]) == "" || asStr(am["type"]) == "" {
				fmt.Printf("      ✗ %s: falta id/type\n", filepath.Base(af))
				fail = true
				continue
			}
			for tg, opts := range normalizeTargets(am) {
				if _, err := resolveEmit(am, tg, opts, degr); err != nil {
					fmt.Printf("      ✗ %s\n", err)
					fail = true
				}
			}
		}
	}
	if fail {
		fmt.Println("\n❌ validación FALLÓ")
		return 1
	}
	fmt.Println("\n✅ validación OK")
	return 0
}

func consumerName(manifestPath string) string {
	return filepath.Base(filepath.Dir(manifestPath))
}

func cmdBuild(root, manifestPath, out string) int {
	man, err := loadYAML(manifestPath)
	if err != nil {
		fmt.Println("error leyendo manifiesto:", err)
		return 1
	}
	degr, _ := loadYAML(filepath.Join(root, "governance", "degradation.yaml"))
	groups := strList(man["owner_groups"])
	who := consumerName(manifestPath)
	// limpia solo lo generado (nunca el repo del consumidor entero)
	for _, g := range []string{".claude", ".github", "CLAUDE.md", "ai.lock"} {
		os.RemoveAll(filepath.Join(out, g))
	}
	var claudeCtx, copilotCtx []string
	lockPacks := map[string]string{}
	lockHash := map[string]string{}
	fmt.Printf("== build de %q  grupos=%v ==\n", who, groups)
	for _, pk := range strList(man["packs"]) {
		pdir := filepath.Join(root, "packs", pk)
		meta, err := loadYAML(filepath.Join(pdir, "pack.yaml"))
		if err != nil {
			fmt.Printf("  ✗ pack %s no existe\n", pk)
			continue
		}
		d := decide(meta, groups)
		audit(root, map[string]any{"actor": who, "action": "build", "pack": pk,
			"classification": d.classification, "decision": d.verdict, "groups": groups})
		if d.verdict == "DENEGADO" {
			fmt.Printf("  🔒 %s [%s] DENEGADO (requiere %v) — no se materializa\n", pk, d.classification, d.allowed)
			continue
		}
		fmt.Printf("  ✓ %s@%s [%s] PERMITIDO\n", pk, asStr(meta["version"]), d.classification)
		for _, af := range assetFiles(pdir) {
			am, body := parseAsset(af)
			adir := filepath.Dir(af)
			for tg, opts := range normalizeTargets(am) {
				emit, err := resolveEmit(am, tg, opts, degr)
				if err != nil {
					fmt.Printf("      ✗ %s\n", err)
					return 1
				}
				if tg == "claude" {
					emitClaude(am, body, adir, out, &claudeCtx)
				} else if tg == "copilot" {
					emitCopilot(am, body, out, emit, &copilotCtx)
				}
			}
		}
		lockPacks[pk] = asStr(meta["version"])
		lockHash[pk] = packHash(pdir)
	}
	if len(claudeCtx) > 0 {
		mustWrite(filepath.Join(out, "CLAUDE.md"), "# Contexto (CASKAi)\n\n"+strings.Join(claudeCtx, "\n\n"))
	}
	if len(copilotCtx) > 0 {
		mustWrite(filepath.Join(out, ".github/copilot-instructions.md"), strings.Join(copilotCtx, "\n\n"))
	}
	mustWrite(out+"/ai.lock", renderLock(asStr(man["channel"]), groups, lockPacks, lockHash))
	fmt.Printf("  → ai.lock + artefactos en %s\n", out)
	return 0
}

func renderLock(channel string, groups []string, packs, hashes map[string]string) string {
	var b strings.Builder
	b.WriteString("channel: " + channel + "\n")
	b.WriteString("groups: [" + strings.Join(groups, ", ") + "]\n")
	b.WriteString("packs:\n")
	keys := sortedKeys(packs)
	for _, k := range keys {
		b.WriteString("  " + k + ": \"" + packs[k] + "\"\n")
	}
	b.WriteString("integrity:\n")
	for _, k := range keys {
		b.WriteString("  " + k + ": \"" + hashes[k] + "\"\n")
	}
	return b.String()
}
func sortedKeys(m map[string]string) []string {
	var k []string
	for x := range m {
		k = append(k, x)
	}
	sort.Strings(k)
	return k
}

func cmdAccess(root, manifestPath string) int {
	man, _ := loadYAML(manifestPath)
	groups := strList(man["owner_groups"])
	who := consumerName(manifestPath)
	fmt.Printf("== visibilidad por rol — consumidor %q, grupos=%v ==\n", who, groups)
	fmt.Printf("  %-18s %-14s %s\n", "PACK", "CLASIFICACIÓN", "DECISIÓN")
	for _, pk := range listPacks(root) {
		meta, _ := loadYAML(filepath.Join(root, "packs", pk, "pack.yaml"))
		d := decide(meta, groups)
		mark := "✓"
		if d.verdict == "DENEGADO" {
			mark = "🔒"
		}
		fmt.Printf("  %s %-16s %-14s %s\n", mark, pk, d.classification, d.verdict)
		audit(root, map[string]any{"actor": who, "action": "access-check", "pack": pk,
			"classification": d.classification, "decision": d.verdict, "groups": groups})
	}
	return 0
}

func cmdInventory(root, consumersDir string) int {
	// pack -> version -> []consumer
	agg := map[string]map[string][]string{}
	total := 0
	entries, _ := os.ReadDir(consumersDir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		lockPath := filepath.Join(consumersDir, e.Name(), "ai.lock")
		lock, err := loadYAML(lockPath)
		if err != nil {
			continue
		}
		total++
		for pk, ver := range asMap(mget(lock, "packs")) {
			if agg[pk] == nil {
				agg[pk] = map[string][]string{}
			}
			v := asStr(ver)
			agg[pk][v] = append(agg[pk][v], e.Name())
		}
	}
	fmt.Printf("== inventario de adopción (trazabilidad) — %d consumidores ==\n", total)
	fmt.Printf("  %-18s %-10s %-7s %s\n", "PACK", "VERSIÓN", "USOS", "CONSUMIDORES")
	var packs []string
	for pk := range agg {
		packs = append(packs, pk)
	}
	sort.Strings(packs)
	for _, pk := range packs {
		var vers []string
		for v := range agg[pk] {
			vers = append(vers, v)
		}
		sort.Strings(vers)
		for _, v := range vers {
			cons := agg[pk][v]
			sort.Strings(cons)
			fmt.Printf("  %-18s %-10s %-7d %s\n", pk, v, len(cons), strings.Join(cons, ", "))
		}
		if len(vers) > 1 {
			fmt.Printf("    ⚠ deriva de versión en %q: %v conviven\n", pk, vers)
		}
	}
	return 0
}

func cmdPromote(root, assetRel, toPack string) int {
	// assetRel: "<pack>/assets/<sub>"  ->  packs/<toPack>/assets/<sub>
	src := filepath.Join(root, "packs", assetRel)
	i := strings.Index(assetRel, "assets/")
	if i < 0 {
		fmt.Println("ruta de asset inválida")
		return 1
	}
	sub := assetRel[i:]
	dst := filepath.Join(root, "packs", toPack, sub)
	os.MkdirAll(filepath.Dir(dst), 0755)
	if err := os.Rename(src, dst); err != nil {
		fmt.Println("error moviendo:", err)
		return 1
	}
	fmt.Printf("promocionado: %s  →  packs/%s/%s\n", assetRel, toPack, sub)
	audit(root, map[string]any{"actor": "ai-governance", "action": "promote",
		"pack": toPack, "from": assetRel, "decision": "PROMOVIDO"})
	return 0
}

func main() {
	root := "."
	args := os.Args[1:]
	// flag --root
	var rest []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--root" && i+1 < len(args) {
			root = args[i+1]
			i++
		} else {
			rest = append(rest, args[i])
		}
	}
	if len(rest) == 0 {
		fmt.Printf("caskai %s\n", version)
		fmt.Println("uso: caskai <validate|build|access|inventory|promote|version> [opts]")
		os.Exit(2)
	}
	flag := func(name, def string) string {
		for i := 0; i < len(rest)-1; i++ {
			if rest[i] == name {
				return rest[i+1]
			}
		}
		return def
	}
	switch rest[0] {
	case "validate":
		os.Exit(cmdValidate(root))
	case "build":
		os.Exit(cmdBuild(root, flag("--manifest", ""), flag("--out", "dist/out")))
	case "access":
		os.Exit(cmdAccess(root, flag("--manifest", "")))
	case "inventory":
		os.Exit(cmdInventory(root, flag("--consumers", filepath.Join(root, "consumers"))))
	case "promote":
		os.Exit(cmdPromote(root, flag("--asset", ""), flag("--to", "core")))
	case "version":
		fmt.Printf("caskai %s\n", version)
	default:
		fmt.Println("comando desconocido:", rest[0])
		os.Exit(2)
	}
}
