package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type item struct {
	Typ itemType // Type, such as itemNumber
	Val string   // Value, such as "23.2"
}

type itemType int

type stateFn func(*Lexer) stateFn

const (
	ItemError itemType = iota
	ItemEOF
	ItemText
	ItemNumber
	ItemLeftParen
	ItemRightParen
	ItemIdentifier
)

type Lexer struct {
	name  string // used only for error reports
	input string // the string being parsed
	start int
	pos   int
	width int
	items chan item
}

func (i item) String() string {
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

func Lex(name, input string) (*Lexer, chan item) {
	l := &Lexer{
		name:  name,
		input: input,
		items: make(chan item),
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

func (l *Lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

const EOF = -1

const identifierChar1 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
const identifierCharN = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func lexNormal(l *Lexer) stateFn {
	// fmt.Printf("lexNormal start: start=%d pos=%d width=%d\n", l.start, l.pos, l.width)
	for {
		// fmt.Printf("lexNormal loop: start=%d pos=%d width=%d\n", l.start, l.pos, l.width)
		if l.accept(identifierChar1) {
			return lexIdentifier
		}
		if l.accept(" ") {
			l.ignore()
			return lexNormal
		}
		if l.accept("(") {
			if l.pos > l.start {
				l.emit(ItemLeftParen)
			}
			return lexNormal // next state
		}
		if l.accept(")") {
			if l.pos > l.start {
				l.emit(ItemRightParen)
			}
			return lexNormal // next state
		}
		if l.next() == EOF {
			break
		}
	}
	// Correctly reached EOF
	l.emit(ItemEOF)
	return nil
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
		fmt.Printf("next() returning EOF!\n")
		return EOF
	}
	rune, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	// fmt.Printf("next() returning %v pos is %d width is %d\n", rune, l.pos, l.width)
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
	l.emit(ItemIdentifier)
	return lexNormal
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
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	if unicode.IsLetter(l.peek()) {
		l.next()
		return l.errorf("bad number syntax: %q",
			l.input[l.start:l.pos])
	}
	l.emit(ItemNumber)
	return lexNormal

}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		ItemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}
