package service

import (
	"regexp"
	"strings"
)

var dartConstKeyRe = regexp.MustCompile(`^\s*static\s+const\s+String\s+(\w+)\s*=\s*(.*)$`)

// ParseDartStringConsts extracts static const String key = 'value' pairs from Dart source.
func ParseDartStringConsts(source string) map[string]string {
	keys := map[string]string{}
	lines := strings.Split(source, "\n")
	for i := 0; i < len(lines); i++ {
		line := stripDartLineComment(lines[i])
		m := dartConstKeyRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		key := m[1]
		rest := strings.TrimSpace(m[2])
		for !isCompleteDartLiteral(rest) && i+1 < len(lines) {
			i++
			rest += " " + strings.TrimSpace(stripDartLineComment(lines[i]))
		}
		if val, ok := parseDartStringLiteral(rest); ok {
			keys[key] = val
		}
	}
	return keys
}

func stripDartLineComment(line string) string {
	if idx := strings.Index(line, "//"); idx >= 0 {
		return line[:idx]
	}
	return line
}

func isCompleteDartLiteral(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if !strings.HasSuffix(s, ";") {
		return false
	}
	s = strings.TrimSuffix(s, ";")
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return false
	}
	q := s[0]
	if q != '\'' && q != '"' {
		return false
	}
	return unescapedQuoteCount(s[1:], q)%2 == 1
}

func parseDartStringLiteral(s string) (string, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ";")
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return "", false
	}
	q := s[0]
	if q != '\'' && q != '"' {
		return "", false
	}
	body := s[1:]
	var b strings.Builder
	for i := 0; i < len(body); i++ {
		ch := body[i]
		if ch == '\\' && i+1 < len(body) {
			next := body[i+1]
			switch next {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case '\\', '\'', '"':
				b.WriteByte(next)
			default:
				b.WriteByte('\\')
				b.WriteByte(next)
			}
			i++
			continue
		}
		if ch == q {
			return b.String(), true
		}
		b.WriteByte(ch)
	}
	return "", false
}

func unescapedQuoteCount(s string, q byte) int {
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			continue
		}
		if s[i] == q {
			count++
		}
	}
	return count
}
