package utils

import "strings"

const (
	prefix    = `\"video_url\":`
	quote     = `\"`
	prefixLen = len(prefix) + 1
)

// Extracts the url from escaped json
func ExtractUrl(s string) (string, bool) {
	// Thanks a lot for this tyler
	// Find the first "video_url:"
	startIdx := strings.Index(s, prefix)
	if startIdx == -1 {
		return "", false
	}

	// Offset start by prefix len
	start := startIdx + prefixLen

	end := strings.Index(s[start:], quote)
	if end == -1 {
		return "", false
	}

	result := s[start : start+end]
	result = UnescapeJSONString(result)
	result = strings.ReplaceAll(result, `\/`, `/`)

	return result[1:], true
}
