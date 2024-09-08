package txt

import (
	"regexp"
	"strings"
)

type parser func(input string) []result
type result struct {
	consumed string
	left     string
}

var openTags = map[string]string{
	"*":  "<i>",
	"**": "<b>",
	"_":  "<i>",
	"__": "<b>",
}

var closeTags = map[string]string{
	"*":  "</i>",
	"**": "</b>",
	"_":  "</i>",
	"__": "</b>",
}

// MarkdownToHTML naively converts user's markdown to Telegram-supported subset of HTML.
// We don't need to implement full-blown AST parser because TG only supports a few HTML tags.
// Telegram supports the following HTML tags:
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
func MarkdownToHTML(md string) string {
	mdWithoutCode := EscapeHTML(md)
	mdWithoutCode, codePlaceholders := ReplaceWithPlaceholders(mdWithoutCode, "(?s)```.*?```", "c0debl0ck")
	mdWithoutCode, inlinePlaceholders := ReplaceWithPlaceholders(mdWithoutCode, "`[^`]+`", "inl1ne")
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
	reCodeBlock := regexp.MustCompile("(?s)```(.+?)```")
	mdWithCode = reCodeBlock.ReplaceAllStringFunc(mdWithCode, func(s string) string {
		return "<pre>" + strings.TrimSpace(reCodeBlock.FindStringSubmatch(s)[1]) + "</pre>"
	})
	reInlineCode := regexp.MustCompile("`([^`]+?)`")
	mdWithCode = reInlineCode.ReplaceAllString(mdWithCode, "<code>$1</code>")

	// Convert #+ Header to <b>Header</b>
	reHeader := regexp.MustCompile(`(?m)^#+\s*(.+)`)
	mdWithCode = reHeader.ReplaceAllString(mdWithCode, "<b>$1</b>")

	return mdWithCode
}

// parser Combinators. Watch an amazing video here: https://youtu.be/dDtZLm7HIJs.
// We only support one level of nesting for bold and italic.
func markdown() parser {
	text := notMarkdown()
	italicNoBold := or(
		and(open("*"), text, close("*")),
		and(open("_"), text, close("_")),
	)
	bold := or(
		and(open("**"), some(or(text, italicNoBold)), close("**")),
		and(open("__"), some(or(text, italicNoBold)), close("__")),
	)
	italic := or(
		and(open("*"), some(or(text, bold)), close("*")),
		and(open("_"), some(or(text, bold)), close("_")),
	)
	span := or(bold, italic, text)

	return some(span)
}

// open opens the tag
func open(t string) parser {
	return func(input string) []result {
		if strings.HasPrefix(input, t) {
			return []result{{openTags[t], input[len(t):]}}
		}
		return nil
	}
}

// close closes the tag
func close(t string) parser {
	return func(input string) []result {
		if strings.HasPrefix(input, t) {
			return []result{{closeTags[t], input[len(t):]}}
		}
		return nil
	}
}

// or applies multiple parsers and returns the result of the first successful parser.
// If no parser succeeds, it returns an empty result.
func or(parsers ...parser) parser {
	return func(input string) []result {
		var results []result
		for _, p := range parsers {
			results = append(results, p(input)...)
		}
		return results
	}
}

// and applies multiple parsers in sequence.
// Each parser must succeed in consuming part of the input.
// If any parser fails, the whole parse fails.
func and(parsers ...parser) parser {
	return func(input string) []result {
		results := []result{{"", input}}

		for _, p := range parsers {
			var newResults []result
			for _, r := range results {
				for _, parsed := range p(r.left) {
					if parsed.consumed != "" {
						newResults = append(newResults, result{r.consumed + parsed.consumed, parsed.left})
					}
				}
			}
			if len(newResults) == 0 {
				return nil
			}
			results = newResults
		}

		return results
	}
}

// some applies the parser one or more times. Each parse result is combined with the previous result.
// And each parse can generate multiple results.
func some(parser parser) parser {
	return func(input string) []result {
		return recursive(input, parser, 0)
	}
}

func recursive(input string, parser parser, depth int) []result {
	var results []result
	empty := true
	for _, item := range parser(input) {
		if item.consumed == "" {
			continue
		}
		empty = false
		for _, child := range recursive(item.left, parser, depth+1) {
			results = append(results, result{item.consumed + child.consumed, child.left})
		}
	}
	if empty && depth != 0 {
		results = append(results, result{"", input})
	}

	return results
}

// notMarkdown incrementally yields when it encounters a *, **, _, __
func notMarkdown() parser {
	return func(input string) []result {
		for i, ch := range input {
			if ch == '*' || ch == '_' {
				return []result{{input[:i], input[i:]}}
			}
		}
		if len(input) > 0 && (input[len(input)-1] == '*' || input[len(input)-1] != '_' || input[len(input)-1] != '`') {
			return []result{{input, ""}}
		}
		if len(input) > 0 {
			return []result{{input, ""}}
		}
		return nil
	}
}
