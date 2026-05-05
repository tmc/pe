package security

import (
	"regexp"
	"strings"
)

var secretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(authorization:\s*(?:bearer|token)\s+)[^\s"'{}]+`),
	regexp.MustCompile(`(?i)((?:api[_-]?key|secret|token|password|passwd|credential|client[_-]?secret)\s*[:=]\s*)["']?[^"'\s,}]+`),
	regexp.MustCompile(`sk-[A-Za-z0-9_-]{20,}`),
	regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{20,}`),
	regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	regexp.MustCompile(`AIza[0-9A-Za-z_-]{35}`),
}

// RedactSecrets replaces common API keys and credentials in diagnostic text.
func RedactSecrets(s string) string {
	if s == "" {
		return s
	}
	for _, re := range secretPatterns {
		s = re.ReplaceAllStringFunc(s, func(match string) string {
			if i := strings.IndexAny(match, ":="); i >= 0 {
				return match[:i+1] + " [REDACTED]"
			}
			fields := strings.Fields(match)
			if len(fields) >= 2 && strings.HasSuffix(fields[0], ":") {
				return fields[0] + " [REDACTED]"
			}
			return "[REDACTED]"
		})
	}
	return s
}
