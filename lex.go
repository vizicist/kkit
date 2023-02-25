package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode/utf8"
)

type item struct {
	typ itemType // Type, such as itemNumber
	val string   // Value, such as "23.2"
}

type itemType int

type stateFn func(*lexer) stateFn

const (
	itemError itemType = iota
	itemDot
	itemEOF
	itemElse
	itemIf
	itemFunction
	itemText
	itemLeftParam
	itemRightParam
)

type lexer struct {
	name  string // used only for error reports
	input string // the string being parsed
	start int
	pos   int
	width int
	items chan item
}

func main() {

	flag.Parse()

	fmt.Printf("args = %v\n", flag.Args())

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal(usage())
	}

	file, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := os.ReadFile("file.txt") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	s := string(b)
	gklex, itemChan := lex("keykit", s)
	_ = itemChan
	gklex.run()
}

func usage() string {
	return "usage: gk {file}"
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

func lex(name, input string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l, l.items
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

var textFunction = "function"

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], textFunction) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexFunction // next state
		}
		if l.next() == eof {
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

func lexFunction(l *lexer) stateFn {
	l.pos += len(textFunction)
	l.emit(itemFunction)
	return lexInsideParen
}

var rightParen = ")"
var leftParen = ")"

func lexInsideParen(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], rightParen) {
			return lexRightParen
		}
	}
	switch r := l.next(); {
	case r == eof || r == '\n':
		return l.errorf("unclosed paren")
	case isSpace(r):
		l.ignore()
	default:
		l.backup()
		return lexIdentifier
	}
}

func (l *lexer) next() (rune int) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	run, l.width =
		utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return rune
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// backup steps back one rune
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}
