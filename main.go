package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/vizicist/gk/lexer"
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
		fmt.Print(err)
	}
	s := string(b)
	gklex, itemChan := lexer.Lex("keykit", s)
	for {
		item := <-itemChan
		fmt.Printf("item=%v\n", item)
		if item == itemEOF {
			break
		}
	}
	fmt.Printf("final lex is %v\n", *gklex)
}

func usage() string {
	return "usage: gk {file}"
}
