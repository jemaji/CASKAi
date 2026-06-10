package main

import "strings"

// Mini-parser YAML (subconjunto): mapas anidados por indentación, secuencias de
// escalares (bloque "- x" o inline "[a, b]"), mapas inline "{k: v}", escalares
// entrecomillados. Suficiente para pack.yaml, frontmatter, manifiestos, locks y
// degradation.yaml. SIN dependencias externas: el binario es autocontenido.
//
// Valores: string | []any | map[string]any

type yline struct {
	indent int
	text   string
}

func parseYAML(src string) map[string]any {
	ls := tokenize(src)
	if len(ls) == 0 {
		return map[string]any{}
	}
	v, _ := parseBlock(ls, 0, ls[0].indent)
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func tokenize(src string) []yline {
	var out []yline
	for _, raw := range strings.Split(src, "\n") {
		raw = strings.TrimRight(raw, "\r")
		raw = stripComment(raw)
		if strings.TrimSpace(raw) == "" {
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		out = append(out, yline{indent, strings.TrimSpace(raw)})
	}
	return out
}

func stripComment(s string) string {
	inS, inD := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\'' && !inD:
			inS = !inS
		case c == '"' && !inS:
			inD = !inD
		case c == '#' && !inS && !inD && (i == 0 || s[i-1] == ' ' || s[i-1] == '\t'):
			return s[:i]
		}
	}
	return s
}

func isSeq(l yline) bool { return l.text == "-" || strings.HasPrefix(l.text, "- ") }

func parseBlock(ls []yline, pos, indent int) (any, int) {
	if pos >= len(ls) || ls[pos].indent < indent {
		return nil, pos
	}
	if isSeq(ls[pos]) {
		seq := []any{}
		for pos < len(ls) && ls[pos].indent == indent && isSeq(ls[pos]) {
			content := strings.TrimSpace(strings.TrimPrefix(ls[pos].text, "-"))
			seq = append(seq, parseScalar(content))
			pos++
		}
		return seq, pos
	}
	m := map[string]any{}
	for pos < len(ls) && ls[pos].indent == indent && !isSeq(ls[pos]) {
		key, val, ok := splitKV(ls[pos].text)
		if !ok {
			pos++
			continue
		}
		if strings.TrimSpace(val) != "" {
			m[key] = parseScalar(val)
			pos++
		} else {
			ci := childIndent(ls, pos+1, indent)
			child, np := parseBlock(ls, pos+1, ci)
			m[key] = child
			pos = np
		}
	}
	return m, pos
}

func childIndent(ls []yline, pos, parent int) int {
	if pos < len(ls) {
		if isSeq(ls[pos]) && ls[pos].indent == parent {
			return parent
		}
		if ls[pos].indent > parent {
			return ls[pos].indent
		}
	}
	return parent + 2
}

func splitKV(s string) (string, string, bool) {
	inS, inD := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\'' && !inD:
			inS = !inS
		case c == '"' && !inS:
			inD = !inD
		case c == ':' && !inS && !inD && (i+1 >= len(s) || s[i+1] == ' '):
			val := ""
			if i+1 < len(s) {
				val = strings.TrimSpace(s[i+1:])
			}
			return unquote(strings.TrimSpace(s[:i])), val, true
		}
	}
	return "", "", false
}

func parseScalar(s string) any {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		arr := []any{}
		for _, p := range splitTop(s[1:len(s)-1]) {
			if strings.TrimSpace(p) != "" {
				arr = append(arr, parseScalar(p))
			}
		}
		return arr
	}
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		m := map[string]any{}
		for _, p := range splitTop(s[1 : len(s)-1]) {
			kv := strings.SplitN(p, ":", 2)
			if len(kv) == 2 {
				m[unquote(strings.TrimSpace(kv[0]))] = parseScalar(kv[1])
			}
		}
		return m
	}
	return unquote(s)
}

// splitTop divide por comas que no estén dentro de comillas o corchetes.
func splitTop(s string) []string {
	var out []string
	depth, inS, inD, start := 0, false, false, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\'' && !inD:
			inS = !inS
		case c == '"' && !inS:
			inD = !inD
		case (c == '[' || c == '{') && !inS && !inD:
			depth++
		case (c == ']' || c == '}') && !inS && !inD:
			depth--
		case c == ',' && depth == 0 && !inS && !inD:
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
