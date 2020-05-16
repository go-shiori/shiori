# JS [![GoDoc](http://godoc.org/github.com/tdewolff/parse/js?status.svg)](http://godoc.org/github.com/tdewolff/parse/js) [![GoCover](http://gocover.io/_badge/github.com/tdewolff/parse/js)](http://gocover.io/github.com/tdewolff/parse/js)

This package is a JS lexer (ECMA-262, edition 6.0) written in [Go][1]. It follows the specification at [ECMAScript Language Specification](http://www.ecma-international.org/ecma-262/6.0/). The lexer takes an io.Reader and converts it into tokens until the EOF.

## Installation
Run the following command

	go get github.com/tdewolff/parse/js

or add the following import and run project with `go get`

	import "github.com/tdewolff/parse/js"

## Lexer
### Usage
The following initializes a new Lexer with io.Reader `r`:
``` go
l := js.NewLexer(r)
```

To tokenize until EOF an error, use:
``` go
for {
	tt, text := l.Next()
	switch tt {
	case js.ErrorToken:
		// error or EOF set in l.Err()
		return
	// ...
	}
}
```

All tokens (see [ECMAScript Language Specification](http://www.ecma-international.org/ecma-262/6.0/)):
``` go
ErrorToken          TokenType = iota // extra token when errors occur
UnknownToken                         // extra token when no token can be matched
WhitespaceToken                      // space \t \v \f
LineTerminatorToken                  // \r \n \r\n
CommentToken
IdentifierToken // also: null true false
PunctuatorToken /* { } ( ) [ ] . ; , < > <= >= == != === !==  + - * % ++ -- << >>
   >>> & | ^ ! ~ && || ? : = += -= *= %= <<= >>= >>>= &= |= ^= / /= => */
NumericToken
StringToken
RegexpToken
TemplateToken
```

### Quirks
Because the ECMAScript specification for `PunctuatorToken` (of which the `/` and `/=` symbols) and `RegexpToken` depends on a parser state to differentiate between the two, the lexer (to remain modular) uses different rules. It aims to correctly disambiguate contexts and returns `RegexpToken` or `PunctuatorToken` where appropriate with only few exceptions which don't make much sense in runtime and so don't happen in a real-world code: function literal division (`x = function y(){} / z`) and object literal division (`x = {y:1} / z`).

Another interesting case introduced by ES2015 is `yield` operator in function generators vs `yield` as an identifier in regular functions. This was done for backward compatibility, but is very hard to disambiguate correctly on a lexer level without essentially implementing entire parsing spec as a state machine and hurting performance, code readability and maintainability, so, instead, `yield` is just always assumed to be an operator. In combination with above paragraph, this means that, for example, `yield /x/i` will be always parsed as `yield`-ing regular expression and not as `yield` identifier divided by `x` and then `i`. There is no evidence though that this pattern occurs in any popular libraries.

### Examples
``` go
package main

import (
	"os"

	"github.com/tdewolff/parse/js"
)

// Tokenize JS from stdin.
func main() {
	l := js.NewLexer(os.Stdin)
	for {
		tt, text := l.Next()
		switch tt {
		case js.ErrorToken:
			if l.Err() != io.EOF {
				fmt.Println("Error on line", l.Line(), ":", l.Err())
			}
			return
		case js.IdentifierToken:
			fmt.Println("Identifier", string(text))
		case js.NumericToken:
			fmt.Println("Numeric", string(text))
		// ...
		}
	}
}
```

## License
Released under the [MIT license](https://github.com/tdewolff/parse/blob/master/LICENSE.md).

[1]: http://golang.org/ "Go Language"
