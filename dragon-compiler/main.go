package main

import (
	"lexer"
	"simple_parser"
)

func main() {
	source := `{int x; int y; int z;
		        x = 1;
				y = 2;
				if (y > x) {
					z = 2;
				}
				z = 3;
	}`
	my_lexer := lexer.NewLexer(source)
	parser := simple_parser.NewSimpleParser(my_lexer)
	parser.Parse()

}
