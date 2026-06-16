package entropy

import (
	"testing"
)

func TestStrip(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectedCode string
		expectedDocs string
	}{
		{
			name:         "Go simple comment",
			content:      "func main() {\n\t// this is a comment\n\tx := 1\n}",
			expectedCode: "func main() {\n\t\n\tx := 1\n}",
			expectedDocs: "             \n // this is a comment\n       \n ",
		},
		{
			name:         "Go multi-line comment",
			content:      "/*\n multi \n line \n*/\nvar x = 1;",
			expectedCode: "\nvar x = 1;",
			expectedDocs: "/*\n multi \n line \n*/\n          ",
		},
		{
			name:         "Python string doc",
			content:      "\"\"\"docstring\"\"\"\ndef f(): pass",
			expectedCode: "\ndef f(): pass",
			expectedDocs: "\"\"\"docstring\"\"\"\n             ",
		},
		{
			name:         "Python single quote string doc",
			content:      "'''docstring'''\ndef f(): pass",
			expectedCode: "\ndef f(): pass",
			expectedDocs: "'''docstring'''\n             ",
		},
		{
			name:         "String with escapes",
			content:      "var s = \"string with \\\"escape\\\"\";",
			expectedCode: "var s = \"\";",
			expectedDocs: "        \"string with \\\"escape\\\"\" ",
		},
		{
			name:         "Mixed content",
			content:      "func test() { // comment\n  s := \"hello\" /* inline */\n}",
			expectedCode: "func test() { \n  s := \"\" \n}",
			expectedDocs: "              // comment\n       \"hello\" /* inline */\n ",
		},
		{
			name:         "Python single line comment",
			content:      "def test():\n    # python comment\n    print('hello')",
			expectedCode: "def test():\n    \n    print('')",
			expectedDocs: "           \n    # python comment\n          'hello' ",
		},
		{
			name:         "Empty string",
			content:      "",
			expectedCode: "",
			expectedDocs: "",
		},
		{
			name:         "Only code",
			content:      "var x = 1;",
			expectedCode: "var x = 1;",
			expectedDocs: "          ",
		},
		{
			name:         "Carriage return in single line comment",
			content:      "// comment\rvar x = 1;",
			expectedCode: "\rvar x = 1;",
			expectedDocs: "// comment\r          ",
		},
		{
			name:         "Multiline comment at EOF",
			content:      "/* unclosed comment",
			expectedCode: "",
			expectedDocs: "/* unclosed comment",
		},
		{
			name:         "Backslash escape at EOF",
			content:      "\"unterminated string \\",
			expectedCode: "\"",
			expectedDocs: "\"unterminated string \\",
		},
		{
			name:         "Newline inside string",
			content:      "var s = `line1\nline2`;",
			expectedCode: "var s = `\n`;",
			expectedDocs: "        `line1\nline2` ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := Strip(tc.content)
			if res.CodeOnly != tc.expectedCode {
				t.Errorf("\nExpected CodeOnly:\n%q\nGot:\n%q", tc.expectedCode, res.CodeOnly)
			}
			if res.DocsAndStrings != tc.expectedDocs {
				t.Errorf("\nExpected DocsAndStrings:\n%q\nGot:\n%q", tc.expectedDocs, res.DocsAndStrings)
			}
		})
	}
}
