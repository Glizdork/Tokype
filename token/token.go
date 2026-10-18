package token

type TokenType string

const (
	EOF     TokenType = "EOF"
	ILLEGAL TokenType = "ILLEGAL"
	IDENT   TokenType = "IDENT"
	INT     TokenType = "INT"
	STRING  TokenType = "STRING"
	ASSIGN  TokenType = "="
	FLOAT   TokenType = "FLOAT"
	NULL    TokenType = "NULL"

	OPENLIST    TokenType = "["
	RBRACKET    TokenType = "]"
	OPENMAP     TokenType = "{"
	CLOSEMAP    TokenType = "}"
	DOTLBRACKET TokenType = ".["
	DOTDOT      TokenType = ".."
	DOT         TokenType = "."

	PLUS     TokenType = "+"
	MINUS    TokenType = "-"
	ASTERISK TokenType = "*"
	SLASH    TokenType = "/"
	MODULO   TokenType = "%"

	EQ     TokenType = "=="
	NOT_EQ TokenType = "?="
	LT     TokenType = "<"
	GT     TokenType = ">"
	LTE    TokenType = "<="
	GTE    TokenType = ">="

	COLON     TokenType = ":"
	LPAREN    TokenType = "("
	RPAREN    TokenType = ")"
	COMMA     TokenType = ","
	MAPEQ     TokenType = "::"
	SEMICOLON TokenType = ";"

	FUNCT  TokenType = "FUNCT"
	END    TokenType = "END"
	TRUE   TokenType = "TRUE"
	FALSE  TokenType = "FALSE"
	IF     TokenType = "IF"
	ELIF   TokenType = "ELIF"
	ELSE   TokenType = "ELSE"
	FOR    TokenType = "FOR"
	WHILE  TokenType = "WHILE"
	TO     TokenType = "TO"
	STEP   TokenType = "STEP"
	RETURN TokenType = "RETURN"

	AS      TokenType = "AS"
	INCLUDE TokenType = "INCLUDE"
)

var keywords = map[string]TokenType{
	"funct":   FUNCT,
	"end":     END,
	"true":    TRUE,
	"false":   FALSE,
	"null":    NULL,
	"if":      IF,
	"elif":    ELIF,
	"else":    ELSE,
	"for":     FOR,
	"while":   WHILE,
	"to":      TO,
	"step":    STEP,
	"return":  RETURN,
	"as":      AS,
	"include": INCLUDE,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}
