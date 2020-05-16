// Package js is an ECMAScript5.1 lexer following the specifications at http://www.ecma-international.org/ecma-262/5.1/.
package js // import "github.com/tdewolff/parse/js"

import (
	"io"
	"strconv"
	"unicode"

	"github.com/tdewolff/parse/buffer"
)

var identifierStart = []*unicode.RangeTable{unicode.Lu, unicode.Ll, unicode.Lt, unicode.Lm, unicode.Lo, unicode.Nl, unicode.Other_ID_Start}
var identifierContinue = []*unicode.RangeTable{unicode.Lu, unicode.Ll, unicode.Lt, unicode.Lm, unicode.Lo, unicode.Nl, unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc, unicode.Other_ID_Continue}

////////////////////////////////////////////////////////////////

// TokenType determines the type of token, eg. a number or a semicolon.
type TokenType uint32

// TokenType values.
const (
	ErrorToken          TokenType = iota // extra token when errors occur
	UnknownToken                         // extra token when no token can be matched
	WhitespaceToken                      // space \t \v \f
	LineTerminatorToken                  // \r \n \r\n
	SingleLineCommentToken
	MultiLineCommentToken // token for comments with line terminators (not just any /*block*/)
	IdentifierToken
	PunctuatorToken /* { } ( ) [ ] . ; , < > <= >= == != === !==  + - * % ++ -- << >>
	   >>> & | ^ ! ~ && || ? : = += -= *= %= <<= >>= >>>= &= |= ^= / /= >= */
	NumericToken
	StringToken
	RegexpToken
	TemplateToken
)

// TokenState determines a state in which next token should be read
type TokenState uint32

// TokenState values
const (
	ExprState TokenState = iota
	StmtParensState
	SubscriptState
	PropNameState
)

// ParsingContext determines the context in which following token should be parsed.
// This affects parsing regular expressions and template literals.
type ParsingContext uint32

// ParsingContext values
const (
	GlobalContext ParsingContext = iota
	StmtParensContext
	ExprParensContext
	BracesContext
	TemplateContext
)

// String returns the string representation of a TokenType.
func (tt TokenType) String() string {
	switch tt {
	case ErrorToken:
		return "Error"
	case UnknownToken:
		return "Unknown"
	case WhitespaceToken:
		return "Whitespace"
	case LineTerminatorToken:
		return "LineTerminator"
	case SingleLineCommentToken:
		return "SingleLineComment"
	case MultiLineCommentToken:
		return "MultiLineComment"
	case IdentifierToken:
		return "Identifier"
	case PunctuatorToken:
		return "Punctuator"
	case NumericToken:
		return "Numeric"
	case StringToken:
		return "String"
	case RegexpToken:
		return "Regexp"
	case TemplateToken:
		return "Template"
	}
	return "Invalid(" + strconv.Itoa(int(tt)) + ")"
}

////////////////////////////////////////////////////////////////

// Lexer is the state for the lexer.
type Lexer struct {
	r         *buffer.Lexer
	stack     []ParsingContext
	state     TokenState
	emptyLine bool
}

// NewLexer returns a new Lexer for a given io.Reader.
func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		r:         buffer.NewLexer(r),
		stack:     make([]ParsingContext, 0, 16),
		state:     ExprState,
		emptyLine: true,
	}
}

func (l *Lexer) enterContext(context ParsingContext) {
	l.stack = append(l.stack, context)
}

func (l *Lexer) leaveContext() ParsingContext {
	ctx := GlobalContext
	if last := len(l.stack) - 1; last >= 0 {
		ctx, l.stack = l.stack[last], l.stack[:last]
	}
	return ctx
}

// Err returns the error encountered during lexing, this is often io.EOF but also other errors can be returned.
func (l *Lexer) Err() error {
	return l.r.Err()
}

// Restore restores the NULL byte at the end of the buffer.
func (l *Lexer) Restore() {
	l.r.Restore()
}

// Next returns the next Token. It returns ErrorToken when an error was encountered. Using Err() one can retrieve the error message.
func (l *Lexer) Next() (TokenType, []byte) {
	tt := UnknownToken
	c := l.r.Peek(0)
	switch c {
	case '(':
		if l.state == StmtParensState {
			l.enterContext(StmtParensContext)
		} else {
			l.enterContext(ExprParensContext)
		}
		l.state = ExprState
		l.r.Move(1)
		tt = PunctuatorToken
	case ')':
		if l.leaveContext() == StmtParensContext {
			l.state = ExprState
		} else {
			l.state = SubscriptState
		}
		l.r.Move(1)
		tt = PunctuatorToken
	case '{':
		l.enterContext(BracesContext)
		l.state = ExprState
		l.r.Move(1)
		tt = PunctuatorToken
	case '}':
		if l.leaveContext() == TemplateContext && l.consumeTemplateToken() {
			tt = TemplateToken
		} else {
			// will work incorrectly for objects or functions divided by something,
			// but that's an extremely rare case
			l.state = ExprState
			l.r.Move(1)
			tt = PunctuatorToken
		}
	case ']':
		l.state = SubscriptState
		l.r.Move(1)
		tt = PunctuatorToken
	case '[', ';', ',', '~', '?', ':':
		l.state = ExprState
		l.r.Move(1)
		tt = PunctuatorToken
	case '<', '>', '=', '!', '+', '-', '*', '%', '&', '|', '^':
		if l.consumeHTMLLikeCommentToken() {
			return SingleLineCommentToken, l.r.Shift()
		} else if l.consumeLongPunctuatorToken() {
			l.state = ExprState
			tt = PunctuatorToken
		}
	case '/':
		if tt = l.consumeCommentToken(); tt != UnknownToken {
			return tt, l.r.Shift()
		} else if l.state == ExprState && l.consumeRegexpToken() {
			l.state = SubscriptState
			tt = RegexpToken
		} else if l.consumeLongPunctuatorToken() {
			l.state = ExprState
			tt = PunctuatorToken
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
		if l.consumeNumericToken() {
			tt = NumericToken
			l.state = SubscriptState
		} else if c == '.' {
			l.state = PropNameState
			l.r.Move(1)
			tt = PunctuatorToken
		}
	case '\'', '"':
		if l.consumeStringToken() {
			l.state = SubscriptState
			tt = StringToken
		}
	case ' ', '\t', '\v', '\f':
		l.r.Move(1)
		for l.consumeWhitespace() {
		}
		return WhitespaceToken, l.r.Shift()
	case '\n', '\r':
		l.r.Move(1)
		for l.consumeLineTerminator() {
		}
		tt = LineTerminatorToken
	case '`':
		if l.consumeTemplateToken() {
			tt = TemplateToken
		}
	default:
		if l.consumeIdentifierToken() {
			tt = IdentifierToken
			if l.state != PropNameState {
				switch hash := ToHash(l.r.Lexeme()); hash {
				case 0, This, False, True, Null:
					l.state = SubscriptState
				case If, While, For, With:
					l.state = StmtParensState
				default:
					// This will include keywords that can't be followed by a regexp, but only
					// by a specified char (like `switch` or `try`), but we don't check for syntax
					// errors as we don't attempt to parse a full JS grammar when streaming
					l.state = ExprState
				}
			} else {
				l.state = SubscriptState
			}
		} else if c >= 0xC0 {
			if l.consumeWhitespace() {
				for l.consumeWhitespace() {
				}
				return WhitespaceToken, l.r.Shift()
			} else if l.consumeLineTerminator() {
				for l.consumeLineTerminator() {
				}
				tt = LineTerminatorToken
			}
		} else if l.Err() != nil {
			return ErrorToken, nil
		}
	}

	l.emptyLine = tt == LineTerminatorToken

	if tt == UnknownToken {
		_, n := l.r.PeekRune(0)
		l.r.Move(n)
	}
	return tt, l.r.Shift()
}

////////////////////////////////////////////////////////////////

/*
The following functions follow the specifications at http://www.ecma-international.org/ecma-262/5.1/
*/

func (l *Lexer) consumeWhitespace() bool {
	c := l.r.Peek(0)
	if c == ' ' || c == '\t' || c == '\v' || c == '\f' {
		l.r.Move(1)
		return true
	} else if c >= 0xC0 {
		if r, n := l.r.PeekRune(0); r == '\u00A0' || r == '\uFEFF' || unicode.Is(unicode.Zs, r) {
			l.r.Move(n)
			return true
		}
	}
	return false
}

func (l *Lexer) consumeLineTerminator() bool {
	c := l.r.Peek(0)
	if c == '\n' {
		l.r.Move(1)
		return true
	} else if c == '\r' {
		if l.r.Peek(1) == '\n' {
			l.r.Move(2)
		} else {
			l.r.Move(1)
		}
		return true
	} else if c >= 0xC0 {
		if r, n := l.r.PeekRune(0); r == '\u2028' || r == '\u2029' {
			l.r.Move(n)
			return true
		}
	}
	return false
}

func (l *Lexer) consumeDigit() bool {
	if c := l.r.Peek(0); c >= '0' && c <= '9' {
		l.r.Move(1)
		return true
	}
	return false
}

func (l *Lexer) consumeHexDigit() bool {
	if c := l.r.Peek(0); (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
		l.r.Move(1)
		return true
	}
	return false
}

func (l *Lexer) consumeBinaryDigit() bool {
	if c := l.r.Peek(0); c == '0' || c == '1' {
		l.r.Move(1)
		return true
	}
	return false
}

func (l *Lexer) consumeOctalDigit() bool {
	if c := l.r.Peek(0); c >= '0' && c <= '7' {
		l.r.Move(1)
		return true
	}
	return false
}

func (l *Lexer) consumeUnicodeEscape() bool {
	if l.r.Peek(0) != '\\' || l.r.Peek(1) != 'u' {
		return false
	}
	mark := l.r.Pos()
	l.r.Move(2)
	if c := l.r.Peek(0); c == '{' {
		l.r.Move(1)
		if l.consumeHexDigit() {
			for l.consumeHexDigit() {
			}
			if c := l.r.Peek(0); c == '}' {
				l.r.Move(1)
				return true
			}
		}
		l.r.Rewind(mark)
		return false
	} else if !l.consumeHexDigit() || !l.consumeHexDigit() || !l.consumeHexDigit() || !l.consumeHexDigit() {
		l.r.Rewind(mark)
		return false
	}
	return true
}

func (l *Lexer) consumeSingleLineComment() {
	for {
		c := l.r.Peek(0)
		if c == '\r' || c == '\n' || c == 0 {
			break
		} else if c >= 0xC0 {
			if r, _ := l.r.PeekRune(0); r == '\u2028' || r == '\u2029' {
				break
			}
		}
		l.r.Move(1)
	}
}

////////////////////////////////////////////////////////////////

func (l *Lexer) consumeHTMLLikeCommentToken() bool {
	c := l.r.Peek(0)
	if c == '<' && l.r.Peek(1) == '!' && l.r.Peek(2) == '-' && l.r.Peek(3) == '-' {
		// opening HTML-style single line comment
		l.r.Move(4)
		l.consumeSingleLineComment()
		return true
	} else if l.emptyLine && c == '-' && l.r.Peek(1) == '-' && l.r.Peek(2) == '>' {
		// closing HTML-style single line comment
		// (only if current line didn't contain any meaningful tokens)
		l.r.Move(3)
		l.consumeSingleLineComment()
		return true
	}
	return false
}

func (l *Lexer) consumeCommentToken() TokenType {
	c := l.r.Peek(0)
	if c == '/' {
		c = l.r.Peek(1)
		if c == '/' {
			// single line comment
			l.r.Move(2)
			l.consumeSingleLineComment()
			return SingleLineCommentToken
		} else if c == '*' {
			// block comment (potentially multiline)
			tt := SingleLineCommentToken
			l.r.Move(2)
			for {
				c := l.r.Peek(0)
				if c == '*' && l.r.Peek(1) == '/' {
					l.r.Move(2)
					break
				} else if c == 0 {
					break
				} else if l.consumeLineTerminator() {
					tt = MultiLineCommentToken
					l.emptyLine = true
				} else {
					l.r.Move(1)
				}
			}
			return tt
		}
	}
	return UnknownToken
}

func (l *Lexer) consumeLongPunctuatorToken() bool {
	c := l.r.Peek(0)
	if c == '!' || c == '=' || c == '+' || c == '-' || c == '*' || c == '/' || c == '%' || c == '&' || c == '|' || c == '^' {
		l.r.Move(1)
		if l.r.Peek(0) == '=' {
			l.r.Move(1)
			if (c == '!' || c == '=') && l.r.Peek(0) == '=' {
				l.r.Move(1)
			}
		} else if (c == '+' || c == '-' || c == '&' || c == '|') && l.r.Peek(0) == c {
			l.r.Move(1)
		} else if c == '=' && l.r.Peek(0) == '>' {
			l.r.Move(1)
		}
	} else { // c == '<' || c == '>'
		l.r.Move(1)
		if l.r.Peek(0) == c {
			l.r.Move(1)
			if c == '>' && l.r.Peek(0) == '>' {
				l.r.Move(1)
			}
		}
		if l.r.Peek(0) == '=' {
			l.r.Move(1)
		}
	}
	return true
}

func (l *Lexer) consumeIdentifierToken() bool {
	c := l.r.Peek(0)
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '$' || c == '_' {
		l.r.Move(1)
	} else if c >= 0xC0 {
		if r, n := l.r.PeekRune(0); unicode.IsOneOf(identifierStart, r) {
			l.r.Move(n)
		} else {
			return false
		}
	} else if !l.consumeUnicodeEscape() {
		return false
	}
	for {
		c := l.r.Peek(0)
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '$' || c == '_' {
			l.r.Move(1)
		} else if c >= 0xC0 {
			if r, n := l.r.PeekRune(0); r == '\u200C' || r == '\u200D' || unicode.IsOneOf(identifierContinue, r) {
				l.r.Move(n)
			} else {
				break
			}
		} else {
			break
		}
	}
	return true
}

func (l *Lexer) consumeNumericToken() bool {
	// assume to be on 0 1 2 3 4 5 6 7 8 9 .
	mark := l.r.Pos()
	c := l.r.Peek(0)
	if c == '0' {
		l.r.Move(1)
		if l.r.Peek(0) == 'x' || l.r.Peek(0) == 'X' {
			l.r.Move(1)
			if l.consumeHexDigit() {
				for l.consumeHexDigit() {
				}
			} else {
				l.r.Move(-1) // return just the zero
			}
			return true
		} else if l.r.Peek(0) == 'b' || l.r.Peek(0) == 'B' {
			l.r.Move(1)
			if l.consumeBinaryDigit() {
				for l.consumeBinaryDigit() {
				}
			} else {
				l.r.Move(-1) // return just the zero
			}
			return true
		} else if l.r.Peek(0) == 'o' || l.r.Peek(0) == 'O' {
			l.r.Move(1)
			if l.consumeOctalDigit() {
				for l.consumeOctalDigit() {
				}
			} else {
				l.r.Move(-1) // return just the zero
			}
			return true
		}
	} else if c != '.' {
		for l.consumeDigit() {
		}
	}
	if l.r.Peek(0) == '.' {
		l.r.Move(1)
		if l.consumeDigit() {
			for l.consumeDigit() {
			}
		} else if c != '.' {
			// . could belong to the next token
			l.r.Move(-1)
			return true
		} else {
			l.r.Rewind(mark)
			return false
		}
	}
	mark = l.r.Pos()
	c = l.r.Peek(0)
	if c == 'e' || c == 'E' {
		l.r.Move(1)
		c = l.r.Peek(0)
		if c == '+' || c == '-' {
			l.r.Move(1)
		}
		if !l.consumeDigit() {
			// e could belong to the next token
			l.r.Rewind(mark)
			return true
		}
		for l.consumeDigit() {
		}
	}
	return true
}

func (l *Lexer) consumeStringToken() bool {
	// assume to be on ' or "
	mark := l.r.Pos()
	delim := l.r.Peek(0)
	l.r.Move(1)
	for {
		c := l.r.Peek(0)
		if c == delim {
			l.r.Move(1)
			break
		} else if c == '\\' {
			l.r.Move(1)
			if !l.consumeLineTerminator() {
				if c := l.r.Peek(0); c == delim || c == '\\' {
					l.r.Move(1)
				}
			}
			continue
		} else if c == '\n' || c == '\r' {
			l.r.Rewind(mark)
			return false
		} else if c >= 0xC0 {
			if r, _ := l.r.PeekRune(0); r == '\u2028' || r == '\u2029' {
				l.r.Rewind(mark)
				return false
			}
		} else if c == 0 {
			break
		}
		l.r.Move(1)
	}
	return true
}

func (l *Lexer) consumeRegexpToken() bool {
	// assume to be on / and not /*
	mark := l.r.Pos()
	l.r.Move(1)
	inClass := false
	for {
		c := l.r.Peek(0)
		if !inClass && c == '/' {
			l.r.Move(1)
			break
		} else if c == '[' {
			inClass = true
		} else if c == ']' {
			inClass = false
		} else if c == '\\' {
			l.r.Move(1)
			if l.consumeLineTerminator() {
				l.r.Rewind(mark)
				return false
			} else if l.r.Peek(0) == 0 {
				return true
			}
		} else if l.consumeLineTerminator() {
			l.r.Rewind(mark)
			return false
		} else if c == 0 {
			return true
		}
		l.r.Move(1)
	}
	// flags
	for {
		c := l.r.Peek(0)
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '$' || c == '_' {
			l.r.Move(1)
		} else if c >= 0xC0 {
			if r, n := l.r.PeekRune(0); r == '\u200C' || r == '\u200D' || unicode.IsOneOf(identifierContinue, r) {
				l.r.Move(n)
			} else {
				break
			}
		} else {
			break
		}
	}
	return true
}

func (l *Lexer) consumeTemplateToken() bool {
	// assume to be on ` or } when already within template
	mark := l.r.Pos()
	l.r.Move(1)
	for {
		c := l.r.Peek(0)
		if c == '`' {
			l.state = SubscriptState
			l.r.Move(1)
			return true
		} else if c == '$' && l.r.Peek(1) == '{' {
			l.enterContext(TemplateContext)
			l.state = ExprState
			l.r.Move(2)
			return true
		} else if c == '\\' {
			l.r.Move(1)
			if c := l.r.Peek(0); c != 0 {
				l.r.Move(1)
			}
			continue
		} else if c == 0 {
			l.r.Rewind(mark)
			return false
		}
		l.r.Move(1)
	}
}
