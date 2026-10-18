package lexer

import (
	"Tokype/evaluator"
	"Tokype/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 1,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	if l.ch == '/' && l.peekChar() == '/' {
		l.skipComment()
		return l.NextToken()
	}

	var tok token.Token
	tok.Line = l.line
	tok.Column = l.column - 1

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok.Type = token.EQ
			tok.Literal = string(ch) + string(l.ch)
		} else {
			tok.Type = token.ASSIGN
			tok.Literal = string(l.ch)
		}
	case '+':
		tok.Type = token.PLUS
		tok.Literal = string(l.ch)
	case '-':
		tok.Type = token.MINUS
		tok.Literal = string(l.ch)
	case '*':
		tok.Type = token.ASTERISK
		tok.Literal = string(l.ch)
	case '/':
		tok.Type = token.SLASH
		tok.Literal = string(l.ch)
	case '%':
		tok.Type = token.MODULO
		tok.Literal = string(l.ch)
	case ':':
		if l.peekChar() == ':' {
			l.readChar()
			tok.Type = token.MAPEQ
			tok.Literal = "::"
		} else {
			tok.Type = token.COLON
			tok.Literal = string(l.ch)
		}
	case '.':
		if l.peekChar() == '[' {
			l.readChar()
			tok.Type = token.DOTLBRACKET
			tok.Literal = ".["
		} else if l.peekChar() == '.' {
			l.readChar()
			tok.Type = token.DOTDOT
			tok.Literal = ".."
		} else if isDigit(l.peekChar()) {
			literal, isFloat := l.readNumber()
			if isFloat {
				tok.Type = token.FLOAT
			} else {
				tok.Type = token.INT
			}
			tok.Literal = literal
			return tok
		} else {
			tok.Type = token.DOT
			tok.Literal = string(l.ch)
		}
	case '(':
		tok.Type = token.LPAREN
		tok.Literal = string(l.ch)
	case ')':
		tok.Type = token.RPAREN
		tok.Literal = string(l.ch)
	case ',':
		tok.Type = token.COMMA
		tok.Literal = string(l.ch)
	case '[':
		tok.Type = token.OPENLIST
		tok.Literal = string(l.ch)
	case ']':
		tok.Type = token.RBRACKET
		tok.Literal = string(l.ch)
	case '{':
		tok.Type = token.OPENMAP
		tok.Literal = string(l.ch)
	case '}':
		tok.Type = token.CLOSEMAP
		tok.Literal = string(l.ch)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		return tok
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.LTE
			tok.Literal = "<="
		} else {
			tok.Type = token.LT
			tok.Literal = string(l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.GTE
			tok.Literal = ">="
		} else {
			tok.Type = token.GT
			tok.Literal = string(l.ch)
		}
	case '?':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = token.NOT_EQ
			tok.Literal = "?="
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = string(l.ch)
		}
	case 0:
		tok.Type = token.EOF
		tok.Literal = ""
		tok.Line = l.line
		tok.Column = l.column
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			literal, isFloat := l.readNumber()
			if isFloat {
				tok.Type = token.FLOAT
			} else {
				tok.Type = token.INT
			}
			tok.Literal = literal
			return tok
		} else {
			tok.Type = token.ILLEGAL
			tok.Literal = string(l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	result := l.input[position:l.position]
	l.readChar()
	return result
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return evaluator.Intern(l.input[position:l.position])
}

func (l *Lexer) readNumber() (string, bool) {
	isFloat := false
	position := l.position

	if l.ch == '.' {
		isFloat = true
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
		return l.input[position:l.position], isFloat
	}

	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return l.input[position:l.position], isFloat
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
		l.readChar()
	}
}
