package utils

import (
	"regexp"
	"strings"
)

var (
	scriptTagRegex  = regexp.MustCompile(`(?i)<script[\s\S]*?>[\s\S]*?</script>`)
	iframeTagRegex  = regexp.MustCompile(`(?i)<iframe[\s\S]*?>[\s\S]*?</iframe>`)
	svgTagRegex     = regexp.MustCompile(`(?i)<svg[\s\S]*?>[\s\S]*?</svg>`)
	onEventRegex    = regexp.MustCompile(`(?i)\son[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	javascriptRegex = regexp.MustCompile(`(?i)(href|src|action)\s*=\s*("javascript:[^"]*"|'javascript:[^']*')`)
	expressionRegex = regexp.MustCompile(`(?i)\sexpression\s*\(`)
)

func SanitizeHTML(input string) string {
	if input == "" {
		return input
	}

	result := input

	result = scriptTagRegex.ReplaceAllString(result, "")
	result = iframeTagRegex.ReplaceAllString(result, "")
	result = svgTagRegex.ReplaceAllString(result, "")
	result = onEventRegex.ReplaceAllString(result, "")
	result = javascriptRegex.ReplaceAllString(result, "$1=\"#\"")
	result = expressionRegex.ReplaceAllString(result, "")

	result = strings.TrimSpace(result)

	return result
}
