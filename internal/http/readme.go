package httpserver

import (
	"html"
	"strings"
)

// readmeHTML converts the small README subset used by the project into HTML:
// headings, paragraphs, lists and fenced code blocks. Mermaid fences get the
// class the vendored renderer looks for.
func readmeHTML(markdown string) string {
	var out strings.Builder
	lines := strings.Split(markdown, "\n")
	inCode := false
	var code strings.Builder
	lang := ""
	paragraph := make([]string, 0)
	flush := func() {
		if len(paragraph) == 0 {
			return
		}
		out.WriteString("<p>" + html.EscapeString(strings.Join(paragraph, " ")) + "</p>\n")
		paragraph = paragraph[:0]
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if inCode {
				cls := ""
				if lang == "mermaid" {
					cls = ` class="language-mermaid"`
				}
				out.WriteString("<pre><code" + cls + ">" + html.EscapeString(code.String()) + "</code></pre>\n")
				code.Reset()
				inCode = false
			} else {
				flush()
				inCode = true
				lang = strings.TrimPrefix(line, "```")
			}
			continue
		}
		if inCode {
			code.WriteString(line + "\n")
			continue
		}
		switch {
		case strings.HasPrefix(line, "### "):
			flush()
			out.WriteString("<h3>" + html.EscapeString(strings.TrimPrefix(line, "### ")) + "</h3>\n")
		case strings.HasPrefix(line, "## "):
			flush()
			out.WriteString("<h2>" + html.EscapeString(strings.TrimPrefix(line, "## ")) + "</h2>\n")
		case strings.HasPrefix(line, "# "):
			flush()
			out.WriteString("<h1>" + html.EscapeString(strings.TrimPrefix(line, "# ")) + "</h1>\n")
		case strings.HasPrefix(line, "- "):
			flush()
			out.WriteString("<li>" + html.EscapeString(strings.TrimPrefix(line, "- ")) + "</li>\n")
		case strings.TrimSpace(line) == "":
			flush()
		default:
			paragraph = append(paragraph, strings.TrimSpace(line))
		}
	}
	flush()
	return out.String()
}
