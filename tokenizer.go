package jsongo

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

type JsonTokenType uint8

const (
	JTT_NONE JsonTokenType = iota
	JTT_STRING
	JTT_INTEGER
	JTT_FLOAT
	JTT_BOOLEAN
	JTT_NULL
	JTT_LBRACE
	JTT_RBRACE
	// JTT_LBRACKET
	// JTT_RBRACKET
	JTT_COMMA
	JTT_COLON
	JTT_EOF
)

func (jtt JsonTokenType) String() string {
	switch jtt {
	case JTT_NONE:
		return "JTT_NONE"
	case JTT_STRING:
		return "JTT_STRING"
	case JTT_INTEGER:
		return "JTT_INTEGER"
	case JTT_FLOAT:
		return "JTT_FLOAT"
	case JTT_BOOLEAN:
		return "JTT_BOOLEAN"
	case JTT_NULL:
		return "JTT_NULL"
	case JTT_LBRACE:
		return "JTT_LBRACE"
	case JTT_RBRACE:
		return "JTT_RBRACE"
	// case JTT_LBRACKET:
	// 	return "JTT_LBRACKET"
	// case JTT_RBRACKET:
	// 	return "JTT_RBRACKET"
	case JTT_COMMA:
		return "JTT_COMMA"
	case JTT_COLON:
		return "JTT_COLON"
	case JTT_EOF:
		return "JTT_EOF"
	default:
		return "UNKNOWN"
	}
}

type Location struct {
	Line   int
	Column int
}

func (l Location) String() string {
	return fmt.Sprintf("<%d:%d>", l.Line, l.Column)
}

type JsonToken struct {
	Type     JsonTokenType
	Value    string
	Location Location
}

func (jt *JsonToken) String() string {
	return fmt.Sprintf("[%-15s] ~ %-60s %s", jt.Type.String(), jt.Value, jt.Location.String())
}

type JsonTokenizer struct {
	FilePath      string
	FileContent   string
	ContentLength int
	Cursor        int
	Line          int
	Column        int
	CurrentToken  *JsonToken
}

func NewJsonTokenizer(filePath string) (*JsonTokenizer, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return &JsonTokenizer{
		FilePath:      filePath,
		FileContent:   string(bytes),
		ContentLength: len(bytes),
		Cursor:        0,
		Line:          1,
		Column:        1,
	}, nil
}

func (jt *JsonTokenizer) validCursor() bool {
	return jt.Cursor < jt.ContentLength
}

func (jt *JsonTokenizer) consumeChar() (byte, bool) {
	if jt.validCursor() {
		c := jt.FileContent[jt.Cursor]
		jt.Cursor++
		jt.Column++
		if c == '\n' {
			jt.Line++
			jt.Column = 1
		}
		return c, true
	}

	return 0, false
}

func (jt *JsonTokenizer) peekChar() (byte, bool) {
	if jt.validCursor() {
		return jt.FileContent[jt.Cursor], true
	}

	return 0, false
}

func (jt *JsonTokenizer) trimLeft() {
	c, ok := jt.peekChar()
	for ok && unicode.IsSpace(rune(c)) {
		jt.consumeChar()
		c, ok = jt.peekChar()
	}
}

func (jt *JsonTokenizer) consumeString() (string, bool) {
	// consume and check first char is '"'
	c, ok := jt.consumeChar()
	if !ok {
		return "", false
	}
	if c != '"' {
		return "", false
	}

	// consume until next '"'
	c, ok = jt.consumeChar()
	var str []byte
	for ok && c != '"' {
		str = append(str, c)
		c, ok = jt.consumeChar()
	}

	return string(str), true
}

func isValidNumberStart(c byte) bool {
	return unicode.IsDigit(rune(c))
}

func isValidNumber(c byte) bool {
	return unicode.IsDigit(rune(c)) || c == '.'
}

func (jt *JsonTokenizer) consumeNumber() (string, bool, bool) { // value, isFloat, ok
	c, ok := jt.peekChar()
	if !ok {
		return "", false, false
	}

	var str []byte
	isFloat := false
	for ok && isValidNumber(c) {
		if c == '.' {
			if isFloat {
				return "", false, false
			}
			isFloat = true
		}
		str = append(str, c)
		jt.consumeChar()
		c, ok = jt.peekChar()
	}

	return string(str), isFloat, true
}

func (jt *JsonTokenizer) consumeBool() (string, bool) {
	c, ok := jt.peekChar()
	if !ok {
		return "", false
	}

	var str []byte
	for ok && unicode.IsLetter(rune(c)) {
		str = append(str, c)
		jt.consumeChar()
		c, ok = jt.peekChar()
	}

	_str := strings.ToLower(string(str))

	if _str == "true" || _str == "false" {
		return string(str), true
	}

	return "", false
}

func (jt *JsonTokenizer) consumeNull() (string, bool) {
	c, ok := jt.peekChar()
	if !ok {
		return "", false
	}

	var str []byte
	for ok && unicode.IsLetter(rune(c)) {
		str = append(str, c)
		jt.consumeChar()
		c, ok = jt.peekChar()
	}

	_str := strings.ToLower(string(str))

	if _str == "null" {
		return string(str), true
	}

	return "", false
}

func (jt *JsonTokenizer) consumeSingleCharToken(char byte, token *JsonToken) {
	jt.consumeChar()
	token.Value = string(char)
	switch char {
	case '{':
		token.Type = JTT_LBRACE
	case '}':
		token.Type = JTT_RBRACE
	// case '[':
	// 	token.Type = JTT_LBRACKET
	// case ']':
	// 	token.Type = JTT_RBRACKET
	case ',':
		token.Type = JTT_COMMA
	case ':':
		token.Type = JTT_COLON
	}
}

func (jt *JsonTokenizer) NextToken() (*JsonToken, error) {
	jt.trimLeft()
	token := &JsonToken{
		Location: Location{
			Line:   jt.Line,
			Column: jt.Column,
		},
	}
	jt.CurrentToken = token

	c, ok := jt.peekChar()

	if !ok {
		token.Type = JTT_EOF
		return token, nil
	}

	switch c {
	case '"':
		token.Type = JTT_STRING
		token.Value, ok = jt.consumeString()
		token.Location.Column++
		if !ok {
			return nil, fmt.Errorf("invalid string at line %d, column %d", jt.Line, jt.Column)
		}
	case '{', '}', '[', ']', ',', ':':
		jt.consumeSingleCharToken(c, token)
	default:
		if isValidNumberStart(c) {
			consumed, isFloat, ok := jt.consumeNumber()
			if !ok {
				return nil, fmt.Errorf("invalid Token at line %d, column %d", jt.Line, jt.Column)
			}
			token.Value = consumed
			if isFloat {
				token.Type = JTT_FLOAT
			} else {
				token.Type = JTT_INTEGER
			}
		} else if c == 't' || c == 'f' {
			consumed, ok := jt.consumeBool()
			if !ok {
				return nil, fmt.Errorf("invalid Token at line %d, column %d", jt.Line, jt.Column)
			}
			token.Value = consumed
			token.Type = JTT_BOOLEAN
		} else if c == 'n' {
			consumed, ok := jt.consumeNull()
			if !ok {
				return nil, fmt.Errorf("invalid null at line %d, column %d", jt.Line, jt.Column)
			}
			token.Value = consumed
			token.Type = JTT_NULL
		} else {
			return nil, fmt.Errorf("invalid token at line %d, column %d : [%c]", jt.Line, jt.Column, c)
		}
	}

	return token, nil
}
