package mdd

import (
	"strings"

	"github.com/russross/blackfriday/v2"
)

// plainTextRenderer removes inline markups and return plain text
func plainTextRenderer(node *blackfriday.Node) string {
	var builder strings.Builder
	node.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		switch node.Type {
		case blackfriday.Image:
			return blackfriday.SkipChildren // Image label is not needed
		case blackfriday.Text:
			builder.Write(node.Literal)
		case blackfriday.Code:
			builder.Write(node.Literal)
		}
		return blackfriday.GoToNext
	})
	return builder.String()
}
