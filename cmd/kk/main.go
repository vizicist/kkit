package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vizicist/kkit/kit"
)

func main() {

	flag.Parse()

	fmt.Printf("args = %v\n", flag.Args())

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal(usage())
	}

	b, err := os.ReadFile(args[0])
	if err != nil {
		log.Fatal(err.Error())
	}
	s := string(b)
	_, itemChan := kit.Lex("keykit", s)

	tokeArr := []kit.Token{}
	for {
		item := <-itemChan
		fmt.Printf("%s", item.Val)
		if item.Typ == kit.ItemEOF {
			break
		}
		tokeArr = append(tokeArr, Token{Typ})
	}

	tokens := []kit.Token{"a", "b"}
	b := kit.NewBuilder(tokens)
	if ok := A(b); ok {
		fmt.Print(b.ParseTree())
	}
}

func usage() string {
	return "usage: gk {file}"
}
