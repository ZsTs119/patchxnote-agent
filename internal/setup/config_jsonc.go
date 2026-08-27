package setup

import "strings"

func looksLikeJSONC(input []byte) bool {
	text := string(input)
	inString := false
	escaped := false
	for index := 0; index < len(text); index++ {
		ch := text[index]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			continue
		}
		if ch == '/' && index+1 < len(text) {
			next := text[index+1]
			if next == '/' || next == '*' {
				return true
			}
		}
	}
	return strings.Contains(text, ",\n}") || strings.Contains(text, ",\r\n}")
}
