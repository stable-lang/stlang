package token

import (
	"testing"
)

func TestNoPos(t *testing.T) {
	if NoPos.IsValid() {
		t.Errorf("NoPos should not be valid")
	}

	var fset *FileSet
	checkPos(t, "nil NoPos", fset.Position(NoPos), Position{})
	fset = NewFileSet()
	checkPos(t, "fset NoPos", fset.Position(NoPos), Position{})
}

func checkPos(t testing.TB, msg string, have, want Position) {
	t.Helper()

	if have.Filename != want.Filename {
		t.Errorf("%s: have filename = %q; want %q", msg, have.Filename, want.Filename)
	}
	if have.Offset != want.Offset {
		t.Errorf("%s: have offset = %d; want %d", msg, have.Offset, want.Offset)
	}
	if have.Line != want.Line {
		t.Errorf("%s: have line = %d; want %d", msg, have.Line, want.Line)
	}
	if have.Column != want.Column {
		t.Errorf("%s: have column = %d; want %d", msg, have.Column, want.Column)
	}
}
