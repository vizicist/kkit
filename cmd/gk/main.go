package main

import (
	"fmt"
	"os"

	"io/ioutil"

	"github.com/vizicist/geekit/kit"
)

func main() {
	if len(os.Args) != 2 {
		printExit("invalid arguments. pass program file as an argument")
	}
	code, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		printExit("could not open file", os.Args[1], "err:", err)
	}

	_, itemChan := kit.Lex("gk", string(code))

	for {
		item := <-itemChan
		if item.Typ == kit.ItemEOF {
			break
		}
		fmt.Printf("%s", item.Val)
	}

	fmt.Println("\nGrammar:", kit.Grammar)

	parseTree, debugTree, err := kit.Parse(tokens)
	if err != nil {
		fmt.Print("Debug Tree:\n\n", debugTree)
		printExit("parsing failed.", err)
	}

	fmt.Print("Parse Tree:\n\n", parseTree)
}

func printExit(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}
