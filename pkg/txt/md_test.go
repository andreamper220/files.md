package txt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarkdownToHtmlHeader(t *testing.T) {
	r := require.New(t)

	md := `# Header`
	html := MDtoHTML(md)

	r.Equal("<b>Header</b>", html)
}

func TestMarkdownToHtmlHeaderAndText(t *testing.T) {
	r := require.New(t)

	md := "# Header\nText"
	html := MDtoHTML(md)

	r.Equal("<b>Header</b>\nText", html)
}

func TestMarkdownToHtmlBold(t *testing.T) {
	r := require.New(t)

	md := "**bold**"
	html := MDtoHTML(md)

	r.Equal("<b>bold</b>", html)
}

func TestMarkdownToHtmlMultilineBold(t *testing.T) {
	r := require.New(t)

	md := "**bold\nstill bold**"
	html := MDtoHTML(md)

	r.Equal("<b>bold\nstill bold</b>", html)
}

func TestMDToHTMLEmptyBold(t *testing.T) {
	r := require.New(t)

	md := "**"
	html := MDtoHTML(md)

	r.Equal("**", html)
}

func TestMarkdownToHtmlNewLineChar(t *testing.T) {
	r := require.New(t)

	bold := "**\n**"
	r.Equal("<b>\n</b>", MDtoHTML(bold))

	italic := "*\n*"
	r.Equal("<i>\n</i>", MDtoHTML(italic))
}

func TestMarkdownToHtmlCharAndNewLineChar(t *testing.T) {
	r := require.New(t)

	bold := "**a\n**"
	r.Equal("<b>a\n</b>", MDtoHTML(bold))

	italic := "*a\n*"
	r.Equal("<i>a\n</i>", MDtoHTML(italic))

}

func TestMarkdownToHtmlNewLineAndChar(t *testing.T) {
	r := require.New(t)

	bold := "**\na**"
	r.Equal("<b>\na</b>", MDtoHTML(bold))

	italic := "*\na*"
	r.Equal("<i>\na</i>", MDtoHTML(italic))
}

func TestMarkdownToHtmlTwoNewlinesBreakFormatting(t *testing.T) {
	r := require.New(t)

	bold := "**no bold\n\nno bold**"
	r.Equal("**no bold\n\nno bold**", MDtoHTML(bold))

	italic := "*no italic\n\nno italic*"
	r.Equal("*no italic\n\nno italic*", MDtoHTML(italic))
}

func TestMarkdownToHtmlMultilineBoldAndItalic(t *testing.T) {
	r := require.New(t)

	md := "Some _italic text\nin two lines_, **bold text\nin two lines**, and ***bold italic text\nin two lines***."
	html := MDtoHTML(md)

	r.Equal("Some <i>italic text\nin two lines</i>, <b>bold text\nin two lines</b>, and <b><i>bold italic text\nin two lines</i></b>.", html)
}

func TestMDtoHTMLHtmlInsideCode(t *testing.T) {
	r := require.New(t)

	md := "```some code a > b```"
	html := MDtoHTML(md)

	r.Equal("<pre>some code a &gt; b</pre>", html)
}

func TestMarkdownToHtmlItalic(t *testing.T) {
	r := require.New(t)

	md := "*italic*"
	html := MDtoHTML(md)

	r.Equal("<i>italic</i>", html)
}

func TestMarkdownToHtmlInvalid(t *testing.T) {
	r := require.New(t)

	md := "__valid__**invalid"
	html := MDtoHTML(md)

	r.Equal("<b>valid</b>**invalid", html)
}

func TestMarkdownToHtmlMultiline(t *testing.T) {
	r := require.New(t)

	md := "line1\n**line2**\nline3"
	html := MDtoHTML(md)

	r.Equal("line1\n<b>line2</b>\nline3", html)
}

func TestMarkdownToHtmlBoldInsideItalic(t *testing.T) {
	r := require.New(t)

	md := "*italic and __bold__*"
	r.Equal("<i>italic and <b>bold</b></i>", MDtoHTML(md))

	md = "*italic and **bold***"
	r.Equal("<i>italic and <b>bold</b></i>", MDtoHTML(md))
}

func TestMarkdownToHtmlNoLists(t *testing.T) {
	r := require.New(t)

	md := "list\n1) item1\n2) item2"
	html := MDtoHTML(md)

	r.Equal("list\n1) item1\n2) item2", html)
}

func TestMarkdownToHtmlEscapeHtml(t *testing.T) {
	r := require.New(t)

	html := MDtoHTML("<a> &b")

	r.Equal("&lt;a&gt; &amp;b", html)
}

func TestMDToHTMLHeader(t *testing.T) {
	r := require.New(t)

	md := "Multiline\n# Header"
	html := MDtoHTML(md)

	r.Equal("Multiline\n<b>Header</b>", html)
}
