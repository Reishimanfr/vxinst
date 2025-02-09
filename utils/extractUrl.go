package utils

import "strings"

const (
	prefix    = `\"video_url\":`
	quote     = `\"`
	prefixLen = len(prefix) + 1

	gqlPrefix    = `"video_url":`
	gqlQuote     = `"`
	gqlPrefixLen = len(gqlPrefix) + 1
)

// Extracts the url from escaped json
func ExtractUrl(s string, gql bool) (string, bool) {
	pref := prefix
	qt := quote
	len := prefixLen
	slice := true

	if gql {
		pref = gqlPrefix
		qt = gqlQuote
		len = gqlPrefixLen
		slice = false
	}

	// Thanks a lot for this tyler
	// Find the first prefix
	startIdx := strings.Index(s, pref)
	if startIdx == -1 {
		return "", false
	}

	// Offset start by prefix len
	start := startIdx + len

	end := strings.Index(s[start:], qt)
	if end == -1 {
		return "", false
	}

	result := s[start : start+end]
	result = UnescapeJSONString(result)
	result = strings.ReplaceAll(result, `\/`, `/`)

	if slice {
		return result[1:], true
	}

	return result, true
}
