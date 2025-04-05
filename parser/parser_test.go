package parser

import (
	"strings"
	"testing"

	"github.com/stable-lang/stlang/token"
)

func TestParseDecl(t *testing.T) {
	type testCase struct {
		src     string
		wantErr string
	}

	const pkgPrefix = "package p;"

	t.Run("package", func(t *testing.T) {
		testCases := []testCase{
			{"package p\n", ""},
			{`package p;`, ""},
			{`package main;`, ""},

			{`package _;`, "invalid package name _"},
			{`package builtin;`, "package name 'builtin' is reserved"},
			{`package init;`, "package name 'init' is reserved"},
			{`package internal;`, "package name 'internal' is reserved"},
			{`package vendor;`, "package name 'vendor' is reserved"},

			{`package 123;`, `expected 'Ident', found 123`},
			{`package 'a';`, `expected 'Ident', found 'a'`},
			{`package "pkg";`, `expected 'Ident', found "pkg"`},
		}

		for _, tc := range testCases {
			checkParse(t, tc.src, tc.wantErr)
		}
	})

	t.Run("const", func(t *testing.T) {
		testCases := []testCase{
			{`const a = b;`, ``},
			{`const a b = c;`, ``},

			{`const X any;`, `expected '=', found ';'`},
			{`const a`, `expected ';', found 'EOF'`},
			{`const a;`, `expected '=', found ';'`},
			{`const a 10;`, `expected '=', found 10`},
			{`const a b c;`, `expected '=', found c`},
		}

		for _, tc := range testCases {
			checkParse(t, pkgPrefix+tc.src, tc.wantErr)
		}
	})

	t.Run("func", func(t *testing.T) {
		testCases := []testCase{
			{"func f() { } ;", ""},
			{"func foo() string {}", ""},
			{`func fun() (foo,bar) {}`, ``},
			{`func (R) foo(){}`, ``},
			{`func f() {};`, ``},

			{"func f()\n{};", `unexpected semicolon or newline before {`},
			{"func f()\nfoo", `expected '{', found foo`},
		}

		for _, tc := range testCases {
			checkParse(t, pkgPrefix+tc.src, tc.wantErr)
		}
	})

	t.Run("import", func(t *testing.T) {
		testCases := []testCase{
			{`import "a"`, ``},
			{`import "foo"`, ``},
			{`import foo "bar"`, ``},
			{`import _ "bar"`, ``},
			{`import . "baz"`, ``},

			{`import _ ;`, `missing import path`},
			{`import baz`, `missing import path`},
			{`import _ baz`, `import path must be a string`},
			{
				`import "bar"; var _ a = a; import "baz"`,
				`imports must appear before other declarations`,
			},
		}

		for _, tc := range testCases {
			checkParse(t, pkgPrefix+tc.src, tc.wantErr)
		}
	})

	t.Run("struct", func(t *testing.T) {
		testCases := []testCase{
			{`struct foo{}`, ``},
			{`struct _{}`, ``},
			{`struct _{ A int }`, ``},

			{`struct foo bar{}`, `expected '{', found bar`},
		}

		for _, tc := range testCases {
			checkParse(t, pkgPrefix+tc.src, tc.wantErr)
		}
	})

	t.Run("typedef", func(t *testing.T) {
		testCases := []testCase{
			{`typedef foo bar`, ``},
			{`typedef foo = bar`, ``},
			{`typedef T = int`, ``},
		}

		for _, tc := range testCases {
			checkParse(t, pkgPrefix+tc.src, tc.wantErr)
		}
	})

	t.Run("var", func(t *testing.T) {
		testCases := []testCase{
			{`var a = b;`, ``},
			{`var a b = c;`, ``},
			{`var a bool = empty;`, ``},
		}

		for _, tc := range testCases {
			checkParse(t, pkgPrefix+tc.src, tc.wantErr)
		}
	})
}

func checkParse(t testing.TB, src, wantErr string) {
	fset := token.NewFileSet()
	_, err := ParseFile(fset, "", src)
	if err == nil && wantErr == "" {
		return
	}

	found := err.(ErrorList)
	found.removeMultiples()

	switch have := found.Error(); {
	case wantErr == "":
		t.Errorf("%s: unmatched error:\n%s\n", src, have)
	case !strings.HasSuffix(have, wantErr):
		t.Errorf("%s: error mismatch:\nhave: %s\nwant: %s\n", src, have, wantErr)
	}
}
