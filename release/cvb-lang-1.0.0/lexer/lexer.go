package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType string

const (
	// 关键字
	TOKEN_IMPORT   TokenType = "IMPORT"
	TOKEN_IMPORTF  TokenType = "IMPORTF"
	TOKEN_PRINT    TokenType = "PRINT"
	TOKEN_WHILE    TokenType = "WHILE"
	TOKEN_FOR      TokenType = "FOR"
	TOKEN_IF       TokenType = "IF"
	TOKEN_ELSE     TokenType = "ELSE"
	TOKEN_DEF      TokenType = "DEF"
	TOKEN_BREAK    TokenType = "BREAK"
	TOKEN_IN       TokenType = "IN"
	TOKEN_AND      TokenType = "AND"
	TOKEN_OR       TokenType = "OR"
	TOKEN_TRUE     TokenType = "TRUE"
	TOKEN_FALSE    TokenType = "FALSE"

	// 类型
	TOKEN_STR   TokenType = "STR"
	TOKEN_INT   TokenType = "INT"
	TOKEN_LIST  TokenType = "LIST"
	TOKEN_DIC   TokenType = "DIC"
	TOKEN_MATH  TokenType = "MATH"
	TOKEN_FILE  TokenType = "FILE"
	TOKEN_RANDOM TokenType = "RANDOM"
	TOKEN_NET   TokenType = "NET"
	TOKEN_SHELL TokenType = "SHELL"

	// 字面量
	TOKEN_STRING  TokenType = "STRING"
	TOKEN_NUMBER  TokenType = "NUMBER"
	TOKEN_IDENTIFIER TokenType = "IDENTIFIER"

	// 运算符
	TOKEN_ARROW     TokenType = "ARROW"     // =>
	TOKEN_ASSIGN    TokenType = "ASSIGN"    // =
	TOKEN_EQ        TokenType = "EQ"        // ==
	TOKEN_LT        TokenType = "LT"        // <
	TOKEN_GT        TokenType = "GT"        // >
	TOKEN_LE        TokenType = "LE"        // <=
	TOKEN_GE        TokenType = "GE"        // >=
	TOKEN_NE        TokenType = "NE"        // !=
	TOKEN_PLUS      TokenType = "PLUS"      // +
	TOKEN_MINUS     TokenType = "MINUS"     // -
	TOKEN_MUL       TokenType = "MUL"       // *
	TOKEN_DIV       TokenType = "DIV"       // /
	TOKEN_MOD       TokenType = "MOD"       // %
	TOKEN_AND_OP    TokenType = "AND_OP"    // &&
	TOKEN_OR_OP     TokenType = "OR_OP"     // ||

	// 分隔符
	TOKEN_LPAREN    TokenType = "LPAREN"    // (
	TOKEN_RPAREN    TokenType = "RPAREN"    // )
	TOKEN_LBRACE    TokenType = "LBRACE"    // {
	TOKEN_RBRACE    TokenType = "RBRACE"    // }
	TOKEN_LBRACKET  TokenType = "LBRACKET"  // [
	TOKEN_RBRACKET  TokenType = "RBRACKET"  // ]
	TOKEN_COLON     TokenType = "COLON"     // :
	TOKEN_SEMICOLON TokenType = "SEMICOLON" // ;
	TOKEN_COMMA     TokenType = "COMMA"     // ,
	TOKEN_DOT       TokenType = "DOT"       // .
	TOKEN_PIPE      TokenType = "PIPE"      // |
	TOKEN_AMPERSAND TokenType = "AMPERSAND" // &
	TOKEN_DOLLAR    TokenType = "DOLLAR"    // $
	TOKEN_HASH      TokenType = "HASH"      // #
	TOKEN_AT        TokenType = "AT"        // @

	// 特殊
	TOKEN_NEWLINE   TokenType = "NEWLINE"
	TOKEN_EOF       TokenType = "EOF"
	TOKEN_ILLEGAL   TokenType = "ILLEGAL"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
	line    int
	column  int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	l.column++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	if l.ch == '(' && l.peekChar() == '#' {
		for l.ch != 0 && !(l.ch == '#' && l.peekChar() == ')') {
			l.readChar()
		}
		if l.ch == '#' && l.peekChar() == ')' {
			l.readChar()
			l.readChar()
		}
	}
}

func (l *Lexer) readString() string {
	quote := l.ch
	l.readChar()
	var sb strings.Builder
	for l.ch != quote && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			default:
				sb.WriteByte(l.ch)
			}
		} else {
			sb.WriteByte(l.ch)
		}
		l.readChar()
	}
	if l.ch == quote {
		l.readChar()
	}
	return sb.String()
}

func (l *Lexer) readNumber() string {
	var sb strings.Builder
	for isDigit(l.ch) {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		sb.WriteByte(l.ch)
		l.readChar()
		for isDigit(l.ch) {
			sb.WriteByte(l.ch)
			l.readChar()
		}
	}
	return sb.String()
}

func (l *Lexer) readIdentifier() string {
	var sb strings.Builder
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	return sb.String()
}

func (l *Lexer) NextToken() Token {
	var tok Token
	l.skipWhitespace()
	l.skipComment()
	l.skipWhitespace()

	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case '#':
		l.readChar()
		ident := l.readIdentifier()
		switch ident {
		case "import":
			tok = l.makeToken(TOKEN_IMPORT, "#import")
		case "importf":
			tok = l.makeToken(TOKEN_IMPORTF, "#importf")
		case "print":
			tok = l.makeToken(TOKEN_PRINT, "#print")
		case "while":
			tok = l.makeToken(TOKEN_WHILE, "#while")
		case "for":
			tok = l.makeToken(TOKEN_FOR, "#for")
		case "if":
			tok = l.makeToken(TOKEN_IF, "#if")
		case "else":
			tok = l.makeToken(TOKEN_ELSE, "#else")
		case "def":
			tok = l.makeToken(TOKEN_DEF, "#def")
		case "break":
			tok = l.makeToken(TOKEN_BREAK, "#break")
		default:
			tok = l.makeToken(TOKEN_ILLEGAL, "#"+ident)
		}
	case '=':
		if l.peekChar() == '>' {
			l.readChar()
			tok = l.makeToken(TOKEN_ARROW, "=>")
			l.readChar()
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(TOKEN_EQ, "==")
			l.readChar()
		} else {
			tok = l.makeToken(TOKEN_ASSIGN, "=")
			l.readChar()
		}
		return tok
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(TOKEN_LE, "<=")
			l.readChar()
		} else {
			tok = l.makeToken(TOKEN_LT, "<")
			l.readChar()
		}
		return tok
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(TOKEN_GE, ">=")
			l.readChar()
		} else {
			tok = l.makeToken(TOKEN_GT, ">")
			l.readChar()
		}
		return tok
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(TOKEN_NE, "!=")
			l.readChar()
		} else {
			tok = l.makeToken(TOKEN_ILLEGAL, string(l.ch))
			l.readChar()
		}
		return tok
	case '+':
		tok = l.makeToken(TOKEN_PLUS, "+")
		l.readChar()
	case '-':
		tok = l.makeToken(TOKEN_MINUS, "-")
		l.readChar()
	case '*':
		tok = l.makeToken(TOKEN_MUL, "*")
		l.readChar()
	case '/':
		tok = l.makeToken(TOKEN_DIV, "/")
		l.readChar()
	case '%':
		tok = l.makeToken(TOKEN_MOD, "%")
		l.readChar()
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = l.makeToken(TOKEN_AND_OP, "&&")
			l.readChar()
		} else {
			tok = l.makeToken(TOKEN_AMPERSAND, "&")
			l.readChar()
		}
		return tok
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = l.makeToken(TOKEN_OR_OP, "||")
			l.readChar()
		} else {
			tok = l.makeToken(TOKEN_PIPE, "|")
			l.readChar()
		}
		return tok
	case '(':
		tok = l.makeToken(TOKEN_LPAREN, "(")
		l.readChar()
	case ')':
		tok = l.makeToken(TOKEN_RPAREN, ")")
		l.readChar()
	case '{':
		tok = l.makeToken(TOKEN_LBRACE, "{")
		l.readChar()
	case '}':
		tok = l.makeToken(TOKEN_RBRACE, "}")
		l.readChar()
	case '[':
		tok = l.makeToken(TOKEN_LBRACKET, "[")
		l.readChar()
	case ']':
		tok = l.makeToken(TOKEN_RBRACKET, "]")
		l.readChar()
	case ':':
		tok = l.makeToken(TOKEN_COLON, ":")
		l.readChar()
	case ';':
		tok = l.makeToken(TOKEN_SEMICOLON, ";")
		l.readChar()
	case ',':
		tok = l.makeToken(TOKEN_COMMA, ",")
		l.readChar()
	case '.':
		tok = l.makeToken(TOKEN_DOT, ".")
		l.readChar()
	case '$':
		tok = l.makeToken(TOKEN_DOLLAR, "$")
		l.readChar()
	case '@':
		tok = l.makeToken(TOKEN_AT, "@")
		l.readChar()
	case '\n':
		tok = l.makeToken(TOKEN_NEWLINE, "\n")
		l.readChar()
	case '"', '\'':
		str := l.readString()
		tok = l.makeToken(TOKEN_STRING, str)
	case 0:
		tok = l.makeToken(TOKEN_EOF, "")
	default:
		if isLetter(l.ch) {
			ident := l.readIdentifier()
			tokType := LookupIdent(ident)
			tok = l.makeToken(tokType, ident)
		} else if isDigit(l.ch) {
			num := l.readNumber()
			tok = l.makeToken(TOKEN_NUMBER, num)
		} else {
			tok = l.makeToken(TOKEN_ILLEGAL, string(l.ch))
			l.readChar()
		}
	}

	return tok
}

func (l *Lexer) makeToken(tokenType TokenType, literal string) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  l.column,
	}
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isDigit(ch byte) bool {
	return unicode.IsDigit(rune(ch))
}

func LookupIdent(ident string) TokenType {
	switch ident {
	case "str":
		return TOKEN_STR
	case "int":
		return TOKEN_INT
	case "list":
		return TOKEN_LIST
	case "dic":
		return TOKEN_DIC
	case "math":
		return TOKEN_MATH
	case "file":
		return TOKEN_FILE
	case "random":
		return TOKEN_RANDOM
	case "net":
		return TOKEN_NET
	case "shell":
		return TOKEN_SHELL
	case "in":
		return TOKEN_IN
	case "and":
		return TOKEN_AND
	case "or":
		return TOKEN_OR
	case "TRUE":
		return TOKEN_TRUE
	case "FALSE":
		return TOKEN_FALSE
	case "if":
		return TOKEN_IF
	case "else":
		return TOKEN_ELSE
	default:
		return TOKEN_IDENTIFIER
	}
}

func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Literal: %q, Line: %d, Col: %d}",
		t.Type, t.Literal, t.Line, t.Column)
}
