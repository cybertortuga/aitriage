package entropy

import (
	"strings"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func collectRanges(node *tree_sitter.Node, ranges *[][2]uint) {
	nodeType := node.Kind()
	if nodeType == "comment" || nodeType == "string" || nodeType == "string_literal" || nodeType == "template_string" || nodeType == "raw_string_literal" || nodeType == "interpreted_string_literal" {
		*ranges = append(*ranges, [2]uint{node.StartByte(), node.EndByte()})
		return
	}

	count := node.ChildCount()
	for i := uint(0); i < count; i++ {
		child := node.Child(i)
		collectRanges(child, ranges)
	}
}

func StripWithAST(content string, tree *tree_sitter.Tree) StripResult {
	var ranges [][2]uint
	root := tree.RootNode()
	collectRanges(root, &ranges)

	var codeBuilder strings.Builder
	var docsBuilder strings.Builder

	codeBuilder.Grow(len(content))
	docsBuilder.Grow(len(content))

	contentBytes := []byte(content)
	isDoc := make([]bool, len(contentBytes))
	for _, r := range ranges {
		for i := r[0]; i < r[1]; i++ {
			if int(i) < len(isDoc) {
				isDoc[i] = true
			}
		}
	}

	for i, b := range contentBytes {
		if isDoc[i] {
			docsBuilder.WriteByte(b)
			if b == '\n' {
				codeBuilder.WriteByte('\n')
			} else {
				codeBuilder.WriteByte(' ')
			}
		} else {
			codeBuilder.WriteByte(b)
			if b == '\n' {
				docsBuilder.WriteByte('\n')
			} else {
				docsBuilder.WriteByte(' ')
			}
		}
	}

	return StripResult{
		CodeOnly:       codeBuilder.String(),
		DocsAndStrings: docsBuilder.String(),
	}
}
