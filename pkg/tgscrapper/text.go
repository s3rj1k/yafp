package tgscrapper

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aquilax/truncate"
	"golang.org/x/net/html"
)

var ellipsizeRegExp = regexp.MustCompile(`(\(.+\)|\[.+\]|\{.+\})`)

func Textify(s *goquery.Selection) string {
	var buf bytes.Buffer

	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			_, err := buf.WriteString(n.Data)
			if err != nil {
				panic(err)
			}
		}

		if n.Type == html.ElementNode && (n.Data == "br" || n.Data == "p") {
			_, err := buf.WriteString("\n")
			if err != nil {
				panic(err)
			}
		}

		if n.FirstChild != nil {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
	}
	for _, n := range s.Nodes {
		f(n)
	}

	return buf.String()
}

func Ellipsize(s string) string {
	lines := strings.FieldsFunc(strings.TrimSpace(s), func(r rune) bool {
		return r == '\n' || r == '\f' || r == '\t' || r == '\v'
	})

	return strings.TrimSpace(
		truncate.Truncate(ellipsizeRegExp.
			ReplaceAllLiteralString(lines[0], ""),
			maxNumberOfSymbolsInEllipsizeMessageTitle, "...", truncate.PositionEnd,
		),
	)
}
