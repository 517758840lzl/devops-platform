package service

import "strings"

// ParseLineList splits text into non-empty lines, skipping blank lines and # comments.
func ParseLineList(source string) []string {
	lines := strings.Split(source, "\n")
	out := make([]string, 0, len(lines))
	seen := map[string]struct{}{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		out = append(out, line)
	}
	return out
}
