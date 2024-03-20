package gonk

import (
	"strings"
)

func (prefix envLoader) getEnvName(tag Tag) string {
	replacer := strings.NewReplacer(
		"-", "_",
		".", "_",
	)
	parts := []string{}
	if string(prefix) != "" {
		parts = append(parts, string(prefix))
	}
	for _, part := range tag.NamedKeys() {
		parts = append(parts, part)
	}
	out := strings.Join(parts, "_")
	out = strings.ToUpper(out)
	out = replacer.Replace(out)
	return out
}
