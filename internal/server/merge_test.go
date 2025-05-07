package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergePrefixCases(t *testing.T) {
	r := require.New(t)

	original := "line 1\nline 2"
	modified := "line 1\nline 2\nline 3\nline 4"
	r.Equal(modified, Merge(original, modified), "Should keep longer string when one is prefix")
	r.Equal(modified, Merge(modified, original), "Should keep longer string when one is prefix")
}

func TestMergeCommonPrefixDifferentSuffixes(t *testing.T) {
	r := require.New(t)

	// Both have common prefix but different additional lines
	original := "line 1\nline 2\nline 3\nline original 4"
	modified := "line 1\nline 2\nline 3\nline modified 4"
	merged := Merge(original, modified)
	r.Equal("line 1\nline 2\nline 3\nline original 4\nline modified 4", merged, "Should merge lines after common prefix")
}

func TestMergeDifferentPrefixCommonSuffix(t *testing.T) {
	r := require.New(t)

	// Both have different prefixes but common suffix
	original := "line original 1\nline original 2\nline 3"
	modified := "new\nline original 1\nline original 2\nline 3"
	merged := Merge(original, modified)
	r.Equal("new\nline original 1\nline original 2\nline 3", merged, "Should merge lines before common suffix")
}

func TestMergeDivergentBody(t *testing.T) {
	r := require.New(t)

	// Divergent content with common prefix and suffix
	original := "header\noriginal A\noriginal B\nfooter"
	modified := "header\nmodified X\nmodified Y\nfooter"
	merged := Merge(original, modified)
	r.Equal("header\noriginal A\noriginal B\nmodified X\nmodified Y\nfooter", merged)
}

func TestMergeDivergentContent(t *testing.T) {
	r := require.New(t)

	// Complete divergence with small common prefix
	original := "header\noriginal A\noriginal B"
	modified := "header\nmodified X\nmodified Y"
	merged := Merge(original, modified)
	r.Equal("header\noriginal A\noriginal B\nmodified X\nmodified Y", merged)
}

func TestMergeEmptyStrings(t *testing.T) {
	r := require.New(t)

	r.Equal("", Merge("", ""), "Empty strings should merge to empty string")
	r.Equal("content", Merge("", "content"), "Empty original should return modified")
	r.Equal("content", Merge("content", ""), "Empty modified should return original")
}

func TestMergeTrailingNewlines(t *testing.T) {
	r := require.New(t)

	original := "line 1\nline 2\n"
	modified := "line 1\nline 2\nline 3\n"
	r.Equal(modified, Merge(original, modified), "Should handle trailing newlines correctly")
}

func TestMergeDivergentChars(t *testing.T) {
	r := require.New(t)

	original := "abc"
	modified := "adc"
	merged := Merge(original, modified)
	r.Equal("abc\nadc", merged)
}
