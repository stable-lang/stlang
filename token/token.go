// Package token defines lexical tokens of the Stable programming language.
package token

import (
	"strconv"
)

// Token is the set of lexical tokens of the Stable programming language.
type Token int

const (
	// Special tokens
	Illegal Token = iota
	EOF
	Comment

	// Identifiers and basic type literals
	literalA
	Ident  // main
	Int    // 12345
	Float  // 123.45
	Char   // 'a'
	String // "abc"
	literalZ

	// Operators and delimiters
	operatorA
	Add    // +
	Sub    // -
	Mul    // *
	Quo    // /
	Rem    // %
	And    // &
	Or     // |
	Xor    // ^
	AndNot // &^
	Shl    // <<
	Shr    // >>
	Concat // ++

	Assign       // =
	AddAssign    // +=
	SubAssign    // -=
	MulAssign    // *=
	QuoAssign    // /=
	RemAssign    // %=
	AndAssign    // &=
	OrAssign     // |=
	XorAssign    // ^=
	AndNotAssign // &^=
	ShlAssign    // <<=
	ShrAssign    // >>=
	ConcatAssign // ++=

	LogicAnd     // &&
	LogicOr      // ||
	LogicNot     // !
	Equal        // ==
	NotEqual     // !=
	Less         // <
	Greater      // >
	LessEqual    // <=
	GreaterEqual // >=

	Define    // :=
	Ellipsis  // ...
	Comma     // ,
	Period    // .
	Colon     // :
	Semicolon // ;

	LeftParen  // (
	LeftBrack  // [
	LeftBrace  // {
	RightParen // )
	RightBrack // ]
	RightBrace // }
	operatorZ

	// Keywords
	keywordA
	Any
	Bool
	Break
	Case
	Const
	Continue
	Defer
	Else
	Fallthrough
	False
	For
	Func
	Goto
	If
	Import
	Nil
	Package
	Return
	Struct
	Switch
	True
	Typedef
	Var
	Void
	keywordZ

	tokenMax
)

var tokens = [...]string{
	Illegal: "ILLEGAL",
	EOF:     "EOF",
	Comment: "COMMENT",

	Ident:  "IDENT",
	Int:    "INT",
	Float:  "FLOAT",
	Char:   "CHAR",
	String: "STRING",

	Add:    "+",
	Sub:    "-",
	Mul:    "*",
	Quo:    "/",
	Rem:    "%",
	And:    "&",
	Or:     "|",
	Xor:    "^",
	AndNot: "&^",
	Shl:    "<<",
	Shr:    ">>",
	Concat: "++",

	Assign:       "=",
	AddAssign:    "+=",
	SubAssign:    "-=",
	MulAssign:    "*=",
	QuoAssign:    "/=",
	RemAssign:    "%=",
	AndAssign:    "&=",
	OrAssign:     "|=",
	XorAssign:    "^=",
	AndNotAssign: "&^=",
	ShlAssign:    "<<=",
	ShrAssign:    ">>=",
	ConcatAssign: "++=",

	LogicAnd:     "&&",
	LogicOr:      "||",
	LogicNot:     "!",
	Equal:        "==",
	NotEqual:     "!=",
	Less:         "<",
	Greater:      ">",
	LessEqual:    "<=",
	GreaterEqual: ">=",

	Define:    ":=",
	Ellipsis:  "...",
	Comma:     ",",
	Period:    ".",
	Colon:     ":",
	Semicolon: ";",

	LeftParen:  "(",
	LeftBrack:  "[",
	LeftBrace:  "{",
	RightParen: ")",
	RightBrack: "]",
	RightBrace: "}",

	Any:         "any",
	Bool:        "bool",
	Break:       "break",
	Case:        "case",
	Const:       "const",
	Continue:    "continue",
	Defer:       "defer",
	Else:        "else",
	Fallthrough: "fallthrough",
	False:       "false",
	For:         "for",
	Func:        "func",
	Goto:        "goto",
	If:          "if",
	Import:      "import",
	Nil:         "nil",
	Package:     "package",
	Return:      "return",
	Struct:      "struct",
	Switch:      "switch",
	True:        "true",
	Typedef:     "typedef",
	Var:         "var",
	Void:        "void",
}

// String returns the string representation of the token.
func (tok Token) String() string {
	if 0 <= tok && tok < Token(len(tokens)) {
		return tokens[tok]
	}
	return "Token(" + strconv.Itoa(int(tok)) + ")"
}

// A set of constants for precedence-based expression parsing.
// Non-operators have lowest precedence, followed by operators
// starting with precedence 1 up to unary operators.
// The highest precedence serves as "catch-all" precedence for selector,
// indexing, and other operator and delimiter tokens.
const (
	LowestPrec  = 0 // non-operators
	UnaryPrec   = 6
	HighestPrec = 7
)

// Precedence returns the operator precedence of the binary operator.
// If operator is not a binary operator, the result is LowestPrecedence.
func (tok Token) Precedence() int {
	switch tok {
	case LogicOr:
		return 1
	case LogicAnd:
		return 2
	case Equal, NotEqual, Less, Greater, LessEqual, GreaterEqual:
		return 3
	case Add, Sub, Or, Xor:
		return 4
	case Mul, Quo, Rem, Shl, Shr, And, AndNot, Concat:
		return 5
	default:
		return LowestPrec
	}
}

// IsLiteral reports whether token corresponding to identifiers and basic type literals.
func (tok Token) IsLiteral() bool {
	return literalA < tok && tok < literalZ
}

// IsOperator reports whether token corresponding to operators and delimiters.
func (tok Token) IsOperator() bool {
	return operatorA < tok && tok < operatorZ
}

// IsKeyword reports whether token corresponding to keywords.
func (tok Token) IsKeyword() bool {
	return keywordA < tok && tok < keywordZ
}

// IsExported reports whether name starts with an upper-case letter.
func IsExported(name string) bool {
	return name != "" && ('A' <= name[0] && name[0] <= 'Z')
}

// IsIdentifier reports whether name is a Stlang identifier.
// Identifier is:
// - a non-empty string made up of letters, digits, and underscores,
// - where the first character is not a digit,
// - keywords are not identifiers.
func IsIdentifier(name string) bool {
	if name == "" || IsKeyword(name) {
		return false
	}

	for i, c := range name {
		switch {
		case 'a' <= c && c <= 'z':
		case 'A' <= c && c <= 'Z':
		case (i != 0 && '0' <= c && c <= '9'):
		case c == '_':
		default:
			return false
		}
	}
	return true
}

// IsKeyword reports whether name is a Stable keyword, such as "func" or "return".
func IsKeyword(name string) bool {
	_, ok := keywords[name]
	return ok
}

// Lookup an identifier to its keyword token or [Ident] (if not a keyword).
func Lookup(ident string) Token {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return Ident
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token, keywordZ-(keywordA+1))
	for i := Any + 1; i < Void; i++ {
		keywords[tokens[i]] = i
	}
}
