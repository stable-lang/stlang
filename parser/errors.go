package parser

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/stable-lang/stlang/token"
)

// Error from [Parser] process.
type Error struct {
	Pos token.Position
	Msg string
}

// Error implements the error interface.
func (e Error) Error() string {
	if e.Pos.Filename != "" || e.Pos.IsValid() {
		return e.Pos.String() + ": " + e.Msg
	}
	return e.Msg
}

// ErrorList is a list of [Error].
type ErrorList []Error

// Error implements the error interface.
func (p ErrorList) Error() string {
	switch len(p) {
	case 0:
		return "no errors"
	case 1:
		return p[0].Error()
	default:
		return fmt.Sprintf("%s (and %d more errors)", p[0], len(p)-1)
	}
}

// Err returns an error equivalent to this error list.
// If the list is empty, Err returns nil.
func (p ErrorList) Err() error {
	if len(p) == 0 {
		return nil
	}
	return p
}

func (p ErrorList) Len() int { return len(p) }
func (p *ErrorList) Reset()  { *p = (*p)[0:0] }

// Add an [Error] with given position and error message.
func (p *ErrorList) Add(pos token.Position, msg string) {
	*p = append(*p, Error{
		Pos: pos,
		Msg: msg,
	})
}

// removeMultiples sorts an [ErrorList] and removes all but the first error per line.
func (p *ErrorList) removeMultiples() {
	p.sort()

	var last token.Position // initial last.Line is != any legal error line
	i := 0
	for _, e := range *p {
		if e.Pos.Filename != last.Filename || e.Pos.Line != last.Line {
			last = e.Pos
			(*p)[i] = e
			i++
		}
	}
	*p = (*p)[0:i]
}

func (p ErrorList) sort() {
	slices.SortFunc(p, func(ee, ff Error) int {
		e, f := ee.Pos, ff.Pos
		return cmp.Or(
			strings.Compare(e.Filename, f.Filename),
			cmp.Compare(e.Line, f.Line),
			cmp.Compare(e.Column, f.Column),
			strings.Compare(ee.Msg, ff.Msg),
		)
	})
}
