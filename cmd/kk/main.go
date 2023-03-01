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

	parseTree, debugTree, err := kit.ParseString(s)
	if err != nil {
		fmt.Print("Debug Tree:\n\n", debugTree)
		fmt.Printf("Parsing failed. err=%s\n", err.Error())
	}
	fmt.Print("Parse Tree:\n\n", parseTree)
}

func usage() string {
	return "usage: gk {file}"
}
