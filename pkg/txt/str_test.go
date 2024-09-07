package txt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPositiveI64ToStr(t *testing.T) {
	r := require.New(t)

	s := I64(1)

	r.Equal("1", s)
}

func TestNegativeI64ToStr(t *testing.T) {
	r := require.New(t)

	s := I64(-1)

	r.Equal("-1", s)
}

func TestZeroI64ToStr(t *testing.T) {
	r := require.New(t)

	s := I64(0)

	r.Equal("0", s)
}

func TestUcfirst(t *testing.T) {
	r := require.New(t)

	res := Ucfirst("abc")

	r.Equal("Abc", res)
}

func TestUcfirstRu(t *testing.T) {
	r := require.New(t)

	res := Ucfirst("абв")

	r.Equal("Абв", res)
}

func TestLcfirst(t *testing.T) {
	r := require.New(t)

	res := Lcfirst("ABC")

	r.Equal("aBC", res)
}

func TestLcfirstRu(t *testing.T) {
	r := require.New(t)

	res := Lcfirst("АБВ")

	r.Equal("аБВ", res)
}

func TestInsertTextAfterHeaderNoHeader(t *testing.T) {
	r := require.New(t)

	content := InsertTextAfterHeader("### header 1\nitem1\nitem2", "### header 5", "new item")

	r.Equal("### header 5\nnew item\n### header 1\nitem1\nitem2", content)
}

func TestInsertTextAfterHeader(t *testing.T) {
	r := require.New(t)

	content := InsertTextAfterHeader("### header 1\nitem1\nitem2\n### header 2", "### header 1", "new item")

	r.Equal("### header 1\nnew item\nitem1\nitem2\n### header 2", content)
}

func TestInsertTextAfterHeaderInTheMiddle(t *testing.T) {
	r := require.New(t)

	content := InsertTextAfterHeader("### header 0\n### header 1\nitem1\nitem2\n### header 2", "### header 1", "new item")

	r.Equal("### header 0\n### header 1\nnew item\nitem1\nitem2\n### header 2", content)
}

func TestInsertTextAfterHeaderInTheMiddleOnlyHeader(t *testing.T) {
	r := require.New(t)

	content := InsertTextAfterHeader("### header 0\n### header 1\n### header 2", "### header 1", "new item")

	r.Equal("### header 0\n### header 1\nnew item\n### header 2", content)
}

func TestEscapeHTMLInMarkdown(t *testing.T) {
	r := require.New(t)

	// Test case 1: Simple text with HTML
	input := "<b>bold</b>"
	expected := "&lt;b&gt;bold&lt;/b&gt;"
	r.Equal(expected, EscapeHTMLInMarkdown(input), "should escape HTML tags")

	// Test case 2: Text with inline code blocks
	input = "a`<b>bold</b>`"
	expected = "a`<b>bold</b>`" // Inline code block should not be escaped
	r.Equal(expected, EscapeHTMLInMarkdown(input), "should preserve inline code blocks")

	// Test case 3: Text with multiline code blocks
	input = "a```<b>bold</b>\n<p>paragraph</p>\n```"
	expected = "a```<b>bold</b>\n<p>paragraph</p>\n```" // Multiline code block should not be escaped
	r.Equal(expected, EscapeHTMLInMarkdown(input), "should preserve multiline code blocks")

	// Test case 4: Mixed text with both HTML and code blocks
	input = "Some text `<i>italic</i>` and a code block:\n```\n<h1>Title</h1>\n```"
	expected = "Some text `<i>italic</i>` and a code block:\n```\n<h1>Title</h1>\n```"
	r.Equal(expected, EscapeHTMLInMarkdown(input), "should escape text outside code blocks and preserve code blocks")

	// Test case 5: Text with special HTML characters
	input = "5 > 3 & 2 < 4"
	expected = "5 &gt; 3 &amp; 2 &lt; 4"
	r.Equal(expected, EscapeHTMLInMarkdown(input), "should escape special HTML characters")
}
