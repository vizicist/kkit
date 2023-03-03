package kit

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Token struct {
	Typ tokenType // Type, such as itemNumber
	Val string    // Value, such as "23.2"
}

type tokenType int

type stateFn func(*Lexer) stateFn

const (
	ItemNothing tokenType = iota
	ItemError
	ItemEOF
	ItemText
	ItemNumber
	ItemPeriod
	ItemMinus
	ItemPercent
	ItemPlus
	ItemEquals
	ItemOr
	ItemAnd
	ItemBitwiseOr
	ItemBitwiseAnd
	ItemNotEquals
	ItemEqualsEquals
	ItemLessThan
	ItemGreaterThanOrEqual
	ItemLessThanOrEqual
	ItemGreaterThan
	ItemAsterisk
	ItemForwardSlash
	ItemNot
	ItemPound
	ItemInvalid
	ItemAt
	ItemIf
	ItemElse
	ItemDollar
	ItemComment
	ItemDoubleQuote
	ItemSingleQuote
	ItemBackQuote
	ItemComma
	ItemTilda
	ItemDoubleTilda
	ItemColon
	ItemSemiColon
	ItemNewline
	ItemReturn
	ItemQuestionMark
	ItemLeftParen
	ItemRightParen
	ItemLeftBrace
	ItemRightBrace
	ItemLeftBracket
	ItemRightBracket
	ItemIdentifier
)

type Lexer struct {
	name   string // used only for error reports
	input  string // the string being parsed
	start  int
	pos    int
	width  int
	items  chan Token
	lineno int
}

func (i Token) String() string {
	switch i.Typ {
	case ItemEOF:
		return "EOF"
	case ItemError:
		return i.Val
	}
	if len(i.Val) > 10 {
		return fmt.Sprintf("%.10q...", i.Val)
	}
	return fmt.Sprintf("%q", i.Val)
}

func Lex(name, input string) (*Lexer, chan Token) {
	l := &Lexer{
		name:  name,
		input: input,
		items: make(chan Token),
	}
	go l.Run()
	return l, l.items
}

func (l *Lexer) Run() {
	for state := lexNormal; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered
}

func (l *Lexer) emit(t tokenType) {
	val := l.input[l.start:l.pos]
	l.items <- Token{t, val}
	l.start = l.pos
}

const EOF = -1

const numberChars = "0123456789"
const identifierChar1 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
const identifierCharN = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"

func lexNormal(l *Lexer) stateFn {

	firstChar := l.next()
	if firstChar == EOF {
		l.emit(ItemEOF)
		return nil
	}
	switch firstChar {
	case '#':
		return lexComment
	case '$':
		l.emit(ItemDollar)
	case '(':
		l.emit(ItemLeftParen)
	case ')':
		l.emit(ItemRightParen)
	case '{':
		l.emit(ItemLeftBrace)
	case '}':
		l.emit(ItemRightBrace)
	case '[':
		l.emit(ItemLeftBracket)
	case ']':
		l.emit(ItemRightBracket)
	case '+':
		l.emit(ItemPlus)
	case '-':
		l.emit(ItemMinus)
	case '*':
		l.emit(ItemAsterisk)
	case '/':
		l.emit(ItemForwardSlash)
	case '%':
		l.emit(ItemPercent)
	case '\\', '@':
		return l.errorf("Unexpected char in lexNormal")
	case '.':
		l.emit(ItemPeriod)
	case ',':
		l.emit(ItemComma)
	case ':':
		l.emit(ItemColon)
	case ';':
		l.emit(ItemSemiColon)
	case '?':
		l.emit(ItemQuestionMark)
	case '!':
		if l.peek() == '=' {
			l.next()
			l.emit(ItemNotEquals)
		} else {
			l.emit(ItemNot)
		}
	case '<':
		if l.peek() == '=' {
			l.next()
			l.emit(ItemLessThanOrEqual)
		} else {
			l.emit(ItemLessThan)
		}
	case '>':
		if l.peek() == '=' {
			l.next()
			l.emit(ItemGreaterThanOrEqual)
		} else {
			l.emit(ItemGreaterThan)
		}
	case '=':
		if l.peek() == '=' {
			l.next()
			l.emit(ItemEqualsEquals)
		} else {
			l.emit(ItemEquals)
		}
	case '|':
		if l.peek() == '|' {
			l.next()
			l.emit(ItemOr)
		} else {
			l.emit(ItemBitwiseOr)
		}
	case '~':
		if l.peek() == '~' {
			l.next()
			l.emit(ItemDoubleTilda)
		} else {
			l.emit(ItemTilda)
		}
	case '&':
		if l.peek() == '&' {
			l.next()
			l.emit(ItemAnd)
		} else {
			l.emit(ItemBitwiseAnd)
		}
	case ' ', '\t':
		l.acceptRun(" \t")
		l.ignore()
		// Don't emit anything
	case '\n':
		l.lineno++
		l.ignore()
		// Don't emit anything
	case '\r':
		l.ignore()
		// Don't emit anything
	case '"':
		return lexDoubleQuote
	case '\'':
		return lexSingleQuote
	case '`':
		return lexBackQuote
	default:
		if strings.ContainsRune(numberChars, firstChar) {
			return lexNumber
		}
		if strings.ContainsRune(identifierChar1, firstChar) {
			return lexIdentifier
		}
		return l.errorf("Unexpected non-identifier in lexNormal")
	}
	return lexNormal // though we could use a loop, here
}

/*
func lexFunction(l *Lexer) stateFn {
	l.pos += len(textFunction)
	l.emit(itemFunction)
	return lexInsideParen
}
*/

func (l *Lexer) next() (rune rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return EOF
	}
	rune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return rune
}

// ignore skips over the pending input before this point.
func (l *Lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune
// Can be called only once per call of next.
func (l *Lexer) backup() {
	l.pos -= l.width
}

func (l *Lexer) peek() rune {
	rune := l.next()
	l.backup()
	return rune
}

func (l *Lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *Lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

func lexIdentifier(l *Lexer) stateFn {
	l.acceptRun(identifierCharN)
	val := l.input[l.start:l.pos]
	if val == "if" {
		l.emit(ItemIf)
	} else {
		l.emit(ItemIdentifier)
	}
	return lexNormal
}

func lexSingleQuote(l *Lexer) stateFn {
	inEscape := false
	for {
		rune := l.next()
		if rune == EOF {
			return l.errorf("Unterminated single quote!?")
		}
		if inEscape {
			if !strings.ContainsRune("bnrt\\", rune) {
				return l.errorf("Bad single quote escape char!?")
			}
			inEscape = false
			continue
		}
		if rune == '\'' {
			l.emit(ItemSingleQuote)
			return lexNormal
		}
		if rune == '\\' {
			inEscape = true
		}
	}
	// not reached
}

func lexDoubleQuote(l *Lexer) stateFn {
	inEscape := false
	for {
		rune := l.next()
		if rune == EOF {
			return l.errorf("Unterminated double quote!?")
		}
		if inEscape {
			if !strings.ContainsRune("bnrt\"\\", rune) {
				return l.errorf("Bad double quote escape char!?")
			}
			inEscape = false
			continue
		}
		if rune == '"' {
			l.emit(ItemDoubleQuote)
			return lexNormal
		}
		if rune == '\\' {
			inEscape = true
		}
	}
	// not reached
}

func lexBackQuote(l *Lexer) stateFn {
	inEscape := false
	for {
		rune := l.next()
		if rune == EOF {
			return l.errorf("Unterminated back quote!?")
		}
		if inEscape {
			if !strings.ContainsRune("bnrt\\", rune) {
				return l.errorf("Bad back quote escape char!?")
			}
			inEscape = false
			continue
		}
		if rune == '`' {
			l.emit(ItemBackQuote)
			return lexNormal
		}
		if rune == '\\' {
			inEscape = true
		}
	}
	// not reached
}

func lexComment(l *Lexer) stateFn {
	for {
		rune := l.next()
		if rune == EOF {
			return l.errorf("Unterminated comment!?")
		}
		if rune == '\n' {
			l.emit(ItemComment)
			return lexNormal
		}
	}
	// not reached
}

func lexNumber(l *Lexer) stateFn {
	// Optional leading sign.
	l.accept("+-")
	// Is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	l.accept("b")
	l.emit(ItemNumber)
	return lexNormal

}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Token{
		ItemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}
