package ast

import (
	"testing"
)

func TestCommentText(t *testing.T) {
	testCases := []struct {
		list []string
		text string
	}{
		{[]string{"//"}, ""},
		{[]string{"//   "}, ""},
		{[]string{"//", "//", "//   "}, ""},
		{[]string{"// foo   "}, "foo\n"},
		{[]string{"//", "//", "// foo"}, "foo\n"},
		{[]string{"// foo  bar  "}, "foo  bar\n"},
		{[]string{"// foo", "// bar"}, "foo\nbar\n"},
		{[]string{"// foo", "//", "//", "//", "// bar"}, "foo\n\nbar\n"},
		{[]string{"// foo", "/* bar */"}, "foo\n bar\n"},
		{[]string{"//", "//", "//", "// foo", "//", "//", "//"}, "foo\n"},

		{[]string{"/**/"}, ""},
		{[]string{"/*   */"}, ""},
		{[]string{"/**/", "/**/", "/*   */"}, ""},
		{[]string{"/* Foo   */"}, " Foo\n"},
		{[]string{"/* Foo  Bar  */"}, " Foo  Bar\n"},
		{[]string{"/* Foo*/", "/* Bar*/"}, " Foo\n Bar\n"},
		{[]string{"/* Foo*/", "/**/", "/**/", "/**/", "// Bar"}, " Foo\n\nBar\n"},
		{[]string{"/* Foo*/", "/*\n*/", "//", "/*\n*/", "// Bar"}, " Foo\n\nBar\n"},
		{[]string{"/* Foo*/", "// Bar"}, " Foo\nBar\n"},
		{[]string{"/* Foo\n Bar*/"}, " Foo\n Bar\n"},
	}

	for i, tt := range testCases {
		list := make([]*Comment, len(tt.list))
		for i, s := range tt.list {
			list[i] = &Comment{Text: s}
		}

		cg := &CommentGroup{list}
		text := cg.Text()
		if text != tt.text {
			t.Errorf("case %d: have %q; want %q", i, text, tt.text)
		}
	}
}
