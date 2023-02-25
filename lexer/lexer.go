package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type item struct {
	typ itemType // Type, such as itemNumber
	val string   // Value, such as "23.2"
}

type itemType int

type stateFn func(*Lexer) stateFn

const (
	itemError itemType = iota
	itemDot
	itemEOF
	itemElse
	itemIf
	itemFunction
	itemText
	itemNumber
	itemLeftParen
	itemRightParen
	itemIdentifier
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
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
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
	fmt.Printf("lexNormal start: l = %v\n", l)
	for {
		fmt.Printf("lexNormal : l = %v\n", l)
		if l.accept(identifierChar1) {
			return lexIdentifier
		}
		if l.accept(" ") {
			l.ignore()
			return lexNormal
		}
		if l.accept("(") {
			if l.pos > l.start {
				l.emit(itemLeftParen)
			}
			return lexNormal // next state
		}
		if l.accept(")") {
			if l.pos > l.start {
				l.emit(itemRightParen)
			}
			return lexNormal // next state
		}
		if l.next() == EOF {
			break
		}
	}
	// Correctly reached EOF
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
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
	// fmt.Printf("next() returning %v, pos is now %d\n", rune, l.pos, l.width)
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
	l.emit(itemIdentifier)
	return lexIdentifier
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
	l.emit(itemNumber)
	return lexNormal

}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}
