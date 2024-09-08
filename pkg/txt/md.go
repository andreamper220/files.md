package txt

import (
	"regexp"
	"strings"
)

type Parser func(input string) []token

type token struct {
	consumed string
	left     string
}

var openTags = map[string]string{
	"*":  "<i>",
	"**": "<b>",
	"_":  "<i>",
	"__": "<b>",
	"`":  "<code>",
}

var closeTags = map[string]string{
	"*":  "</i>",
	"**": "</b>",
	"_":  "</i>",
	"__": "</b>",
	"`":  "</code>",
}

// MDtoHTML naively converts user's markdown to Telegram-supported subset of HTML.
// We don't need to implement full-blown AST parser because TG only supports a few HTML tags.
// Telegram supported tags:
// <b>bold</b>, <strong>bold</strong>
// <i>italic</i>, <em>italic</em>
// <u>underline</u>, <ins>underline</ins>
// <s>strikethrough</s>, <strike>strikethrough</strike>, <del>strikethrough</del>
// <span class="tg-spoiler">spoiler</span>, <tg-spoiler>spoiler</tg-spoiler>
// <b>bold <i>italic bold <s>italic bold strikethrough <span class="tg-spoiler">italic bold strikethrough spoiler</span></s> <u>underline italic bold</u></i> bold</b>
// <a href="http://www.example.com/">inline URL</a>
// <a href="tg://user?id=123456789">inline mention of a user</a>
// <tg-emoji emoji-id="5368324170671202286">👍</tg-emoji>
// <code>inline fixed-width code</code>
// <pre>pre-formatted fixed-width code block</pre>
// <pre><code class="language-python">pre-formatted fixed-width code block written in the Python programming language</code></pre>
// <blockquote>Block quotation started\nBlock quotation continued\nThe last line of the block quotation</blockquote>
// <blockquote expandable>Expandable block quotation started\nExpandable block quotation continued\nExpandable block quotation continued\nHidden by default part of the block quotation started\nExpandable block quotation continued\nThe last line of the block quotation</blockquote>
func MDtoHTML(md string) string {
	mdWithoutCode := EscapeHTML(md)
	mdWithoutCode, codePlaceholders := ReplaceWithPlaceholders(mdWithoutCode, "(?s)```.*?```", "c0debl0ck")
	mdWithoutCode, inlinePlaceholders := ReplaceWithPlaceholders(mdWithoutCode, "`[^`]*`", "inl1ne")
	// By this point our markdown is safe to send as HTML via Telegram.
	// There won't be any issues like "missing closing HTML tag",
	// for the cases when our markdown has some html tags.
	// We try to convert as much markdown as possible to Telegram HTML.

	// We split by \n\n, because markdown context is broken by \n\n (excluding code inside ```)
	segments := strings.Split(mdWithoutCode, "\n\n")
	processedSegments := make([]string, len(segments))
	for i, segment := range segments {
		// Process each segment separately
		docs := markdown()(segment)
		if len(docs) > 0 {
			segment = docs[0].consumed + docs[0].left
		}
		processedSegments[i] = segment
	}
	mdWithoutCode = strings.Join(processedSegments, "\n\n")

	mdWithCode := RestoreFromPlaceholders(mdWithoutCode, codePlaceholders)
	mdWithCode = RestoreFromPlaceholders(mdWithCode, inlinePlaceholders)

	// We do dirty but simple md -> html conversion.
	// Covert ` and ``` to <pre> and <code> HTML tags
	reCodeBlock := regexp.MustCompile("(?s)```(.*?)```")
	mdWithCode = reCodeBlock.ReplaceAllString(mdWithCode, "<pre>$1</pre>")
	reInlineCode := regexp.MustCompile("`(.*?)`")
	mdWithCode = reInlineCode.ReplaceAllString(mdWithCode, "<code>$1</code>")

	// Convert #+ Header to <b>Header</b>
	reHeader := regexp.MustCompile(`(?m)^#+\s*(.+)`)
	mdWithCode = reHeader.ReplaceAllString(mdWithCode, "<b>$1</b>")

	return mdWithCode
}

// Parser Combinators. Watch an amazing video here: https://youtu.be/dDtZLm7HIJs
func markdown() Parser {
	text := notMarkdown()
	onlyBold := or(
		and(openTerm("**"), and(text, closeTerm("**"))),
		and(openTerm("__"), and(text, closeTerm("__"))),
	)
	italicNoCyclic := or(
		and(openTerm("*"), and(oneOrMore(or(
			onlyBold,
			text)),
			closeTerm("*"))),
		and(openTerm("_"), and(oneOrMore(or(
			onlyBold,
			text)),
			closeTerm("_"))),
	)
	onlyItalic := or(
		and(openTerm("*"), and(text, closeTerm("*"))),
		and(openTerm("_"), and(text, closeTerm("_"))),
	)
	boldNoCyclic := or(
		and(openTerm("**"), and(oneOrMore(or(
			onlyItalic,
			text)),
			closeTerm("**"))),
		and(openTerm("__"), and(oneOrMore(or(
			onlyItalic,
			text)),
			closeTerm("__"))),
	)
	italic := or(
		and(openTerm("*"), and(oneOrMore(or(boldNoCyclic, text)), closeTerm("*"))),
		and(openTerm("_"), and(oneOrMore(or(boldNoCyclic, text)), closeTerm("_"))),
	)
	bold := or(
		and(openTerm("**"), and(oneOrMore(or(italicNoCyclic, text)), closeTerm("**"))),
		and(openTerm("__"), and(oneOrMore(or(italicNoCyclic, text)), closeTerm("__"))),
	)
	span := or(bold, or(italic, text))

	return oneOrMore(span)
}

func openTerm(t string) Parser {
	return func(input string) []token {
		if strings.HasPrefix(input, t) {
			return []token{{openTags[t], input[len(t):]}}
		}
		return nil
	}
}

func closeTerm(t string) Parser {
	return func(input string) []token {
		if strings.HasPrefix(input, t) {
			return []token{{closeTags[t], input[len(t):]}}
		}
		return nil
	}
}

func or(lhs, rhs Parser) Parser {
	return func(input string) []token {
		return append(lhs(input), rhs(input)...)
	}
}

func and(lhs, rhs Parser) Parser {
	return func(input string) []token {
		var results []token
		for _, litem := range lhs(input) {
			for _, ritem := range rhs(litem.left) {
				if litem.consumed != "" && ritem.consumed != "" {
					results = append(results, token{litem.consumed + ritem.consumed, ritem.left})
				}
			}
		}
		return results
	}
}

func recursive(input string, parser Parser, depth int) []token {
	var results []token
	empty := true
	for _, item := range parser(input) {
		if item.consumed == "" {
			continue
		}
		empty = false
		for _, child := range recursive(item.left, parser, depth+1) {
			results = append(results, token{item.consumed + child.consumed, child.left})
		}
	}
	if empty && depth != 0 {
		results = append(results, token{"", input})
	}

	return results
}

// oneOrMore applies the parser for more than one time. Each parse result is combined with the previous result.
// And each parse can generate multiple results.
func oneOrMore(parser Parser) Parser {
	return func(input string) []token {
		return recursive(input, parser, 0)
	}
}

// notMarkdown incrementally yields when it encounters a *, **, _, __
func notMarkdown() Parser {
	return func(input string) []token {
		for i, ch := range input {
			if ch == '*' || ch == '_' {
				return []token{{input[:i], input[i:]}}
			}
		}
		if len(input) > 0 && (input[len(input)-1] == '*' || input[len(input)-1] != '_' || input[len(input)-1] != '`') {
			return []token{{input, ""}}
		}
		if len(input) > 0 {
			return []token{{input, ""}}
		}
		return nil
	}
}
