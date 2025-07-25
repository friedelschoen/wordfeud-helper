package main

import (
	"bytes"
	"flag"
	"os"
	"strings"
)

func getWordList(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	list := make([]string, 0, strings.Count(string(content), "\n"))
	for line := range strings.SplitSeq(string(content), "\n") {
		if len(line) > 0 {
			list = append(list, line)
		}
	}
	return list
}

var (
	wordlistFlag = flag.String("wordlist", "wordlist.txt", "specifies path to wordlist")
	workdirFlag  = flag.String("workdir", ".", "specifies path to index.html, template.html...")
	addrFlag     = flag.String("address", "0.0.0.0:8080", "specifies address to bind to")
)

func main() {
	flag.Parse()

	wordlist = getWordList(*workdirFlag + "/" + *wordlistFlag)
	file, err := os.ReadFile(*workdirFlag + "/template.html")
	if err != nil {
		panic(err)
	}
	prologue, epilogue, _ = bytes.Cut(file, []byte("%%%"))

	indexpage, err = os.ReadFile(*workdirFlag + "/index.html")
	if err != nil {
		panic(err)
	}

	addr = "0.0.0.0:8080"

	serve()
}
