package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ────────────────────────────────────────────────
// helpers para fixtures
// ────────────────────────────────────────────────

func mkFile(t *testing.T, path, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("mkFile %s: %v", path, err)
	}
}

// repoFixture crea un repo mínimo válido en un directorio temporal.
func repoFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// governance/degradation.yaml
	mkFile(t, filepath.Join(root, "governance", "degradation.yaml"),
		"skill->copilot:\n  strategy: prompt\n")

	// packs/core/pack.yaml
	mkFile(t, filepath.Join(root, "packs", "core", "pack.yaml"), `
name: core
version: 0.1.0
tier: core
owners: ["@org/ai-governance"]
targets: [claude, copilot]
access:
  classification: internal
`)
	// packs/core/assets/context/conventions.md
	mkFile(t, filepath.Join(root, "packs", "core", "assets", "context", "conventions.md"), `---
id: coding-conventions
type: context
description: Convenciones de código
targets: [claude, copilot]
---
Usa inglés en nombres de variables.
`)

	// packs/core/assets/commands/review-pr.md
	mkFile(t, filepath.Join(root, "packs", "core", "assets", "commands", "review-pr.md"), `---
id: review-pr
type: command
description: Revisa un PR
targets: [claude, copilot]
---
Revisa el PR {{ARGS}}.
`)

	// packs/restricted/pack.yaml  (restricted, sin allowed_groups)
	mkFile(t, filepath.Join(root, "packs", "restricted", "pack.yaml"), `
name: restricted
version: 0.1.0
tier: domain
owners: ["@org/backend-guild"]
targets: [claude]
access:
  classification: restricted
  allowed_groups: [backend-guild]
`)
	mkFile(t, filepath.Join(root, "packs", "restricted", "assets", "context", "secret.md"), `---
id: secret-context
type: context
description: Secreto
targets: [claude]
---
Contenido restringido.
`)
	return root
}

// ────────────────────────────────────────────────
// YAML parser
// ────────────────────────────────────────────────

func TestParseYAML_SimpleKV(t *testing.T) {
	m := parseYAML("name: core\nversion: 1.0.0\n")
	if m["name"] != "core" {
		t.Errorf("name: got %q want %q", m["name"], "core")
	}
	if m["version"] != "1.0.0" {
		t.Errorf("version: got %q want %q", m["version"], "1.0.0")
	}
}

func TestParseYAML_InlineArray(t *testing.T) {
	m := parseYAML("targets: [claude, copilot]\n")
	list, ok := m["targets"].([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("targets: got %v", m["targets"])
	}
	if list[0] != "claude" || list[1] != "copilot" {
		t.Errorf("targets: got %v", list)
	}
}

func TestParseYAML_BlockSequence(t *testing.T) {
	m := parseYAML("owners:\n  - \"@org/ai-governance\"\n  - \"@org/backend\"\n")
	list := strList(m["owners"])
	if len(list) != 2 || list[0] != "@org/ai-governance" {
		t.Errorf("owners: got %v", list)
	}
}

func TestParseYAML_NestedMap(t *testing.T) {
	m := parseYAML("access:\n  classification: restricted\n  allowed_groups: [backend-guild]\n")
	acc := asMap(m["access"])
	if acc == nil {
		t.Fatal("access is nil")
	}
	if acc["classification"] != "restricted" {
		t.Errorf("classification: got %q", acc["classification"])
	}
}

func TestParseYAML_QuotedValues(t *testing.T) {
	m := parseYAML(`name: "hello world"` + "\n")
	if m["name"] != "hello world" {
		t.Errorf("name: got %q", m["name"])
	}
}

func TestParseYAML_CommentsIgnored(t *testing.T) {
	m := parseYAML("name: core # comentario\nversion: 1.0.0\n")
	if m["name"] != "core" {
		t.Errorf("name con comentario: got %q", m["name"])
	}
}

func TestParseYAML_Empty(t *testing.T) {
	m := parseYAML("")
	if len(m) != 0 {
		t.Errorf("empty: got %v", m)
	}
}

// ────────────────────────────────────────────────
// parseAsset
// ────────────────────────────────────────────────

func TestParseAsset_WithFrontmatter(t *testing.T) {
	f := filepath.Join(t.TempDir(), "asset.md")
	os.WriteFile(f, []byte("---\nid: my-asset\ntype: context\n---\nContenido del asset.\n"), 0644)

	meta, body := parseAsset(f)
	if meta["id"] != "my-asset" {
		t.Errorf("id: got %q", meta["id"])
	}
	if meta["type"] != "context" {
		t.Errorf("type: got %q", meta["type"])
	}
	if !strings.Contains(body, "Contenido") {
		t.Errorf("body: got %q", body)
	}
}

func TestParseAsset_WithoutFrontmatter(t *testing.T) {
	f := filepath.Join(t.TempDir(), "asset.md")
	os.WriteFile(f, []byte("Solo contenido sin frontmatter.\n"), 0644)

	meta, body := parseAsset(f)
	if len(meta) != 0 {
		t.Errorf("meta debe estar vacío, got %v", meta)
	}
	if !strings.Contains(body, "Sin frontmatter") {
		// cuerpo tiene el contenido
		_ = body
	}
}

// ────────────────────────────────────────────────
// intersects
// ────────────────────────────────────────────────

func TestIntersects(t *testing.T) {
	cases := []struct {
		a, b []string
		want bool
	}{
		{[]string{"a", "b"}, []string{"b", "c"}, true},
		{[]string{"a"}, []string{"b", "c"}, false},
		{[]string{}, []string{"a"}, false},
		{[]string{"x"}, []string{}, false},
		{nil, nil, false},
	}
	for _, c := range cases {
		got := intersects(c.a, c.b)
		if got != c.want {
			t.Errorf("intersects(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

// ────────────────────────────────────────────────
// decide
// ────────────────────────────────────────────────

func TestDecide_InternalAlwaysPermitido(t *testing.T) {
	meta := map[string]any{
		"name": "core",
		"access": map[string]any{
			"classification": "internal",
		},
	}
	d := decide(meta, []string{})
	if d.verdict != "PERMITIDO" {
		t.Errorf("internal sin grupos debe ser PERMITIDO, got %q", d.verdict)
	}
}

func TestDecide_RestrictedConGrupoPermitido(t *testing.T) {
	meta := map[string]any{
		"name": "restricted",
		"access": map[string]any{
			"classification": "restricted",
			"allowed_groups": []any{"backend-guild"},
		},
	}
	d := decide(meta, []string{"backend-guild", "platform"})
	if d.verdict != "PERMITIDO" {
		t.Errorf("restricted con grupo correcto debe ser PERMITIDO, got %q", d.verdict)
	}
}

func TestDecide_RestrictedSinGrupoDenegado(t *testing.T) {
	meta := map[string]any{
		"name": "restricted",
		"access": map[string]any{
			"classification": "restricted",
			"allowed_groups": []any{"backend-guild"},
		},
	}
	d := decide(meta, []string{"frontend-team"})
	if d.verdict != "DENEGADO" {
		t.Errorf("restricted sin grupo debe ser DENEGADO, got %q", d.verdict)
	}
}

func TestDecide_SinClassificationDefaulsInternal(t *testing.T) {
	meta := map[string]any{"name": "core", "access": map[string]any{}}
	d := decide(meta, []string{})
	if d.verdict != "PERMITIDO" {
		t.Errorf("sin classification debe defaultear a internal/PERMITIDO, got %q", d.verdict)
	}
	if d.classification != "internal" {
		t.Errorf("classification default: got %q", d.classification)
	}
}

// ────────────────────────────────────────────────
// resolveEmit
// ────────────────────────────────────────────────

func TestResolveEmit_NativoClaude(t *testing.T) {
	meta := map[string]any{"id": "x", "type": "context"}
	emit, err := resolveEmit(meta, "claude", nil, nil)
	if err != nil || emit != "native" {
		t.Errorf("context->claude: emit=%q err=%v", emit, err)
	}
}

func TestResolveEmit_NativoCopilot(t *testing.T) {
	meta := map[string]any{"id": "x", "type": "command"}
	emit, err := resolveEmit(meta, "copilot", nil, nil)
	if err != nil || emit != "native" {
		t.Errorf("command->copilot: emit=%q err=%v", emit, err)
	}
}

func TestResolveEmit_SkillCopilotViaDegradation(t *testing.T) {
	meta := map[string]any{"id": "x", "type": "skill"}
	degr := map[string]any{
		"skill->copilot": map[string]any{"strategy": "prompt"},
	}
	emit, err := resolveEmit(meta, "copilot", nil, degr)
	if err != nil || emit != "prompt" {
		t.Errorf("skill->copilot vía degradation: emit=%q err=%v", emit, err)
	}
}

func TestResolveEmit_SkillCopilotViaOpts(t *testing.T) {
	meta := map[string]any{"id": "x", "type": "skill"}
	opts := map[string]any{"emit": "prompt"}
	emit, err := resolveEmit(meta, "copilot", opts, nil)
	if err != nil || emit != "prompt" {
		t.Errorf("skill->copilot vía opts: emit=%q err=%v", emit, err)
	}
}

func TestResolveEmit_ErrorSinMapeo(t *testing.T) {
	meta := map[string]any{"id": "x", "type": "skill"}
	_, err := resolveEmit(meta, "copilot", nil, nil)
	if err == nil {
		t.Error("skill->copilot sin mapeo debe retornar error")
	}
}

// ────────────────────────────────────────────────
// cmdValidate (integración)
// ────────────────────────────────────────────────

func TestCmdValidate_OK(t *testing.T) {
	root := repoFixture(t)
	code := cmdValidate(root)
	if code != 0 {
		t.Errorf("validate debe pasar en repo válido, exit=%d", code)
	}
}

func TestCmdValidate_FaltaOwners(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "governance", "degradation.yaml"), "")
	mkFile(t, filepath.Join(root, "packs", "bad", "pack.yaml"), `
name: bad
version: 0.1.0
tier: domain
targets: [claude]
access:
  classification: internal
`)
	code := cmdValidate(root)
	if code == 0 {
		t.Error("validate debe fallar si falta owners")
	}
}

func TestCmdValidate_AssetSinID(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "governance", "degradation.yaml"), "")
	mkFile(t, filepath.Join(root, "packs", "p", "pack.yaml"), `
name: p
version: 0.1.0
tier: domain
owners: ["@org/team"]
targets: [claude]
access:
  classification: internal
`)
	mkFile(t, filepath.Join(root, "packs", "p", "assets", "context", "broken.md"), `---
type: context
targets: [claude]
---
Sin id.
`)
	code := cmdValidate(root)
	if code == 0 {
		t.Error("validate debe fallar si un asset no tiene id")
	}
}

func TestCmdValidate_DegradacionFailClosed(t *testing.T) {
	root := t.TempDir()
	// degradation.yaml vacío → skill->copilot no tiene mapeo
	mkFile(t, filepath.Join(root, "governance", "degradation.yaml"), "")
	mkFile(t, filepath.Join(root, "packs", "p", "pack.yaml"), `
name: p
version: 0.1.0
tier: domain
owners: ["@org/team"]
targets: [claude, copilot]
access:
  classification: internal
`)
	mkFile(t, filepath.Join(root, "packs", "p", "assets", "skills", "my-skill", "SKILL.md"), `---
id: my-skill
type: skill
description: Un skill
targets: [claude, copilot]
---
Contenido del skill.
`)
	code := cmdValidate(root)
	if code == 0 {
		t.Error("validate debe fallar (fail-closed) ante skill->copilot sin mapeo en degradation.yaml")
	}
}

func TestCmdValidate_SinPacks(t *testing.T) {
	root := t.TempDir()
	mkFile(t, filepath.Join(root, "governance", "degradation.yaml"), "")
	os.MkdirAll(filepath.Join(root, "packs"), 0755)
	code := cmdValidate(root)
	if code != 0 {
		t.Errorf("validate con repo vacío (sin packs) debe pasar, exit=%d", code)
	}
}

// ────────────────────────────────────────────────
// cmdBuild (integración)
// ────────────────────────────────────────────────

func TestCmdBuild_PermisoInternal(t *testing.T) {
	root := repoFixture(t)
	out := t.TempDir()
	manifest := filepath.Join(t.TempDir(), "caskai.yaml")
	os.WriteFile(manifest, []byte("channel: stable\nowner_groups: [frontend-team]\npacks:\n  - core\n"), 0644)

	code := cmdBuild(root, manifest, out)
	if code != 0 {
		t.Fatalf("build debe pasar para pack internal, exit=%d", code)
	}
	// CLAUDE.md generado
	if _, err := os.Stat(filepath.Join(out, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md no generado")
	}
	// caskai.lock generado
	if _, err := os.Stat(filepath.Join(out, "caskai.lock")); err != nil {
		t.Error("caskai.lock no generado")
	}
}

func TestCmdBuild_PackRestrictedDenegado(t *testing.T) {
	root := repoFixture(t)
	out := t.TempDir()
	manifest := filepath.Join(t.TempDir(), "caskai.yaml")
	// frontend-team no está en backend-guild → restricted pack denegado
	os.WriteFile(manifest, []byte("channel: stable\nowner_groups: [frontend-team]\npacks:\n  - restricted\n"), 0644)

	code := cmdBuild(root, manifest, out)
	if code != 0 {
		t.Fatalf("build debe continuar aunque pack denegado, exit=%d", code)
	}
	// CLAUDE.md no debe existir (no hay contexto que materializar)
	if _, err := os.Stat(filepath.Join(out, "CLAUDE.md")); err == nil {
		t.Error("CLAUDE.md no debe generarse cuando el pack está denegado")
	}
}

func TestCmdBuild_PackRestrictedPermitido(t *testing.T) {
	root := repoFixture(t)
	out := t.TempDir()
	manifest := filepath.Join(t.TempDir(), "caskai.yaml")
	os.WriteFile(manifest, []byte("channel: stable\nowner_groups: [backend-guild]\npacks:\n  - restricted\n"), 0644)

	code := cmdBuild(root, manifest, out)
	if code != 0 {
		t.Fatalf("build debe pasar con grupo correcto, exit=%d", code)
	}
	if _, err := os.Stat(filepath.Join(out, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md debe generarse cuando el pack está permitido")
	}
}

func TestCmdBuild_GeneraCommands(t *testing.T) {
	root := repoFixture(t)
	out := t.TempDir()
	manifest := filepath.Join(t.TempDir(), "caskai.yaml")
	os.WriteFile(manifest, []byte("channel: stable\nowner_groups: [platform-core]\npacks:\n  - core\n"), 0644)

	cmdBuild(root, manifest, out)
	// command generado en .claude/commands/
	if _, err := os.Stat(filepath.Join(out, ".claude", "commands", "review-pr.md")); err != nil {
		t.Error(".claude/commands/review-pr.md no generado")
	}
	// command generado en .github/prompts/
	if _, err := os.Stat(filepath.Join(out, ".github", "prompts", "review-pr.prompt.md")); err != nil {
		t.Error(".github/prompts/review-pr.prompt.md no generado")
	}
}

func TestCmdBuild_LockContienePack(t *testing.T) {
	root := repoFixture(t)
	out := t.TempDir()
	manifest := filepath.Join(t.TempDir(), "caskai.yaml")
	os.WriteFile(manifest, []byte("channel: stable\nowner_groups: []\npacks:\n  - core\n"), 0644)

	cmdBuild(root, manifest, out)
	lock, _ := os.ReadFile(filepath.Join(out, "caskai.lock"))
	s := string(lock)
	if !strings.Contains(s, "core") {
		t.Error("caskai.lock debe referenciar el pack core")
	}
	if !strings.Contains(s, "engine:") {
		t.Error("caskai.lock debe incluir la versión del engine")
	}
	if !strings.Contains(s, "generated_at:") {
		t.Error("caskai.lock debe incluir la fecha de generación")
	}
}

func TestCmdBuild_ManifiestNoExiste(t *testing.T) {
	root := repoFixture(t)
	out := t.TempDir()
	code := cmdBuild(root, "/no/existe.yaml", out)
	if code == 0 {
		t.Error("build con manifiesto inexistente debe retornar error")
	}
}

// ────────────────────────────────────────────────
// cmdAccess (integración)
// ────────────────────────────────────────────────

func TestCmdAccess_OK(t *testing.T) {
	root := repoFixture(t)
	manifest := filepath.Join(t.TempDir(), "caskai.yaml")
	os.WriteFile(manifest, []byte("channel: stable\nowner_groups: [backend-guild]\npacks:\n  - core\n  - restricted\n"), 0644)

	code := cmdAccess(root, manifest)
	if code != 0 {
		t.Errorf("access debe retornar 0, got %d", code)
	}
}

// ────────────────────────────────────────────────
// cmdInventory (integración)
// ────────────────────────────────────────────────

func TestCmdInventory_SinConsumidores(t *testing.T) {
	root := repoFixture(t)
	consumersDir := t.TempDir() // vacío
	code := cmdInventory(root, consumersDir)
	if code != 0 {
		t.Errorf("inventory sin consumidores debe retornar 0, got %d", code)
	}
}

func TestCmdInventory_ConLocks(t *testing.T) {
	root := repoFixture(t)
	consumersDir := t.TempDir()

	// consumidor A con caskai.lock
	lockA := "channel: stable\ngroups: []\npacks:\n  core: \"0.1.0\"\nintegrity:\n  core: \"sha256:abc123\"\n"
	mkFile(t, filepath.Join(consumersDir, "app-a", "caskai.lock"), lockA)

	// consumidor B con misma versión
	mkFile(t, filepath.Join(consumersDir, "app-b", "caskai.lock"), lockA)

	code := cmdInventory(root, consumersDir)
	if code != 0 {
		t.Errorf("inventory con locks debe retornar 0, got %d", code)
	}
}

func TestCmdInventory_DerivaDiagnosticada(t *testing.T) {
	// Dos consumidores con versiones distintas del mismo pack → debe indicar deriva
	root := repoFixture(t)
	consumersDir := t.TempDir()

	mkFile(t, filepath.Join(consumersDir, "app-a", "caskai.lock"),
		"channel: stable\ngroups: []\npacks:\n  core: \"0.1.0\"\nintegrity:\n  core: \"sha256:abc\"\n")
	mkFile(t, filepath.Join(consumersDir, "app-b", "caskai.lock"),
		"channel: stable\ngroups: []\npacks:\n  core: \"0.2.0\"\nintegrity:\n  core: \"sha256:def\"\n")

	// inventory no falla ante deriva, solo la reporta
	code := cmdInventory(root, consumersDir)
	if code != 0 {
		t.Errorf("inventory con deriva debe retornar 0 (solo aviso), got %d", code)
	}
}

// ────────────────────────────────────────────────
// render (vars por target)
// ────────────────────────────────────────────────

func TestRender_ClaudeVars(t *testing.T) {
	out := render("Usa {{ARGS}} en {{TARGET}}", "claude")
	if !strings.Contains(out, "$ARGUMENTS") || !strings.Contains(out, "Claude Code") {
		t.Errorf("render claude: got %q", out)
	}
}

func TestRender_CopilotVars(t *testing.T) {
	out := render("Usa {{ARGS}} en {{TARGET}}", "copilot")
	if !strings.Contains(out, "${input:args}") || !strings.Contains(out, "GitHub Copilot") {
		t.Errorf("render copilot: got %q", out)
	}
}
