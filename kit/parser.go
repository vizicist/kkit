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
	ok := Program(b)
	if ok && b.Err() == nil {
		return b.ParseTree(), b.DebugTree(), nil
	}
	return nil, b.DebugTree(), b.Err()
}

// func Parse(tokens []Token) (parseTree *Tree, debugTree *DebugTree, err error) {

func Program(b *Builder) (ok bool) {
	defer b.Enter("Program").Exit(&ok)

	// A Program is a series of Stmts
	for Stmt(b) {
		fmt.Printf("Stmt okay, going for another\n")
	}
	return true
}

func Stmts(b *Builder) (ok bool) {
	defer b.Enter("Stmts").Exit(&ok)

	for Stmt(b) {
		fmt.Printf("Stmt okay in Stmts, going for another\n")
	}
	return true
}

func OneStatement(b *Builder) (ok bool) {
	defer b.Enter("OneStatement").Exit(&ok)

	// curly-brace-enclosed statements
	if Expect(b, ItemLeftBrace) && Stmts(b) && Expect(b, ItemRightBrace) {
		return true
	}
	b.Backtrack()
	return Stmt(b)
}

func Stmt(b *Builder) (ok bool) {
	defer b.Enter("Stmt").Exit(&ok)

	// A function call
	if b.Match(ItemIdentifier) && Expect(b, ItemLeftParen) && Expr(b) && Expect(b, ItemRightParen) {
		return true
	}
	b.Backtrack()
	if b.Match(ItemComment) {
		return true
	}
	b.Backtrack()
	if b.Match(ItemIf) && Expect(b, ItemLeftParen) && Expr(b) && Expect(b, ItemRightParen) && OneStatement(b) {
		return true
	}
	b.Backtrack()
	return false
}

func Expect(b *Builder, tokenType tokenType) (ok bool) {
	next, ok := b.Peek(1)
	if ok && next.Typ == tokenType {
		b.Next()
		return true
	}
	return false
}

func LeftParen(b *Builder) (ok bool) {
	next, ok := b.Peek(1)
	if ok && next.Typ == ItemLeftParen {
		b.Next()
		return true
	}
	return false

}

func Expr(b *Builder) (ok bool) {
	defer b.Enter("Expr").Exit(&ok)

	if b.Match(ItemLeftParen) && Expr(b) && b.Match(ItemRightParen) {
		return true
	}
	b.Backtrack()

	if b.Match(ItemNumber) && ConditionOp(b) && Expr(b) {
		return true
	}
	b.Backtrack()

	if b.Match(ItemIdentifier) && ConditionOp(b) && Expr(b) {
		return true
	}
	b.Backtrack()

	if b.Match(ItemNumber) {
		return true
	}
	b.Backtrack()

	if b.Match(ItemIdentifier) {
		return true
	}
	b.Backtrack()

	return false
}

func ConditionOp(b *Builder) (ok bool) {
	defer b.Enter("ConditionOp").Exit(&ok)

	if b.Match(ItemEqualsEquals) {
		return true
	}
	b.Backtrack()

	if b.Match(ItemNotEquals) {
		return true
	}
	b.Backtrack()

	if b.Match(ItemLessThan) {
		return true
	}
	b.Backtrack()

	if b.Match(ItemLessThanOrEqual) {
		return true
	}
	b.Backtrack()

	if b.Match(ItemGreaterThan) {
		return true
	}
	b.Backtrack()

	if b.Match(ItemGreaterThanOrEqual) {
		return true
	}
	b.Backtrack()
	return false
}

/*
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

*/
