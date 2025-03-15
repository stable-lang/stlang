package token

import "testing"

func TestTokenPredicate(t *testing.T) {
	for tok := Token(0); tok <= tokenMax; tok++ {
		wantLiteral := literalA < tok && tok < literalZ
		wantOperator := operatorA < tok && tok < operatorZ
		wantKeywors := keywordA < tok && tok < keywordZ

		haveLiteral := tok.IsLiteral()
		haveOperator := tok.IsOperator()
		haveKeyword := tok.IsKeyword()

		switch {
		case haveLiteral != wantLiteral:
			t.Errorf("unexpected literal result: %d / %q\n", int(tok), tok.String())
		case haveOperator != wantOperator:
			t.Errorf("unexpected operator result: %d / %q\n", int(tok), tok.String())
		case haveKeyword != wantKeywors:
			t.Errorf("unexpected keywors result: %d / %q\n", int(tok), tok.String())
		}
	}
}

func TestIsIdentifier(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want bool
	}{
		{"Empty", "", false},
		{"Space", " ", false},
		{"SpaceSuffix", "foo ", false},
		{"Number", "123", false},
		{"Keyword", "func", false},

		{"LettersASCII", "foo", true},
		{"MixedASCII", "_bar123", true},
		{"UppercaseKeyword", "Func", true},
		{"LettersUnicode", "fóö", false},
		{"Emojis", "🤔", false},
	}

	for _, test := range tests {
		have := IsIdentifier(test.str)
		if have != test.want {
			t.Errorf("IsIdentifier(%q) = %t, want %v", test.str, have, test.want)
		}
	}
}
