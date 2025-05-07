package server

import "strings"

func Merge(s1, s2 string) string {
	if len(s1) == 0 {
		return s2
	}
	if len(s2) == 0 {
		return s1
	}

	// Split both strings into lines
	originalLines := strings.Split(s1, "\n")
	modifiedLines := strings.Split(s2, "\n")

	// Find common prefix lines
	var commonPrefixLength int
	for commonPrefixLength < len(originalLines) &&
		commonPrefixLength < len(modifiedLines) &&
		originalLines[commonPrefixLength] == modifiedLines[commonPrefixLength] {
		commonPrefixLength++
	}

	// Find common suffix lines
	var commonSuffixLength int
	for commonSuffixLength < len(originalLines)-commonPrefixLength &&
		commonSuffixLength < len(modifiedLines)-commonPrefixLength &&
		originalLines[len(originalLines)-1-commonSuffixLength] == modifiedLines[len(modifiedLines)-1-commonSuffixLength] {
		commonSuffixLength++
	}

	// Build result with common prefix
	var resultLines []string
	resultLines = append(resultLines, originalLines[:commonPrefixLength]...)

	// Add the unique middle parts from both files
	originalMiddleStart := commonPrefixLength
	originalMiddleEnd := len(originalLines) - commonSuffixLength
	for i := originalMiddleStart; i < originalMiddleEnd; i++ {
		resultLines = append(resultLines, originalLines[i])
	}

	modifiedMiddleStart := commonPrefixLength
	modifiedMiddleEnd := len(modifiedLines) - commonSuffixLength
	for i := modifiedMiddleStart; i < modifiedMiddleEnd; i++ {
		resultLines = append(resultLines, modifiedLines[i])
	}

	// Add common suffix
	if commonSuffixLength > 0 {
		suffixStart := len(originalLines) - commonSuffixLength
		resultLines = append(resultLines, originalLines[suffixStart:]...)
	}

	return strings.Join(resultLines, "\n")
}
