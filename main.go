package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"unicode"
)

const WordAlignment = 64

var wordlist = getWordList()
var prologue, epilogue, indexpage []byte

func getWordList() []string {
	content, err := os.ReadFile(os.Args[1])
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

func main() {
	file, err := os.ReadFile("template.html")
	if err != nil {
		panic(err)
	}
	prologue, epilogue, _ = bytes.Cut(file, []byte("%%%"))

	indexpage, err = os.ReadFile("index.html")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/search", handleSearch)
	// http.header.Add("Access-Control-Allow-Origin", "go.googlesource.com")
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	http.Handle("/static/fonts/", http.StripPrefix("/static/fonts", http.FileServer(http.Dir("./go-image/font/gofont/ttfs"))))

	log.Println("Server luistert op http://localhost:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

func cleanString(input string) string {
	return strings.Map(func(r rune) rune {
		r = unicode.ToLower(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("%?&*", r) {
			return r
		}
		return -1
	}, input)
}

func writeEpilog(w io.Writer, letters, pattern string) {
	epi := string(epilogue)
	epi = strings.Replace(epi, "%letters%", letters, 1)
	epi = strings.Replace(epi, "%pattern%", pattern, 1)
	w.Write([]byte(epi))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write(prologue)
	w.Write(indexpage)
	writeEpilog(w, "", "")
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	letters := r.URL.Query().Get("letters")
	pattern := r.URL.Query().Get("pattern")
	letters = cleanString(letters)
	pattern = cleanString(pattern)

	if pattern == "" {
		pattern = "%"
	}
	scores := FindWords(wordlist, letters, pattern)

	w.Write(prologue)

	fmt.Fprintf(w, "<h1>Wordfeud Helper</h1>\n")
	fmt.Fprintf(w, "<h2>%d woorden gevonden</h2>\n", len(scores))
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "<p><strong>Letters:</strong> <code>%s</code></p>\n", letters)
	fmt.Fprintf(w, "<p><strong>Patroon:</strong> <code>%s</code></p>\n", pattern)

	if len(scores) > 0 {
		overflow := 0
		if len(scores) > 100 {
			overflow = len(scores) - 100
			scores = scores[:100]
		}

		curlen := 0
		for _, word := range scores {
			if len(word.Word) != curlen {
				curlen = len(word.Word)
				fmt.Fprintf(w, "\n<h3>%d Letters</h3>\n", curlen)
			} else {
				fmt.Fprintf(w, ", ")
			}
			for _, chr := range word.Word {
				l := byte(chr) - 'a'
				if word.Consumed[l] > 0 {
					fmt.Fprintf(w, "<span style='color: red;'>%c</span>", chr)
					word.Consumed[l]--
				} else {
					fmt.Fprintf(w, "%c", chr)
				}
			}
			fmt.Fprintf(w, "<sub>%d</sub>", word.Score)
		}
		fmt.Fprintln(w)
		if overflow > 0 {
			fmt.Fprintf(w, "<p><strong>en %d meer...</strong></p>\n", overflow)
		}
	} else {
		fmt.Fprintf(w, "<p><em>Geen woorden gevonden</em></p>\n")
	}
	fmt.Fprintf(w, "<p><a href='.'>&ldsh; Ga terug</a></p>\n")
	fmt.Fprintf(w, "<h2>Zoek opnieuw</h2>\n")
	writeEpilog(w, letters, pattern)
}
