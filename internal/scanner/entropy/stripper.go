package entropy

import (
	"strings"
)

// StripResult contains the separate parts of source code
type StripResult struct {
	CodeOnly       string // Source code without comments and string contents
	DocsAndStrings string // Only comments and string values
}

// Strip parses source code and separates logical code from human-readable text (strings/comments).
// It's a lightweight, language-agnostic state machine that avoids heavy AST parsing.
// It supports JS, TS, Go, and Python basic comment and string structures.
func Strip(content string) StripResult {
	var codeBuilder strings.Builder
	var docsBuilder strings.Builder

	inString := false
	stringChar := byte(0) // '"', '\'', '`'

	inSingleLineComment := false

	inMultiLineComment := false
	multiLineCommentType := 0 // 1 for /* */, 2 for """ """, 3 for ''' '''

	escapeNext := false

	length := len(content)

	codeBuilder.Grow(length)
	docsBuilder.Grow(length / 4) // heuristic

	for i := 0; i < length; i++ {
		c := content[i]

		// If we are escaping the next character inside a string
		if inString && escapeNext {
			docsBuilder.WriteByte(c)
			escapeNext = false
			continue
		}

		if inString && c == '\\' {
			docsBuilder.WriteByte(c)
			escapeNext = true
			continue
		}

		// Handle closing string
		if inString && c == stringChar {
			docsBuilder.WriteByte(c)
			codeBuilder.WriteByte(c) // Keep the quotes in code
			inString = false
			continue
		}

		// Handle single-line comment end (newline)
		if inSingleLineComment && (c == '\n' || c == '\r') {
			inSingleLineComment = false
			docsBuilder.WriteByte(c)
			codeBuilder.WriteByte(c) // preserve lines
			continue
		}

		// Handle multi-line comment end
		if inMultiLineComment {
			docsBuilder.WriteByte(c)
			if multiLineCommentType == 1 && c == '*' && i+1 < length && content[i+1] == '/' {
				docsBuilder.WriteByte('/')
				i++
				inMultiLineComment = false
			} else if multiLineCommentType == 2 && c == '"' && i+2 < length && content[i+1] == '"' && content[i+2] == '"' {
				docsBuilder.WriteString(`""`)
				i += 2
				inMultiLineComment = false
			} else if multiLineCommentType == 3 && c == '\'' && i+2 < length && content[i+1] == '\'' && content[i+2] == '\'' {
				docsBuilder.WriteString("''")
				i += 2
				inMultiLineComment = false
			}
			continue
		}

		// Not in any special state, check for transitions
		if !inString && !inSingleLineComment && !inMultiLineComment {
			// Check Python multi-line comment """ or '''
			if c == '"' && i+2 < length && content[i+1] == '"' && content[i+2] == '"' {
				inMultiLineComment = true
				multiLineCommentType = 2
				docsBuilder.WriteString(`"""`)
				i += 2
				continue
			}
			if c == '\'' && i+2 < length && content[i+1] == '\'' && content[i+2] == '\'' {
				inMultiLineComment = true
				multiLineCommentType = 3
				docsBuilder.WriteString(`'''`)
				i += 2
				continue
			}

			// Check JS/TS/Go multi-line comment /*
			if c == '/' && i+1 < length && content[i+1] == '*' {
				inMultiLineComment = true
				multiLineCommentType = 1
				docsBuilder.WriteString("/*")
				i++
				continue
			}

			// Check JS/TS/Go single-line comment //
			if c == '/' && i+1 < length && content[i+1] == '/' {
				inSingleLineComment = true
				docsBuilder.WriteString("//")
				i++
				continue
			}

			// Check Python single-line comment #
			if c == '#' {
				inSingleLineComment = true
				docsBuilder.WriteByte('#')
				continue
			}

			// Check Strings
			if c == '"' || c == '\'' || c == '`' {
				inString = true
				stringChar = c
				docsBuilder.WriteByte(c)
				codeBuilder.WriteByte(c) // preserve quotes in code too
				continue
			}

			// Normal code character
			codeBuilder.WriteByte(c)
			// Docs builder might get spaced out to keep index alignments if we wanted to, but simple concatenation is fine usually.
			// Let's replace code with whitespace in docsBuilder to maintain line counts for regex searching (useful for finding line numbers).
			if c == '\n' {
				docsBuilder.WriteByte('\n')
			} else {
				docsBuilder.WriteByte(' ')
			}
		} else if inString {
			// Inside string content
			docsBuilder.WriteByte(c)
			if c == '\n' {
				codeBuilder.WriteByte('\n') // maintain lines
			}
		} else if inSingleLineComment {
			docsBuilder.WriteByte(c)
		}
	}

	return StripResult{
		CodeOnly:       codeBuilder.String(),
		DocsAndStrings: docsBuilder.String(),
	}
}
