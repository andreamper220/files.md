package txt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarkdownToHTML(t *testing.T) {
	r := require.New(t)

	md := "### Header\n**bold**\n*italic*"
	html := MarkdownToHTML(md)

	r.Equal("<b>Header</b>\n<b>bold</b>\n<i>italic</i>", html)
}

func TestMarkdownToHTMLHeaderAndText(t *testing.T) {
	r := require.New(t)

	md := "# Header\nText"
	html := MarkdownToHTML(md)

	r.Equal("<b>Header</b>\nText", html)
}

func TestMarkdownToHTMLBold(t *testing.T) {
	r := require.New(t)

	md := "**bold**"
	html := MarkdownToHTML(md)

	r.Equal("<b>bold</b>", html)
}

func TestMarkdownToHTMLMultilineBold(t *testing.T) {
	r := require.New(t)

	md := "**bold\nstill bold**"
	html := MarkdownToHTML(md)

	r.Equal("<b>bold\nstill bold</b>", html)
}

func TestMarkdownToHTMLEmptyBoldAndItalic(t *testing.T) {
	r := require.New(t)

	md := "**"
	r.Equal("**", MarkdownToHTML(md))
	md = "*"
	r.Equal("*", MarkdownToHTML(md))

	md = "__"
	r.Equal("__", MarkdownToHTML(md))
	md = "_"
	r.Equal("_", MarkdownToHTML(md))
}

func TestMarkdownToHTMLNewLineChar(t *testing.T) {
	r := require.New(t)

	bold := "**\n**"
	r.Equal("<b>\n</b>", MarkdownToHTML(bold))

	italic := "*\n*"
	r.Equal("<i>\n</i>", MarkdownToHTML(italic))
}

func TestMarkdownToHTMLCharAndNewLineChar(t *testing.T) {
	r := require.New(t)

	bold := "**a\n**"
	r.Equal("<b>a\n</b>", MarkdownToHTML(bold))

	italic := "*a\n*"
	r.Equal("<i>a\n</i>", MarkdownToHTML(italic))

}

func TestMarkdownToHTMLNewLineAndChar(t *testing.T) {
	r := require.New(t)

	bold := "**\na**"
	r.Equal("<b>\na</b>", MarkdownToHTML(bold))

	italic := "*\na*"
	r.Equal("<i>\na</i>", MarkdownToHTML(italic))
}

func TestMarkdownToHTMLTwoNewlinesBreakFormatting(t *testing.T) {
	r := require.New(t)

	bold := "**no bold\n\nno bold**"
	r.Equal("**no bold\n\nno bold**", MarkdownToHTML(bold))

	italic := "*no italic\n\nno italic*"
	r.Equal("*no italic\n\nno italic*", MarkdownToHTML(italic))
}

func TestMarkdownToHTMLMultilineBoldAndItalic(t *testing.T) {
	r := require.New(t)

	md := "Some _italic text\nin two lines_, **bold text\nin two lines**, and ***bold italic text\nin two lines***."
	html := MarkdownToHTML(md)

	r.Equal("Some <i>italic text\nin two lines</i>, <b>bold text\nin two lines</b>, and <b><i>bold italic text\nin two lines</i></b>.", html)
}

func TestMarkdownToHTMLHtmlInsideCode(t *testing.T) {
	r := require.New(t)

	md := "```some code a > b```"
	html := MarkdownToHTML(md)

	r.Equal("<pre>some code a &gt; b</pre>", html)
}

func TestMarkdownToHTMLInlineCodeAndCodeBlock(t *testing.T) {
	r := require.New(t)

	md := "`inline code` and ```code block```"
	html := MarkdownToHTML(md)

	r.Equal("<code>inline code</code> and <pre>code block</pre>", html)
}

func TestMarkdownToHTMLInlineCodeBlockInsideCodeBlock(t *testing.T) {
	r := require.New(t)

	md := "```code block with `inline code` inside```"
	html := MarkdownToHTML(md)

	r.Equal("<pre>code block with <code>inline code</code> inside</pre>", html)
}

func TestMarkdownToHTMLItalic(t *testing.T) {
	r := require.New(t)

	md := "*italic*"
	html := MarkdownToHTML(md)

	r.Equal("<i>italic</i>", html)
}

func TestMarkdownToHTMLInvalidMD(t *testing.T) {
	r := require.New(t)

	md := "__valid__**invalid"
	r.Equal("<b>valid</b>**invalid", MarkdownToHTML(md))

	r.Equal("*invalid_markdown", MarkdownToHTML("*invalid_markdown"))
	r.Equal("*``invalid_markdown", MarkdownToHTML("*``invalid_markdown"))
	r.Equal("*```invalid_markdown", MarkdownToHTML("*```invalid_markdown"))
}

func TestMarkdownToHTMLMultiline(t *testing.T) {
	r := require.New(t)

	md := "line1\n**line2**\nline3"
	html := MarkdownToHTML(md)

	r.Equal("line1\n<b>line2</b>\nline3", html)
}

func TestMarkdownToHTMLBoldInsideItalic(t *testing.T) {
	r := require.New(t)

	md := "*italic and __bold__*"
	r.Equal("<i>italic and <b>bold</b></i>", MarkdownToHTML(md))

	md = "*italic and **bold***"
	r.Equal("<i>italic and <b>bold</b></i>", MarkdownToHTML(md))
}

func TestMarkdownToHTMLItalicInsideBold(t *testing.T) {
	r := require.New(t)

	md := "__bold and _italic___"
	r.Equal("<b>bold and <i>italic</i></b>", MarkdownToHTML(md))

	md = "**bold and *italic***"
	r.Equal("<b>bold and <i>italic</i></b>", MarkdownToHTML(md))
}

func TestMarkdownToHTMLNoLists(t *testing.T) {
	r := require.New(t)

	md := "list\n1) item1\n2) item2"
	html := MarkdownToHTML(md)

	r.Equal("list\n1) item1\n2) item2", html)
}

func TestMarkdownToHTMLEscapeHtml(t *testing.T) {
	r := require.New(t)

	html := MarkdownToHTML("<a> &b")

	r.Equal("&lt;a&gt; &amp;b", html)
}

func TestMarkdownToHTMLHeader(t *testing.T) {
	r := require.New(t)

	md := "Multiline\n# Header"
	html := MarkdownToHTML(md)

	r.Equal("Multiline\n<b>Header</b>", html)
}

func TestMarkdownToHTMLMultipleHeaders(t *testing.T) {
	r := require.New(t)

	md := "# Header1\n## Header2"
	html := MarkdownToHTML(md)

	r.Equal("<b>Header1</b>\n<b>Header2</b>", html)
}

func TestMarkdownToHTMLInlineCode(t *testing.T) {
	r := require.New(t)

	md := "`inline code`"
	html := MarkdownToHTML(md)

	r.Equal("<code>inline code</code>", html)
}

func TestMarkdownToHTMLMultilineCodeBlock(t *testing.T) {
	r := require.New(t)

	md := "```\ncode line 1\ncode line 2\n```"
	html := MarkdownToHTML(md)

	r.Equal("<pre>code line 1\ncode line 2</pre>", html)
}

func TestMarkdownToHTMLCodeWithBold(t *testing.T) {
	r := require.New(t)

	md := "`code` **bold**"
	html := MarkdownToHTML(md)

	r.Equal("<code>code</code> <b>bold</b>", html)
}

func TestMarkdownToHTMLHeaderWithInlineCode(t *testing.T) {
	r := require.New(t)

	md := "# Header\n`inline code`"
	html := MarkdownToHTML(md)

	r.Equal("<b>Header</b>\n<code>inline code</code>", html)
}
