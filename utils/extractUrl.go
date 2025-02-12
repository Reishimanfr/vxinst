/*
VxInstagram - Blazing fast embedder for instagram posts
Copyright (C) 2025 Bash06

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package utils

import (
	"fmt"
	"strings"
)

const (
	prefix    = `\"video_url\":`
	quote     = `\"`
	prefixLen = len(prefix) + 1
)

// Extracts the video URL from response
func ExtractUrl(s string) (string, bool) {
	pref := prefix
	qt := quote
	len := prefixLen

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

	fmt.Println(result)

	return result[1:], true
}
