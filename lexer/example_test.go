package lexer_test

import (
	"fmt"

	"github.com/stable-lang/stlang/lexer"
	"github.com/stable-lang/stlang/token"
)

func ExampleScanner() {
	source := `
package main
	
import "fmt"
	
/* Stable is a general-purpose programming language. */

var A = 10

func main() {
	println(A + 32)
}
`

	fileSet := token.NewFileSet()
	file := fileSet.AddFile("", fileSet.Base(), len(source))
	s := lexer.NewLexer(file, []byte(source), nil)

	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		fmt.Printf("%.10s\t%s\t%q\n", fileSet.Position(pos), tok, lit)
	}

	// Output:
	// 2:1	package	"package"
	// 2:9	IDENT	"main"
	// 2:13	;	"\n"
	// 4:1	import	"import"
	// 4:8	STRING	"\"fmt\""
	// 4:13	;	"\n"
	// 6:1	COMMENT	"/* Stable is a general-purpose programming language. */"
	// 8:1	var	"var"
	// 8:5	IDENT	"A"
	// 8:7	=	""
	// 8:9	INT	"10"
	// 8:11	;	"\n"
	// 10:1	func	"func"
	// 10:6	IDENT	"main"
	// 10:10	(	""
	// 10:11	)	""
	// 10:13	{	""
	// 11:2	IDENT	"println"
	// 11:9	(	""
	// 11:10	IDENT	"A"
	// 11:12	+	""
	// 11:14	INT	"32"
	// 11:16	)	""
	// 11:17	;	"\n"
	// 12:1	}	""
	// 12:2	;	"\n"
}
