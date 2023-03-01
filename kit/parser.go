package kit

import "github.com/shivamMg/rd"

func A(b *rd.Builder) (ok bool) {
	b.Enter("A")
	defer b.Exit(&ok)

	return b.Match("a") && B(b)
}

func B(b *rd.Builder) (ok bool) {
	defer b.Enter("B").Exit(&ok)

	return b.Match("b")
}
