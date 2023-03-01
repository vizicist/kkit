package kit

import (
	"fmt"
)

func ParseString(prog string) (parseTree *Tree, debugTree *DebugTree, err error) {

	_, tokenChan := Lex("keykit", prog)

	tokens := []Token{}
	for {
		token := <-tokenChan
		fmt.Printf("%s", token.Val)
		if token.Typ == ItemEOF {
			break
		}
		tokens = append(tokens, token)
	}

	b := NewBuilder(tokens)
	if ok := Program(b); ok && b.Err() == nil {
		return b.ParseTree(), b.DebugTree(), nil
	}
	return nil, b.DebugTree(), b.Err()
}

// func Parse(tokens []Token) (parseTree *Tree, debugTree *DebugTree, err error) {

func Program(b *Builder) (ok bool) {
	defer b.Enter("Program").Exit(&ok)

	if Term(b) && b.Match(ItemPlus) && Program(b) {
		return true
	}
	b.Backtrack()
	if Term(b) && b.Match(ItemMinus) && Program(b) {
		return true
	}
	b.Backtrack()
	return Term(b)
}

func Term(b *Builder) (ok bool) {
	defer b.Enter("Term").Exit(&ok)

	if Factor(b) && b.Match(ItemAsterisk) && Term(b) {
		return true
	}
	b.Backtrack()
	if Factor(b) && b.Match(ItemForwardSlash) && Term(b) {
		return true
	}
	b.Backtrack()
	return Factor(b)
}

func Factor(b *Builder) (ok bool) {
	defer b.Enter("Factor").Exit(&ok)

	if b.Match(ItemLeftParen) && Program(b) && b.Match(ItemRightParen) {
		return true
	}
	b.Backtrack()
	if b.Match(ItemMinus) && Factor(b) {
		return true
	}
	b.Backtrack()
	return Number(b)
}

func Number(b *Builder) (ok bool) {
	defer b.Enter("Number").Exit(&ok)

	token, ok := b.Next()
	if !ok {
		return false
	}
	if b.Match(ItemNumber) {
		b.Add(token)
		return true
	}
	b.Backtrack()
	return false
}
